package exercisecatalog

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// PostgreSQL deparses the supported IN check as an ANY array. Accept only
// literal text elements, so arbitrary or weakened expressions fail closed.
var membershipCheck = regexp.MustCompile(`^CHECK \(\(catalog_id = ANY \(ARRAY\[(.*)\]\)\)\)$`)
var membershipLiteral = regexp.MustCompile(`^'((?:[^']|'')*)'::text(?:, |$)`)

func integrationURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") ||
		(u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") ||
		(u.Port() != "55432" && u.Port() != "26432") || u.Path == "" || u.Path == "/" || u.Fragment != "" {
		return fmt.Errorf("DATABASE_URL must explicitly select an isolated local test database on port 55432 or 26432")
	}
	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return fmt.Errorf("DATABASE_URL contains invalid connection parameters")
	}
	for key := range params {
		if key != "sslmode" && key != "connect_timeout" {
			return fmt.Errorf("DATABASE_URL contains unsupported connection parameters")
		}
	}
	return nil
}

func checkMembership(ctx context.Context, tx pgx.Tx, table string, expected map[string]bool) error {
	var nullable bool
	if err := tx.QueryRow(ctx, `SELECT NOT attnotnull FROM pg_attribute WHERE attrelid=$1::regclass AND attname='catalog_id' AND NOT attisdropped`, table).Scan(&nullable); err != nil {
		return err
	}
	if !nullable {
		return fmt.Errorf("%s catalog_id must allow NULL", table)
	}
	rows, err := tx.Query(ctx, `SELECT pg_get_constraintdef(c.oid), c.convalidated FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=ANY(c.conkey) WHERE c.conrelid=$1::regclass AND c.contype='c' AND a.attname='catalog_id'`, table)
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var definition string
		var validated bool
		if err := rows.Scan(&definition, &validated); err != nil {
			return err
		}
		count++
		match := membershipCheck.FindStringSubmatch(definition)
		if !validated || match == nil {
			return fmt.Errorf("%s catalog check has unsupported or unvalidated definition: %s", table, definition)
		}
		actual := make(map[string]bool)
		for rest := match[1]; rest != ""; {
			part := membershipLiteral.FindStringSubmatch(rest)
			if part == nil {
				return fmt.Errorf("%s catalog check has unsupported element: %s", table, rest)
			}
			actual[strings.ReplaceAll(part[1], "''", "'")] = true
			rest = rest[len(part[0]):]
		}
		if len(actual) != len(expected) {
			return fmt.Errorf("%s catalog membership differs from embedded JSON", table)
		}
		for id := range expected {
			if !actual[id] {
				return fmt.Errorf("%s catalog membership missing %q", table, id)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("%s has %d catalog checks, expected one", table, count)
	}
	return nil
}

func TestCatalogIntegrationURLSafety(t *testing.T) {
	for _, raw := range []string{"", "postgres://postgres:password@localhost:5432/fittrack_test", "postgres://example.com:55432/test", "postgres://localhost:55432/test?host=example.com"} {
		require.Error(t, integrationURL(raw))
	}
	require.NoError(t, integrationURL("postgres://localhost:55432/isolated_test?sslmode=disable"))
}

func TestCatalogStorageMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	require.NoError(t, integrationURL(url))
	conn, err := pgx.Connect(ctx, url)
	require.NoError(t, err)
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Quote the IDs as SQL literals; the values originate in the embedded JSON.
	quoted := make([]string, 0, len(All()))
	expected := make(map[string]bool, len(All()))
	for _, entry := range All() {
		quoted = append(quoted, "'"+strings.ReplaceAll(entry.ID, "'", "''")+"'")
		expected[entry.ID] = true
	}

	// Load the SQLC model in a transaction-local schema; nothing persists.
	_, err = tx.Exec(ctx, `CREATE SCHEMA catalog_parity_test; SET LOCAL search_path TO catalog_parity_test`)
	require.NoError(t, err)
	schema, err := os.ReadFile("../../schema.sql")
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(schema))
	require.NoError(t, err)
	for _, table := range []string{"public.exercise", "catalog_parity_test.exercise"} {
		require.NoError(t, checkMembership(ctx, tx, table, expected), table)
	}

	// Equivalent membership in another order must not be classified as drift.
	reversed := append([]string(nil), quoted...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE reordered (catalog_id text CHECK (catalog_id IN ("+strings.Join(reversed, ",")+")))")
	require.NoError(t, err)
	require.NoError(t, checkMembership(ctx, tx, "pg_temp.reordered", expected))

	// Mutate only temporary tables, never the migrated or schema tables.
	for _, tc := range []struct{ name, check string }{
		{"missing ID", strings.Join(quoted[1:], ",")},
		{"extra ID", strings.Join(append(append([]string{}, quoted...), "'unknown_catalog_id'"), ",")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tx.Exec(ctx, "CREATE TEMP TABLE drift (catalog_id text CHECK (catalog_id IN ("+tc.check+")))")
			require.NoError(t, err)
			require.ErrorContains(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "catalog membership differs")
			_, err = tx.Exec(ctx, "DROP TABLE pg_temp.drift")
			require.NoError(t, err)
		})
	}
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE drift (catalog_id text)`)
	require.NoError(t, err)
	require.ErrorContains(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "has 0 catalog checks")
	_, err = tx.Exec(ctx, `ALTER TABLE pg_temp.drift ADD CONSTRAINT weakened CHECK (catalog_id IS NULL OR catalog_id IS NOT NULL)`)
	require.NoError(t, err)
	require.ErrorContains(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "unsupported or unvalidated")
	_, err = tx.Exec(ctx, `DROP TABLE pg_temp.drift`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE drift (catalog_id text NOT NULL CHECK (catalog_id IN ("+strings.Join(quoted, ",")+")))")
	require.NoError(t, err)
	require.ErrorContains(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "must allow NULL")
}

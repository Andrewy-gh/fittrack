package exercisecatalog

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// checkMembership compares PostgreSQL's normalized constraint against a check
// built from the embedded catalog. It deliberately fails closed on other check
// shapes rather than pretending to prove arbitrary SQL expressions equivalent.
func checkMembership(ctx context.Context, tx pgx.Tx, table string, expected string) error {
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
		if !validated || definition != expected {
			return fmt.Errorf("%s catalog check differs from embedded JSON: %s (validated=%t)", table, definition, validated)
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

func TestCatalogStorageMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
	}
	conn, err := pgx.Connect(ctx, url)
	require.NoError(t, err)
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Quote the IDs as SQL literals; the values originate in the embedded JSON.
	quoted := make([]string, 0, len(All()))
	for _, entry := range All() {
		quoted = append(quoted, "'"+strings.ReplaceAll(entry.ID, "'", "''")+"'")
	}
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE catalog_expected (catalog_id text CHECK (catalog_id IN ("+strings.Join(quoted, ",")+")))")
	require.NoError(t, err)
	var expected string
	require.NoError(t, tx.QueryRow(ctx, `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='pg_temp.catalog_expected'::regclass AND contype='c'`).Scan(&expected))

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

	// Mutate only a temporary clone, never the migrated or schema tables.
	for _, tc := range []struct{ name, check string }{
		{"missing ID", strings.Join(quoted[1:], ",")},
		{"extra ID", strings.Join(append(append([]string{}, quoted...), "'unknown_catalog_id'"), ",")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tx.Exec(ctx, "CREATE TEMP TABLE drift (catalog_id text CHECK (catalog_id IN ("+tc.check+")))")
			require.NoError(t, err)
			require.Error(t, checkMembership(ctx, tx, "pg_temp.drift", expected))
			_, err = tx.Exec(ctx, "DROP TABLE pg_temp.drift")
			require.NoError(t, err)
		})
	}
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE drift (catalog_id text)`)
	require.NoError(t, err)
	require.Error(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "removed constraint must be detected")
	_, err = tx.Exec(ctx, `ALTER TABLE pg_temp.drift ADD CONSTRAINT weakened CHECK (catalog_id IS NULL OR catalog_id IS NOT NULL)`)
	require.NoError(t, err)
	require.Error(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "weakened constraint must be detected")
	_, err = tx.Exec(ctx, `DROP TABLE pg_temp.drift`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE drift (catalog_id text NOT NULL CHECK (catalog_id IN ("+strings.Join(quoted, ",")+")))")
	require.NoError(t, err)
	require.Error(t, checkMembership(ctx, tx, "pg_temp.drift", expected), "NULL must remain allowed")
}

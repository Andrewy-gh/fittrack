package exercisecatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// PostgreSQL deparses IN as equality for one ID and ANY for multiple IDs.
// Accept only literal text elements, so weakened expressions fail closed.
var membershipCheck = regexp.MustCompile(`^CHECK \(\(catalog_id = ANY \(ARRAY\[(.*)\]\)\)\)$`)
var singletonCheck = regexp.MustCompile(`^CHECK \(\(catalog_id = ('(?:[^']|'')*'::text)\)\)$`)
var membershipLiteral = regexp.MustCompile(`^'((?:[^']|'')*)'::text(?:, |$)`)

func integrationURL(raw, worktreeID string, isolatedCI bool) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") ||
		(u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") || u.Fragment != "" {
		return fmt.Errorf("DATABASE_URL must explicitly select an isolated local test database")
	}
	local := regexp.MustCompile(`^[a-z0-9_]+_[a-f0-9]{12}$`).MatchString(worktreeID) &&
		(u.Port() == "55432" || u.Port() == "26432") && u.Path == "/ft_"+worktreeID+"_test"
	ci := isolatedCI && u.Port() == "5432" && u.Path == "/fittrack_test" && u.User != nil && u.User.Username() == "test_user"
	if !local && !ci {
		return fmt.Errorf("DATABASE_URL must select this checkout's managed test database or the explicit isolated CI service")
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
		if match == nil {
			match = singletonCheck.FindStringSubmatch(definition)
		}
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
	id := "example_0123456789ab"
	for _, raw := range []string{"", "postgres://test_user:password@localhost:5432/fittrack_test", "postgres://localhost:55432/other", "postgres://example.com:55432/ft_" + id + "_test", "postgres://localhost:55432/ft_" + id + "_test?host=example.com"} {
		require.Error(t, integrationURL(raw, id, false))
	}
	require.NoError(t, integrationURL("postgres://localhost:55432/ft_"+id+"_test?sslmode=disable", id, false))
	require.NoError(t, integrationURL("postgres://test_user@localhost:5432/fittrack_test", "", true))
	require.Error(t, integrationURL("postgres://test_user@localhost:5432/other", "", true))
}

func TestCatalogStorageMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	worktreeID := os.Getenv("FITTRACK_WORKTREE_ID")
	isolatedCI := os.Getenv("GITHUB_ACTIONS") == "true" && os.Getenv("FITTRACK_CATALOG_CI_ISOLATED") == "1"
	require.NoError(t, integrationURL(url, worktreeID, isolatedCI))
	conn, err := pgx.Connect(ctx, url)
	require.NoError(t, err)
	defer conn.Close(ctx)
	if worktreeID != "" {
		var marker, owner, current string
		require.NoError(t, conn.QueryRow(ctx, `SELECT coalesce(shobj_description(oid, 'pg_database'), ''), pg_get_userbyid(datdba), current_user FROM pg_database WHERE datname=current_database()`).Scan(&marker, &owner, &current))
		require.Equal(t, "fittrack-worktree:"+worktreeID, marker, "refusing database without this checkout's ownership marker")
		require.Equal(t, current, owner, "refusing database not owned by the connected role")
	}
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

	// Exercise the real generator with an expanded catalog in an isolated copy.
	dir := t.TempDir()
	for _, name := range []string{"scripts", "server/internal/exercisecatalog"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, name), 0o755))
	}
	script, err := os.ReadFile("../../../scripts/catalog-sql.mjs")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts/catalog-sql.mjs"), script, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "server/schema.sql"), schema, 0o600))
	syntheticID := "Synthetic_'Expansion"
	expanded := append(append([]Entry(nil), All()...), Entry{ID: syntheticID})
	contents, err := json.Marshal(expanded)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "server/internal/exercisecatalog/catalog.json"), contents, 0o600))
	runGenerator := func(arg string) ([]byte, error) {
		return exec.Command("node", filepath.Join(dir, "scripts/catalog-sql.mjs"), arg).CombinedOutput()
	}
	_, err = runGenerator("--check")
	require.Error(t, err, "unregenerated schema must fail")
	output, err := runGenerator("--write")
	require.NoError(t, err, string(output))
	output, err = runGenerator("--check")
	require.NoError(t, err, string(output))
	snippet, err := runGenerator("--snippet")
	require.NoError(t, err, string(snippet))
	expandedSchema, err := os.ReadFile(filepath.Join(dir, "server/schema.sql"))
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `CREATE SCHEMA catalog_expanded_test; SET LOCAL search_path TO catalog_expanded_test`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(expandedSchema))
	require.NoError(t, err)
	expandedExpected := make(map[string]bool, len(expected)+1)
	for id := range expected {
		expandedExpected[id] = true
	}
	expandedExpected[syntheticID] = true
	require.NoError(t, checkMembership(ctx, tx, "catalog_expanded_test.exercise", expandedExpected))
	// The emitted forward snapshot applies to a temporary current-constraint table.
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE exercise (catalog_id text CHECK (catalog_id IN ("+strings.Join(quoted, ",")+")))")
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(snippet))
	require.NoError(t, err)
	require.NoError(t, checkMembership(ctx, tx, "pg_temp.exercise", expandedExpected))
	_, err = tx.Exec(ctx, `INSERT INTO pg_temp.exercise (catalog_id) VALUES ($1), (NULL)`, syntheticID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SAVEPOINT invalid_catalog`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO pg_temp.exercise (catalog_id) VALUES ('not_in_catalog')`)
	require.ErrorContains(t, err, "exercise_catalog_id_check")
	_, err = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT invalid_catalog`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `DROP TABLE pg_temp.exercise`)
	require.NoError(t, err)

	// PostgreSQL rewrites a single-element IN to equality, not ANY(array).
	contents, err = json.Marshal([]Entry{{ID: syntheticID}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "server/internal/exercisecatalog/catalog.json"), contents, 0o600))
	output, err = runGenerator("--write")
	require.NoError(t, err, string(output))
	output, err = runGenerator("--check")
	require.NoError(t, err, string(output))
	singleSchema, err := os.ReadFile(filepath.Join(dir, "server/schema.sql"))
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `CREATE SCHEMA catalog_single_test; SET LOCAL search_path TO catalog_single_test`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(singleSchema))
	require.NoError(t, err)
	require.NoError(t, checkMembership(ctx, tx, "catalog_single_test.exercise", map[string]bool{syntheticID: true}))
	snippet, err = runGenerator("--snippet")
	require.NoError(t, err, string(snippet))
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE exercise (catalog_id text CHECK (catalog_id IN ("+strings.Join(quoted, ",")+")))")
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(snippet))
	require.NoError(t, err)
	require.NoError(t, checkMembership(ctx, tx, "pg_temp.exercise", map[string]bool{syntheticID: true}))
	_, err = tx.Exec(ctx, `INSERT INTO pg_temp.exercise (catalog_id) VALUES ($1), (NULL)`, syntheticID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SAVEPOINT singleton_invalid`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO pg_temp.exercise (catalog_id) VALUES ('not_in_catalog')`)
	require.ErrorContains(t, err, "exercise_catalog_id_check")
	_, err = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT singleton_invalid`)
	require.NoError(t, err)

	// Equivalent membership in another order must not be classified as drift.
	reversed := append([]string(nil), quoted...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	_, err = tx.Exec(ctx, "CREATE TEMP TABLE reordered (catalog_id text CHECK (catalog_id IN ("+strings.Join(reversed, ",")+")))")
	require.NoError(t, err)
	require.NoError(t, checkMembership(ctx, tx, "pg_temp.reordered", expected))

	// Mutate only temporary tables, never the migrated or schema tables.
	// A singleton still needs valid SQL with a genuinely missing catalog member.
	unknown := "unknown_catalog_id"
	for expected[unknown] {
		unknown += "_"
	}
	missing := strings.Join(quoted[1:], ",")
	missingError := "catalog membership differs"
	if len(quoted) == 1 {
		missing = "'" + unknown + "'"
		missingError = "catalog membership missing"
	}
	for _, tc := range []struct{ name, check, diagnostic string }{
		{"missing ID", missing, missingError},
		{"extra ID", strings.Join(append(append([]string{}, quoted...), "'"+unknown+"'"), ","), "catalog membership differs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tx.Exec(ctx, "CREATE TEMP TABLE drift (catalog_id text CHECK (catalog_id IN ("+tc.check+")))")
			require.NoError(t, err)
			require.ErrorContains(t, checkMembership(ctx, tx, "pg_temp.drift", expected), tc.diagnostic)
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

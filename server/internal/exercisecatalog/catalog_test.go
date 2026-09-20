package exercisecatalog

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewedCatalog(t *testing.T) {
	require.Len(t, All(), 25)
	migration, err := os.ReadFile("../../migrations/00033_exercise_catalog.sql")
	require.NoError(t, err)
	// Git may check SQL files out with CRLF on Windows.
	migrationSQL := strings.ReplaceAll(string(migration), "\r\n", "\n")
	ids := map[string]bool{}
	for _, entry := range All() {
		require.NotEmpty(t, entry.ID)
		require.False(t, ids[entry.ID], "duplicate ID")
		ids[entry.ID] = true
		require.NotEmpty(t, entry.Name)
		require.NotEmpty(t, entry.Equipment)
		require.NotEmpty(t, entry.PrimaryMuscles)
		require.NotNil(t, entry.SecondaryMuscles)
		require.Contains(t, migrationSQL, "'"+entry.ID+"'")
		for _, muscle := range entry.PrimaryMuscles {
			require.NotContains(t, entry.SecondaryMuscles, muscle)
		}
		require.Equal(t, entry, *Find(entry.ID))
	}
	require.Equal(t, len(ids), strings.Count(migrationSQL, "'\n")+strings.Count(migrationSQL, "',\n"))
	require.Nil(t, Find("Bench Press"), "names must never infer classification")
	require.Nil(t, Find(""))
}

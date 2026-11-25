package exercisecatalog

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewedCatalog(t *testing.T) {
	ids := map[string]bool{}
	for _, entry := range All() {
		require.NotEmpty(t, entry.ID)
		require.False(t, ids[entry.ID], "duplicate ID")
		ids[entry.ID] = true
		require.NotEmpty(t, entry.Name)
		require.NotEmpty(t, entry.Equipment)
		require.NotEmpty(t, entry.PrimaryMuscles)
		require.NotNil(t, entry.SecondaryMuscles)
		for _, muscle := range entry.PrimaryMuscles {
			require.NotContains(t, entry.SecondaryMuscles, muscle)
		}
		require.Equal(t, entry, *Find(entry.ID))
	}
	require.Nil(t, Find("Bench Press"), "names must never infer classification")
	require.Nil(t, Find(""))
}

func TestCatalogVersionCanonicalizesLineEndingsAndChangesWithContent(t *testing.T) {
	lf := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	crlf := bytes.ReplaceAll(lf, []byte("\n"), []byte("\r\n"))

	lfVersion, err := catalogVersionForJSON(lf)
	require.NoError(t, err)
	crlfVersion, err := catalogVersionForJSON(crlf)
	require.NoError(t, err)
	require.Equal(t, lfVersion, crlfVersion)
	require.Equal(t, Version(), lfVersion)

	changed := append([]Entry(nil), All()...)
	changed[0].Name += " changed"
	require.NotEqual(t, Version(), catalogVersionForEntries(changed))
}

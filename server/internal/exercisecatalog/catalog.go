// Package exercisecatalog owns the reviewed, versioned exercise vocabulary.
package exercisecatalog

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
)

//go:embed catalog.json
var source []byte

type Entry struct {
	ID               string   `json:"id" validate:"required"`
	Name             string   `json:"name" validate:"required"`
	Equipment        string   `json:"equipment" validate:"required"`
	PrimaryMuscles   []string `json:"primaryMuscles" validate:"required"`
	SecondaryMuscles []string `json:"secondaryMuscles" validate:"required"`
}

var entries = mustParseEntries(source)

var catalogVersion = mustCatalogVersion(source)

func All() []Entry { return entries }

// Version returns the SHA-256 digest of the catalog's canonical content. Canonical content is
// json.Marshal of the decoded, ordered []Entry value, so whitespace and CRLF/LF differences in
// catalog.json do not affect the version while every entry field and list order does.
func Version() string { return catalogVersion }

func mustParseEntries(contents []byte) []Entry {
	entries, err := parseEntries(contents)
	if err != nil {
		panic(err)
	}
	return entries
}

func parseEntries(contents []byte) ([]Entry, error) {
	var result []Entry
	if err := json.Unmarshal(contents, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func mustCatalogVersion(contents []byte) string {
	version, err := catalogVersionForJSON(contents)
	if err != nil {
		panic(err)
	}
	return version
}

func catalogVersionForJSON(contents []byte) (string, error) {
	entries, err := parseEntries(contents)
	if err != nil {
		return "", err
	}
	return catalogVersionForEntries(entries), nil
}

func catalogVersionForEntries(entries []Entry) string {
	canonical, err := json.Marshal(entries)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:])
}

func Find(id string) *Entry {
	for _, entry := range entries {
		if entry.ID == id {
			return &entry
		}
	}
	return nil
}

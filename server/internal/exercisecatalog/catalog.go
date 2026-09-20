// Package exercisecatalog owns the reviewed, versioned exercise vocabulary.
package exercisecatalog

import (
	_ "embed"
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

var entries = func() []Entry {
	var result []Entry
	if err := json.Unmarshal(source, &result); err != nil {
		panic(err)
	}
	return result
}()

func All() []Entry { return entries }

func Find(id string) *Entry {
	for _, entry := range entries {
		if entry.ID == id {
			return &entry
		}
	}
	return nil
}

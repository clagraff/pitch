package entities

import (
	"github.com/clagraff/pitch/entities/items"
	"github.com/clagraff/pitch/entities/objects"
)

// World represents a container for the Object and Item collections.
type World struct {
	Objects objects.Collection `json:"objects"`
	Items   items.Collection   `json:"items"`
}

// MakeWorld will instantiate and return a new World struct.
func MakeWorld() World {
	w := World{
		Objects: objects.MakeCollection(),
		Items:   items.MakeCollection(),
	}

	return w
}

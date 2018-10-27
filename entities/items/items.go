package items

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/utils"
)

// Collection represents a collection of Items.
type Collection struct {
	mapping map[uuid.UUID]Item
}

// MarshalJSON marshals the current Collection as a list (as opposed to a map).
func (c Collection) MarshalJSON() ([]byte, error) {
	l := make([]Item, len(c.mapping))
	i := 0

	for _, item := range c.mapping {
		l[i] = item
		i++
	}

	return json.Marshal(l)
}

// UnmarshalJSON unmarshals JSON bytes into the current Collection, assuming the
// JSON represents a list (as opposed to a map).
func (c *Collection) UnmarshalJSON(data []byte) error {
	l := make([]Item, 0)
	err := json.Unmarshal(data, &l)
	if err != nil {
		return err
	}

	for _, item := range l {
		c.Insert(item)
	}

	return nil
}

// MakeCollection instantiates a new collection and returns the instance.
func MakeCollection() Collection {
	c := Collection{}
	c.mapping = make(map[uuid.UUID]Item)

	return c
}

// Insert will insert the provided item into the collection.
func (c *Collection) Insert(i Item) {
	id := i.ID

	if c.mapping == nil {
		c.mapping = make(map[uuid.UUID]Item)
	}

	c.mapping[id] = i
}

// FromID will attempt to return an Item from the current collection, as
// specified by the provided UUID.
func (c Collection) FromID(id uuid.UUID) (Item, bool) {
	i, ok := c.mapping[id]
	return i, ok
}

// FromIDString will attempt to return an Item from the current collection, as
// specified by the provided UUID string.
func (c Collection) FromIDString(id string) (Item, bool) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return Item{}, false
	}
	i, ok := c.mapping[uid]
	return i, ok
}

// Remove will the first item in the current collection which matches the
// specified UUID of the provided item.
func (c *Collection) Remove(i Item) bool {
	if _, ok := c.mapping[i.ID]; !ok {
		return false
	}

	delete(c.mapping, i.ID)
	return true
}

// DamageType represents what kind of damage the item does.
type DamageType int

// Various types of DamageType constants.
const (
	MeleeDamage DamageType = iota
	RangeDamage
)

// Damage is used to represent a possible damage amount and type.
// This includes a damage modifier (if applicable), and the die range
// and roll amount. These are used to calculate a one-time damage amount.
type Damage struct {
	DieRange   int        `json:"die_range"`
	Modifier   int        `json:"modifier"`
	RollAmount int        `json:"roll_amount"`
	Type       DamageType `json:"damage_type"`
}

// Calculate returns a randomized damage amount, based on the Damage's current
// modifier, die range, and roll amount.
func (d Damage) Calculate() int {
	roll := utils.Roll(d.RollAmount, d.DieRange)
	return roll + d.Modifier
}

// Armor is used to represent the possible damage reductions, categorized by
// damage types.
type Armor struct {
	MeleeReduction int `json:"melee_reduction"`
	RangeReduction int `json:"range_reduction"`
}

// Item is used to represent items with a UUID and Damage & Armor objects.
type Item struct {
	ID     uuid.UUID `json:"id"`
	Damage Damage    `json:"damage"`
	Armor  Armor     `json:"armor"`
}

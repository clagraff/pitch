package objects

import (
	"fmt"
	"math"
	"time"

	"github.com/clagraff/pitch/logging"
	uuid "github.com/satori/go.uuid"
)

// Collection is used to represent a grouping of Entity objects.
type Collection []Entity

// MakeCollection instantiates a new collection and returns the instance.
func MakeCollection() Collection {
	return make([]Entity, 0)
}

// NewCollection instantiates a new collection and returns its pointer.
func NewCollection() *Collection {
	c := MakeCollection()
	return &c
}

// Prepend will insert the provided entity into the first position.
func (c Collection) Prepend(e Entity) Collection {
	return append([]Entity{e}, c...)
}

// Insert will insert the provided entity into the collection.
func (c Collection) Append(e Entity) Collection {
	return append(c, e)
}

// FromID will attempt to return an Entity from the current collection,
// as specified by the provided UUID.
func (c Collection) FromID(id uuid.UUID) (Entity, bool) {
	for _, e := range c {
		if uuid.Equal(e.ID, id) {
			return e, true
		}
	}

	return Entity{}, false
}

// FromIDString will attempt to return an Entity from the current collection,
// as specified by the provided UUID string.
func (c Collection) FromIDString(id string) (Entity, bool) {
	logger, closeLog := logging.Logger("entities.objects.Collection.FromIDString")
	defer closeLog()

	logger.Println("Looking for ID:", id)

	uid, err := uuid.FromString(id)
	if err != nil {
		logger.Println("Failed to parse UUID string!")
		return Entity{}, false
	}

	foundEntity, wasFound := c.FromID(uid)
	if wasFound {
		logger.Println("Found entity", foundEntity.ID.String())
	} else {
		logger.Println("Could not find entity with ID:", id)
	}

	return foundEntity, wasFound
}

// FromXY will return a map of all entities which exist in the current
// collection at the given XY position.
func (c Collection) FromXY(x, y int) []Entity {
	foundEntities := make([]Entity, 0)

	for _, e := range c {
		if e.Position.X == x && e.Position.Y == y {
			foundEntities = append(foundEntities, e)
		}
	}

	return foundEntities
}

// Remove will remove the first entity in the current collection which
// matches the specified UUID of the provided entity.
func (c Collection) Remove(e Entity) (Collection, bool) {
	newCollection := make(Collection, 0)
	found := false

	for _, currentEntity := range c {
		if !uuid.Equal(currentEntity.ID, e.ID) {
			newCollection = append(newCollection, currentEntity)
		} else {
			found = true
		}
	}

	return newCollection, found
}

// MustRemove will remove return a copy of the current collection minus any
// entities matching the target entity's ID.
// If no entities were removed, this call panics.
func (c Collection) MustRemove(target Entity) Collection {
	col, ok := c.Remove(target)
	if !ok {
		panic(fmt.Sprintf("could not remove target: %s", target))
	}

	return col
}

// Replace returns a new collection where all instances of the original entity
// (matched by ID) are replaced with the replacement entity.
func (c Collection) Replace(original Entity, replacement Entity) (Collection, bool) {
	newCollection := make(Collection, 0)
	found := false

	for _, currentEntity := range c {
		if !uuid.Equal(currentEntity.ID, original.ID) {
			newCollection = append(newCollection, currentEntity)
		} else {
			found = true
			newCollection = append(newCollection, replacement)
		}
	}

	return newCollection, found
}

// Replace returns a new collection where all instances of the original entity
// (matched by ID) are replaced with the replacement entity.
// If no entities were replaced, this call panics.
func (c Collection) MustReplace(original, target Entity) Collection {
	col, ok := c.Replace(original, target)
	if !ok {
		panic(fmt.Sprintf("could not replace target: %s", target))
	}

	return col
}

// Update returns a new collection where all instances of the original entity
// (matched by ID) are updated/replaced with the updated version of the entity.
func (c Collection) Update(target Entity) (Collection, bool) {
	newCollection := make(Collection, 0)
	found := false

	for _, currentEntity := range c {
		if !uuid.Equal(currentEntity.ID, target.ID) {
			newCollection = append(newCollection, currentEntity)
		} else {
			found = true
			newCollection = append(newCollection, target)
		}
	}

	return newCollection, found
}

// Update returns a new collection where all instances of the original entity
// (matched by ID) are updated/replaced with the updated version of the entity.
// If no entities were replaced, this call panics.
func (c Collection) MustUpdate(target Entity) Collection {
	col, ok := c.Update(target)
	if !ok {
		panic(fmt.Sprintf("could not update target: %s", target))
	}

	return col
}

// Clear will remove all entities in the current collection.
// This is just a wrapper over MakeCollection.
func (c Collection) Clear() Collection {
	return MakeCollection()
}

// Timer is used to keep track of when an action is next allowed.
type Timer struct {
	NextTimestamp int64 `json:"next_timestamp"`
}

// Delay is used to increase the delay on the current timer.
func (timer *Timer) Delay(seconds int) {
	timer.NextTimestamp = time.Now().Unix() + int64(seconds)
}

// Ready returns true when the current time elapses the timer's internal
// timestamp.
func (timer Timer) Ready() bool {
	return time.Now().Unix() >= timer.NextTimestamp
}

// NewTimer returns a new timer. I dont know why. TODO: delete this?
func NewTimer() *Timer {
	return new(Timer)
}

// Position is used to represent XY coordinates.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PassabilityType represents the different types of passability for an entity.
type PassabilityType int

// Available passability types:
const (
	AlwaysImpassible PassabilityType = iota
	AlwaysPassible
	//Keyed
	Toggleable
)

// Passability is a component used on entities to describe their state of
// passibility.
type Passability struct {
	Type   PassabilityType `json:"type"`
	IsOpen bool            `json:"is_open"`
}

// Attribute represents a DnD-style attribute.
type Attribute int

// Modifier calculates and returns a modifier based on the state of the current
// attribute.
// Minimum value is -5. At level 30, it would be 10.
func (attr Attribute) Modifier() int {
	if int(attr) < 0 {
		return -5
	}

	// formula f(x) = floor(0.5 * x) - 5
	modifier := math.Floor((0.5 * float64(attr))) - 5
	return int(modifier)
}

// Attributes is a component used to group all available attributes together
// which can exist for an entity.
type Attributes struct {
	Dexterity Attribute `json:"dexterity"`
	Luck      Attribute `json:"attribute"`
	Strength  Attribute `json:"strength"`
	Wisdom    Attribute `json:"wisdom"`
}

// Inventory is a component for entities which is used to store UUIDs
// correlating to items, which implies ownership by the entity over these items.
type Inventory struct {
	ItemIDs []uuid.UUID `json:"item_ids"`
}

// Equipment is a component used for specifying which items are currently
// "equiped" for a given entity.
type Equipment struct {
	HeadID          uuid.UUID `json:"head_id"`
	HandsID         uuid.UUID `json:"hands_id"`
	PrimaryItemID   uuid.UUID `json:"primary_item_id"`
	SecondaryItemID uuid.UUID `json:"secondary_item_id"`
	LegsID          uuid.UUID `json:"legs_id"`
	ChestID         uuid.UUID `json:"chest_id"`
}

// Entity is used to represent the amalgamation of various components and
// attributes for a single in-world object.
type Entity struct {
	ID          uuid.UUID   `json:"id"`
	Timer       Timer       `json:"timer"`
	Position    Position    `json:"position"`
	Passability Passability `json:"passability"`
	Attributes  Attributes  `json:"attributes"`
	Equipment   Equipment   `json:"equipment"`
	Health      int         `json:"health"`
}

func (e Entity) String() string {
	return fmt.Sprintf("Entity(%s)", e.ID.String())
}

// New instantiates a new Entity instance and returns it's pointer.
func New() *Entity {
	var err error

	e := new(Entity)
	e.ID, err = uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return e
}

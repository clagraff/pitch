package responses

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
	"github.com/clagraff/pitch/logging"
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"
)

type payload struct {
	Type     string          `json:"type"`
	Response json.RawMessage `json:"response"`
}

// Unmarshal bites representing a payload and return the embedded Response
// instance.
func Unmarshal(bites []byte) (Response, error) {
	logger, closeLog := logging.Logger("comms.responses.Unmarshal")
	defer closeLog()

	logger.Println("received response json to unmarshal")

	p := payload{}
	err := json.Unmarshal(bites, &p)
	if err != nil {
		return nil, errors.New(err)
	}

	logger.Println("going to unmarshal response:", p.Type)

	var resp Response

	switch p.Type {
	case reflect.TypeOf(MeleeAttackResponse{}).Name():
		r := MeleeAttackResponse{}
		err = json.Unmarshal(p.Response, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		resp = r
	case reflect.TypeOf(RangeAttackResponse{}).Name():
		r := RangeAttackResponse{}
		err = json.Unmarshal(p.Response, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		resp = r
	case reflect.TypeOf(MoveResponse{}).Name():
		r := MoveResponse{}
		err = json.Unmarshal(p.Response, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		resp = r
	case reflect.TypeOf(ToggleResponse{}).Name():
		r := ToggleResponse{}
		err = json.Unmarshal(p.Response, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		resp = r
	case reflect.TypeOf(ViewResponse{}).Name():
		r := ViewResponse{}
		err = json.Unmarshal(p.Response, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		resp = r
	default:
		return nil, errors.New("invalid response type")
	}

	logger.Println("successfully unmashalled response:", p.Type)
	return resp, nil
}

// Marshal a response into bites representing a payload.
func Marshal(resp Response) ([]byte, error) {
	bites, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New(err)
	}

	p := payload{
		Type:     reflect.TypeOf(resp).Name(),
		Response: json.RawMessage(bites),
	}

	bites, err = json.Marshal(p)
	if err != nil {
		return nil, errors.New(err)
	}
	return bites, nil
}

// Response provides an interfaces for exchanging and operating on responses
// from the server based on requests previously sent.
type Response interface {
	Apply(entities.World) (entities.World, error)
	IDs() []uuid.UUID
}

type Wrapper struct {
	Responses []Response `json:"responses"`
}

func (resp Wrapper) Apply(world entities.World) (entities.World, error) {
	var err error
	for _, r := range resp.Responses {
		world, err = r.Apply(world)
		if err != nil {
			return world, err
		}
	}

	return world, nil
}

func (resp Wrapper) IDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, r := range resp.Responses {
		ids = append(ids, r.IDs()...)
	}

	return ids
}

func MakeWrapper(resps ...Response) Wrapper {
	return Wrapper{
		Responses: resps,
	}
}

// MeleeAttackResponse is a response for providing details about the result
// of a melee request.
type MeleeAttackResponse struct {
	AttackerID      uuid.UUID `json:"attacker_id"`
	TargetID        uuid.UUID `json:"target_id"`
	DidHit          bool      `json:"did_hit"`
	Damage          int       `json:"damage"`
	HealthRemaining int       `json:"health_remaining"`
}

// Apply will apply the results of the melee attack against the current game
// world.
func (resp MeleeAttackResponse) Apply(world entities.World) (entities.World, error) {
	target, ok := world.Objects.FromID(resp.TargetID)
	if !ok {
		return world, fmt.Errorf("target could not be found on world")
	}

	target.Health = resp.HealthRemaining
	if target.Health <= 0 {
		target.Health = 0
		world.Objects = world.Objects.MustRemove(target)
	} else {
		world.Objects = world.Objects.MustUpdate(target)
	}

	return world, nil
}

func (resp MeleeAttackResponse) IDs() []uuid.UUID {
	return []uuid.UUID{resp.AttackerID, resp.TargetID}
}

// RangeAttackResponse is a response for providing details about the result
// of a melee request.
type RangeAttackResponse struct {
	AttackerID      uuid.UUID `json:"attacker_id"`
	TargetID        uuid.UUID `json:"target_id"`
	DidHit          bool      `json:"did_hit"`
	Damage          int       `json:"damage"`
	HealthRemaining int       `json:"health_remaining"`
}

// Apply will apply the results of the range attack against the current game
// world.
func (resp RangeAttackResponse) Apply(world entities.World) (entities.World, error) {
	target, ok := world.Objects.FromID(resp.TargetID)
	if !ok {
		return world, fmt.Errorf("target could not be found on world")
	}

	target.Health = resp.HealthRemaining
	if target.Health <= 0 {
		target.Health = 0

		world.Objects = world.Objects.MustRemove(target)
	} else {
		world.Objects = world.Objects.MustUpdate(target)
	}

	return world, nil
}

func (resp RangeAttackResponse) IDs() []uuid.UUID {
	return []uuid.UUID{resp.AttackerID, resp.TargetID}
}

// ToggleResponse is a response for providing details about the result
// of a open or close request.
type ToggleResponse struct {
	ActorID  uuid.UUID `json:"actor_id"`
	TargetID uuid.UUID `json:"target_id"`
	IsOpen   bool      `json:"is_closed"`
}

// Apply will apply the results of the open/close request against the current
// game world.
func (resp ToggleResponse) Apply(world entities.World) (entities.World, error) {
	logger, closeLog := logging.Logger("responses.ToggleResponse.Apply")
	defer closeLog()

	target, ok := world.Objects.FromID(resp.TargetID)
	if !ok {
		return world, fmt.Errorf("target could not be found on world")
	}

	logger.Println("Toggling entity:", target, "to open?", resp.IsOpen)
	target.Passability.IsOpen = resp.IsOpen

	world.Objects = world.Objects.MustUpdate(target)

	return world, nil
}

func (resp ToggleResponse) IDs() []uuid.UUID {
	return []uuid.UUID{resp.ActorID, resp.TargetID}
}

// MoveResponse is a response for providing details about the result
// of a movement request.
type MoveResponse struct {
	ActorID uuid.UUID `json:"actor_id"`
	X       int       `json:"x"`
	Y       int       `json:"y"`
}

// Apply will apply the results of the movementrequest against the current
// game world.
func (resp MoveResponse) Apply(world entities.World) (entities.World, error) {
	actor, ok := world.Objects.FromID(resp.ActorID)
	if !ok {
		return world, fmt.Errorf("actor could not be found on world")
	}

	actor.Position.X = resp.X
	actor.Position.Y = resp.Y

	world.Objects = world.Objects.MustUpdate(actor)
	return world, nil
}

func (resp MoveResponse) IDs() []uuid.UUID {
	return []uuid.UUID{resp.ActorID}
}

// ViewResponse is a response for providing details visible entities near the
// actor.
type ViewResponse struct {
	ActorID uuid.UUID        `json:"actor_id"`
	Objects []objects.Entity `json:"objects"`
}

// Apply will apply the results of the view request by clearing current objects
// and adding the response objects.
func (resp ViewResponse) Apply(world entities.World) (entities.World, error) {
	logger, closeLog := logging.Logger("responses.ViewResponse.Apply")
	defer closeLog()

	logger.Println("Number of cleared objects:", len(world.Objects))

	c := objects.MakeCollection()
	for _, obj := range resp.Objects {
		o := obj
		logger.Println("Inserting object into world:", o.ID.String())
		c = c.Append(o)
	}

	world.Objects = c

	return world, nil
}

func (resp ViewResponse) IDs() []uuid.UUID {
	return []uuid.UUID{resp.ActorID}
}

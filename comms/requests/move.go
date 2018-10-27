package requests

import (
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
)

func coordsFromDirection(actor objects.Entity, direction Direction) (int, int) {
	x := actor.Position.X
	y := actor.Position.Y

	switch direction {
	case North:
		y--
	case East:
		x++
	case South:
		y++
	case West:
		x--
	}

	return x, y
}

// MoveRequest is used to request that the specified target moves in a desired
// direction.
// This action assumes changing by 1 unit-space at a time. Hense the use of a
// `direction` rather than xy-position.
type MoveRequest struct {
	ActorID   uuid.UUID `json:"actor_id"`
	Direction Direction `json:"direction"`
}

// Execute will perform the movement request for the specified actor.
func (req MoveRequest) Execute(world entities.World) (entities.World, responses.Response, error) {
	actor, ok := world.Objects.FromID(req.ActorID)
	if !ok {
		return world, nil, errors.Errorf("actor %s could not be found", req.ActorID)
	}

	x, y := coordsFromDirection(actor, req.Direction)

	// It is okay if there are no objects at the new coords. That is why we
	// ignore the second return arg.
	nearbyEntities := world.Objects.FromXY(x, y)
	for _, e := range nearbyEntities {
		if e.Passability.Type == objects.AlwaysImpassible {
			attackReq := MeleeAttackRequest{
				AttackerID: actor.ID,
				TargetID:   e.ID,
			}
			world, resp, err := attackReq.Execute(world)
			return world, resp, err
		}
		if e.Passability.Type == objects.Toggleable {
			if !e.Passability.IsOpen {
				openReq := OpenRequest{
					ActorID:  actor.ID,
					TargetID: e.ID,
				}
				world, resp, err := openReq.Execute(world)
				return world, resp, err
			}
		}
	}

	actor.Position.X = x
	actor.Position.Y = y

	world.Objects = world.Objects.MustUpdate(actor)

	return world, responses.MoveResponse{
		ActorID: req.ActorID,
		X:       x,
		Y:       y,
	}, nil
}

package requests

import (
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
)

// OpenRequest represents a request to open a toggleable passibility entity.
type OpenRequest struct {
	ActorID  uuid.UUID `json:"actor_id"`
	TargetID uuid.UUID `json:"target_id"`
}

// Execute will attempt to open the specified target by the provided actor.
func (req OpenRequest) Execute(world entities.World) (entities.World, responses.Response, error) {
	actor, ok := world.Objects.FromID(req.ActorID)
	if !ok {
		return world, nil, errors.Errorf("actor could not be found")
	}

	target, ok := world.Objects.FromID(req.TargetID)
	if !ok {
		return world, nil, errors.Errorf("target could not be found")
	}

	if !actor.Timer.Ready() {
		return world, nil, errors.Errorf("timer not ready")
	}

	if target.Passability.Type != objects.Toggleable {
		return world, nil, errors.Errorf("target is not openable")
	}

	target.Passability.IsOpen = true
	world.Objects = world.Objects.MustUpdate(target)

	return world, responses.ToggleResponse{
		ActorID:  req.ActorID,
		TargetID: req.TargetID,
		IsOpen:   true,
	}, nil
}

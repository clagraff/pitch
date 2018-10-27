package requests

import (
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
	"github.com/clagraff/pitch/logging"
)

// ViewRequest represents a request determine perceivable nearby objects.
type ViewRequest struct {
	ActorID uuid.UUID `json:"actor_id"`
}

// Execute will find and return a response containing perceivable objects.
func (req ViewRequest) Execute(world entities.World) (entities.World, responses.Response, error) {
	logger, closeLog := logging.Logger("requests.ViewRequest.Execute")
	defer closeLog()

	actor, ok := world.Objects.FromID(req.ActorID)
	if !ok {
		return world, nil, errors.Errorf("actor could not be found")
	}

	if !actor.Timer.Ready() {
		return world, nil, errors.Errorf("timer not ready")
	}

	logger.Println("Requestor ID:", req.ActorID.String())

	resp := responses.ViewResponse{}
	resp.ActorID = req.ActorID
	resp.Objects = make([]objects.Entity, 0)

	viewDist := actor.Attributes.Wisdom.Modifier() + 10
	logger.Println("View distance:", viewDist)

	startX := actor.Position.X - viewDist
	endX := actor.Position.X + viewDist

	startY := actor.Position.Y - viewDist
	endY := actor.Position.Y + viewDist

	for x := startX; x < endX; x++ {
		for y := startY; y < endY; y++ {
			nearbyEntities := world.Objects.FromXY(x, y)
			for _, obj := range nearbyEntities {
				logger.Println("Object within view distance:", obj.ID.String())
				resp.Objects = append(resp.Objects, obj)
			}
		}
	}

	return world, resp, nil
}

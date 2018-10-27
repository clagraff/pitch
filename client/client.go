package client

import (
	"math"
	"reflect"

	termbox "github.com/nsf/termbox-go"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/requests"
	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
	"github.com/clagraff/pitch/logging"
)

// Scene represents a current state for UI rendering.
type Scene func(chan requests.Request, chan responses.Response, entities.World) Scene

// PosDiff will determine the absolute, rounded, distance between the two
// specified points.
func PosDiff(x1, y1, x2, y2 int) int {
	xDiff := x1 - x2
	yDiff := y1 - y2

	xDiffSq := xDiff * xDiff
	yDiffSq := yDiff * yDiff

	sum := xDiffSq + yDiffSq
	sq := int(math.Sqrt(float64(sum)))

	return sq
}

func asyncEventPoll(events chan termbox.Event) {
	for {
		events <- termbox.PollEvent()
	}
}

func renderCell(entity objects.Entity, playerID uuid.UUID) {
	logger, closeLog := logging.Logger("client.renderCell")
	defer closeLog()

	ch := '#'
	if entity.ID == playerID {
		ch = '@'
	}
	if entity.Passability.Type == objects.Toggleable {
		logger.Println("Rendering toggleable entity:", entity)
		switch entity.Passability.IsOpen {
		case true:
			ch = '\''
		case false:
			ch = '+'
		}
	}
	termbox.SetCell(
		entity.Position.X,
		entity.Position.Y,
		ch,
		termbox.ColorWhite,
		termbox.ColorBlack,
	)
}

func sendMoveRequest(key termbox.Key, actorID uuid.UUID, reqs chan requests.Request) {
	logger, closeLog := logging.Logger("client.sendMoveRequest")
	defer closeLog()

	var dir requests.Direction

	switch key {
	case termbox.KeyArrowUp:
		dir = requests.North
	case termbox.KeyArrowDown:
		dir = requests.South
	case termbox.KeyArrowLeft:
		dir = requests.West
	case termbox.KeyArrowRight:
		dir = requests.East
	default:
		logger.Println("Invalid termbox key")
		panic("invalid termbox key")
	}

	reqs <- requests.MoveRequest{ActorID: actorID, Direction: dir}
	reqs <- requests.ViewRequest{ActorID: actorID}
}

func tryGetTogglable(targets []objects.Entity) (objects.Entity, bool) {
	logger, closeLog := logging.Logger("client.tryGetTogglable")
	defer closeLog()

	for _, e := range targets {
		if e.Passability.Type == objects.Toggleable {
			logger.Println("entity:", e.ID.String(), "can be toggled")
			logger.Println("entity:", e.ID.String(), "is open?", e.Passability.IsOpen)
			logger.Printf("%p\n", e)
			return e, e.Passability.IsOpen
		}
		logger.Println("entity:", e.ID.String(), "is not toggleable")
	}

	return objects.Entity{}, false
}

func sendToggleRequest(world entities.World, actor objects.Entity, reqs chan requests.Request) {
	logger, closeLog := logging.Logger("client.sendToggleRequest")
	defer closeLog()

	x := actor.Position.X
	y := actor.Position.Y

	down := world.Objects.FromXY(x, y+1)
	up := world.Objects.FromXY(x, y-1)
	left := world.Objects.FromXY(x-1, y)
	right := world.Objects.FromXY(x+1, y)

	directions := [][]objects.Entity{down, up, left, right}
	for _, direction := range directions {
		target, isOpen := tryGetTogglable(direction)
		if target != (objects.Entity{}) {
			if isOpen {
				reqs <- requests.CloseRequest{ActorID: actor.ID, TargetID: target.ID}
				logger.Println("Open; closing target")
				return
			} else {
				reqs <- requests.OpenRequest{ActorID: actor.ID, TargetID: target.ID}
				logger.Println("Closed; opening target")
				return
			}
		}
	}
}

func renderWorld(world entities.World, playerID uuid.UUID) {
	logger, closeLog := logging.Logger("client.renderWorld")
	defer closeLog()

	err := termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	if err != nil {
		panic(err)
	}

	for _, e := range world.Objects {
		logger.Println("Rendering entity:", e)
		renderCell(e, playerID)
	}

	err = termbox.Flush()
	if err != nil {
		panic(err)
	}
}

func processResponses(world entities.World, resps chan responses.Response) entities.World {
	logger, closeLog := logging.Logger("client.processResponses")
	defer closeLog()

	i := 0
	doLoop := true
	for doLoop == true {
		select {
		case resp := <-resps:
			i++
			logger.Printf("Handling response type: %s\n", reflect.TypeOf(resp))

			var err error
			world, err = resp.Apply(world)
			if err != nil {
				logger.Println(err)
				panic(err)
			}

		default: // nolint:megacheck
			doLoop = false
			break
		}
	}

	return world
}

func isExitEvent(ev termbox.Event) bool {
	return ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC || ev.Key == termbox.KeyEsc
}

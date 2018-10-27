package asciiclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"reflect"
	"time"

	"github.com/clagraff/pitch/comms/requests"
	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/objects"
	"github.com/clagraff/pitch/logging"
	"github.com/go-errors/errors"
	termbox "github.com/nsf/termbox-go"
	uuid "github.com/satori/go.uuid"
)

var delimiter = byte('\n')

const chanBuffSize = 20

type connHandshake struct {
	ID        uuid.UUID `json:"id"`
	Subscribe bool      `json:"subscribe"`
}

func handshake(id uuid.UUID, conn net.Conn, subscribe bool) error {
	h := connHandshake{
		ID:        id,
		Subscribe: subscribe,
	}

	logger, closeLog := logging.Logger("asciiclient.handshake")
	defer closeLog()

	// Do handshake
	logger.Println("starting handshake for ID:", id.String())

	bites, err := json.Marshal(h)
	if err != nil {
		return errors.New(err)
	}

	logger.Println("sending client id for handshake")

	conn.Write(bites)
	conn.Write([]byte{delimiter})
	conn.SetDeadline(time.Now().Add(2 * time.Minute))

	logger.Println("waiting for handshake response on:", id.String())

	message, err := bufio.NewReader(conn).ReadBytes(delimiter)
	if err != nil {
		return errors.New(err)
	}
	message = message[:len(message)-1]

	validationID := uuid.UUID{}
	err = json.Unmarshal(message, &validationID)
	if err != nil {
		return errors.New(err)
	}

	logger.Println("validating handshake response for:", id.String())
	if !uuid.Equal(id, validationID) {
		logger.Println("handshake failed due to mismatched UUIDs on", id.String())
		return errors.New("handshake failed due to mismatched UUIDs")
	}

	logger.Println("handshake successful with:", id.String())
	return nil
}

func ManageRequests(host string, port int, id uuid.UUID) (chan<- requests.Request, error) {
	logger, closeLog := logging.Logger("asciiclient.ManageRequests")
	defer closeLog()

	reqs := make(chan requests.Request, chanBuffSize)

	logger.Printf("dialing server %s:%d\n", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return reqs, errors.New(err)
	}

	err = handshake(id, conn, false)
	if err != nil {
		logger.Println("handshake failed")
		return reqs, errors.New(err)
	}

	go func(c net.Conn, r <-chan requests.Request) {
		logger, closeLog := logging.Logger("asciiclient.ManageRequests.goFunc")
		defer closeLog()
		defer conn.Close()

		logger.Println("awaiting requests to send")
		for req := range r {
			bites, err := requests.Marshal(req)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}

			logger.Println("sending request:", reflect.TypeOf(req))
			logger.Println("request json:", string(bites))

			conn.Write(bites)
			conn.Write([]byte{delimiter})
			conn.SetDeadline(time.Now().Add(2 * time.Minute))

			logger.Println("request sent:", reflect.TypeOf(req))
		}
	}(conn, reqs)

	return reqs, nil
}

func ManageResponses(host string, port int, id uuid.UUID) (<-chan entities.World, error) {
	logger, closeLog := logging.Logger("asciiclient.ManageResponses")
	defer closeLog()

	worlds := make(chan entities.World, chanBuffSize)

	logger.Printf("dialing server %s:%d\n", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return worlds, errors.New(err)
	}

	err = handshake(id, conn, true)
	if err != nil {
		logger.Println("handshake failed")
		return worlds, errors.New(err)
	}

	go func(c net.Conn, w chan<- entities.World) {
		logger, closeLog := logging.Logger("asciiclient.ManageResponses.goFunc")
		defer closeLog()
		defer conn.Close()

		world := entities.MakeWorld()

		buff := bufio.NewReader(conn)
		for {
			logger.Println("awaiting responses")
			message, err := buff.ReadBytes(delimiter)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}
			conn.SetDeadline(time.Now().Add(2 * time.Minute))
			message = message[:len(message)-1]
			logger.Println("unmarshalling json response")

			resp, err := responses.Unmarshal(message)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}

			logger.Println("applying response:", reflect.TypeOf(resp))
			world, err = resp.Apply(world)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}
			logger.Println("applied response:", reflect.TypeOf(resp))

			logger.Println("sending world into worldChan")
			w <- world
			logger.Println("world sent into worldChan successfully")
		}
	}(conn, worlds)

	return worlds, nil
}

func Run(host string, port int, id uuid.UUID) error {
	logger, closeLog := logging.Logger("asciiclient.Run")
	defer closeLog()

	reqs, err := ManageRequests(host, port, id)
	if err != nil {
		return err
	}

	worlds, err := ManageResponses(host, port, id)
	if err != nil {
		return err
	}

	logger.Println("starting termbox")

	// start termbox
	err = termbox.Init()
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	logger.Println("termbox has been initialized")

	logger.Println("starting async event poll for termbox")

	events := make(chan termbox.Event)
	go asyncEventPoll(events)
	var ev termbox.Event

	logger.Println("emitting initial view request")
	reqs <- requests.ViewRequest{ActorID: id}
	logger.Println("emitted initial view request")

	world := <-worlds

	logger.Println("beginning gameplay loop")

	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	go func(reqs chan<- requests.Request) {
		for {
			select {
			case <-ticker.C:
				reqs <- requests.ViewRequest{ActorID: id}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(reqs)

	for {
		for doLoop := true; doLoop; {
			select {
			case world = <-worlds:
			default:
				doLoop = false
				break
				// pass
			}
		}
		player, found := world.Objects.FromIDString(id.String())
		if !found {
			err = fmt.Errorf("could not find player entity")
			stack := errors.New(err).ErrorStack()
			logger.Printf("%s\n", stack)
			panic(stack)
		}
		renderWorld(world, player.ID)

		select {
		case ev = <-events:
			// capture event in `ev` to be used later...
		default:
			// ...otherwise re-render and wait for an actual event.
			continue
		}

		switch ev.Type {
		case termbox.EventKey:
			if isExitEvent(ev) {
				return nil
			}

			switch ev.Key {
			case termbox.KeyArrowUp, termbox.KeyArrowDown, termbox.KeyArrowLeft, termbox.KeyArrowRight:
				sendMoveRequest(ev.Key, player.ID, reqs)
			}

			switch ev.Ch {
			case 'c', 'C':
				sendToggleRequest(world, player, reqs)
			}

		case termbox.EventError:
			logger.Fatal(ev.Err)
		}
		if err != nil {
			stack := errors.New(err).ErrorStack()
			logger.Printf("%s\n", stack)
			panic(stack)
		}

	}

	return nil
}

func renderWorld(world entities.World, playerID uuid.UUID) {
	logger, closeLog := logging.Logger("asciiclient.renderWorld")
	defer closeLog()

	err := termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	for _, e := range world.Objects {
		logger.Println("Rendering entity:", e)
		renderCell(e, playerID)
	}

	err = termbox.Flush()
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}
}

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
	logger, closeLog := logging.Logger("asciiclient.renderCell")
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

func isExitEvent(ev termbox.Event) bool {
	return ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC || ev.Key == termbox.KeyEsc
}

func sendMoveRequest(key termbox.Key, actorID uuid.UUID, reqs chan<- requests.Request) {
	logger, closeLog := logging.Logger("asciiclient.sendMoveRequest")
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
		err := fmt.Errorf("invalid direction type")
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	reqs <- requests.MoveRequest{ActorID: actorID, Direction: dir}
	reqs <- requests.ViewRequest{ActorID: actorID}
}

func tryGetTogglable(targets []objects.Entity) (objects.Entity, bool) {
	logger, closeLog := logging.Logger("asciiclient.tryGetTogglable")
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

func sendToggleRequest(world entities.World, actor objects.Entity, reqs chan<- requests.Request) {
	logger, closeLog := logging.Logger("asciiclient.sendToggleRequest")
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

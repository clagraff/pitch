package client

import (
	"github.com/clagraff/pitch/comms/requests"
	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/logging"
	termbox "github.com/nsf/termbox-go"
	uuid "github.com/satori/go.uuid"
)

// Gameplay is a scene used for managing the main interactions with the
// player's actor and seeing the immediate environment.
func Gameplay(reqs chan requests.Request, resps chan responses.Response, world entities.World) Scene {
	logger, closeLog := logging.Logger("client.gameplay")
	defer closeLog()

	events := make(chan termbox.Event)
	go asyncEventPoll(events)

	var ev termbox.Event
	var err error

	reqs <- requests.ViewRequest{
		ActorID: uuid.Must(uuid.FromString("b5d9c244-b17d-4845-bd56-07c710536008")),
	}

	select {
	case resp := <-resps:
		world, err = resp.Apply(world)
		if err != nil {
			panic(err)
		}
	}

	for {
		world = processResponses(world, resps)
		player, found := world.Objects.FromIDString("b5d9c244-b17d-4845-bd56-07c710536008")
		if !found {
			panic("could not find player entity")
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
			panic(err)
		}
	}
}

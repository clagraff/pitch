package main

import (
	"os"

	"github.com/clagraff/pitch/asciiclient"
	"github.com/clagraff/pitch/server"
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"
)

/*
func process(requests chan requests.Request, resps chan responses.Response, world entities.World) entities.World {
	logger, closeLog := logging.Logger("main.process")
	defer closeLog()

	var resp responses.Response
	var err error
	for request := range requests {
		logger.Println("received request")
		world, resp, err = request.Execute(world)
		if err != nil {
			if e, ok := err.(*errors.Error); ok {
				panic(e.ErrorStack())
			} else {
				panic(err)
			}
		}
		logger.Println("sending response")
		resps <- resp
		logger.Println("response sent")
	}

	return world
}
*/

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "server" {
		s := server.NewServer("localhost", 8080)
		err := s.Serve()
		if err != nil {
			panic(err)
		}
	} else if len(args) == 2 && args[0] == "client" {
		id, err := uuid.FromString(args[1])
		if err != nil {
			panic("invalid client uuid")
		}

		err = asciiclient.Run("localhost", 8080, id)
		if err != nil {
			if e, ok := err.(*errors.Error); ok {
				panic(e.ErrorStack())
			}
			panic(err)
		}
	} else {
		panic("invalid program arguments")
	}

	/*
		gameData, err := ioutil.ReadFile("game.save")
		if err != nil {
			panic(err)
		}

		world := entities.MakeWorld()
		err = json.Unmarshal(gameData, &world)
		if err != nil {
			panic(err)
		}

		entityList := world.Objects
		square := generation.Square(entityList[0], 5)
		square = generation.Offset(square, 7, 3)
		fill := generation.Fill(entityList[0], 5, 5)
		fill = generation.Offset(fill, 15, 15)
		entityList = append(entityList, square...)
		entityList = append(entityList, fill...)

		for _, e := range entityList {
			world.Objects = world.Objects.Append(e)
		}

		requests := make(chan requests.Request, 100)
		responses := make(chan responses.Response, 100)

		go process(requests, responses, world)

		err = termbox.Init()
		if err != nil {
			panic(err)
		}
		defer termbox.Close()
		termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

		scene := client.Gameplay

		//bites, _ := json.Marshal(&world)
		w := entities.MakeWorld()
		//_ = json.Unmarshal(bites, &w)

		for {
			if scene == nil {
				break
			}
			scene = scene(requests, responses, w)
		}
	*/
}

package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"time"

	"github.com/go-errors/errors"
	"github.com/olebedev/emitter"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/requests"
	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/logging"
)

var delimiter = byte('\n')

// Server is used to manage a TCP game server.
type Server struct {
	Emitter *emitter.Emitter
	Host    string
	Port    int
}

// NewServer returns a pointer to an instantiated Server instance.
func NewServer(host string, port int) *Server {
	logger, closeLog := logging.Logger("server.NewServer")
	defer closeLog()

	logger.Printf("Creating new server on %s:%d\n", host, port)

	s := Server{
		Emitter: &emitter.Emitter{},
		Host:    host,
		Port:    port,
	}

	return &s
}

// Serve incoming connections to the TCP server.
func (s Server) Serve() error {
	logger, closeLog := logging.Logger("server.Server.Serve")
	defer closeLog()

	logger.Printf("Server listening on %s:%d\n", s.Host, s.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return err
	}
	defer listener.Close()

	logger.Printf("starting process goroutine")
	go s.Process()

	for {
		logger.Printf("awaiting connection")
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		logger.Printf("received connection")
		go s.Handle(conn)
	}
}

type connHandshake struct {
	ID        uuid.UUID `json:"id"`
	Subscribe bool      `json:"subscribe"`
}

// Handle accepted connections to the server.
func (s Server) Handle(conn net.Conn) {
	logger, closeLog := logging.Logger("server.Server.Handle")
	defer closeLog()

	logger.Println("starting handshake")

	// Perform handshake
	message, err := bufio.NewReader(conn).ReadBytes(delimiter)
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}
	message = message[:len(message)-1]

	logger.Println("received handshake client id")

	h := connHandshake{}
	err = json.Unmarshal(message, &h)
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	logger.Println("successfully parse client id:", h.ID.String())
	logger.Println("sending handshake response")

	message, err = json.Marshal(h.ID)
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	conn.Write(message)
	conn.Write([]byte{delimiter})

	if h.Subscribe {
		// setup response writer
		writeResponses(s.Emitter, conn, h.ID)
		// listen for further requests
	} else {
		buff := bufio.NewReader(conn)
		for {
			logger.Println("awaiting request from", h.ID.String())

			message, err = buff.ReadBytes(delimiter)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}
			logger.Println("received request from:", h.ID.String())
			message = message[:len(message)-1]
			req, err := requests.Unmarshal(message)
			if err != nil {
				logger.Println("failed to unmarshal message from:", h.ID.String())
				logger.Println(string(message))
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
				//continue
			}
			logger.Println("emitting request:", reflect.TypeOf(req))
			<-s.Emitter.Emit("request", req)
			logger.Println("emitted request:", reflect.TypeOf(req))
			conn.SetDeadline(time.Now().Add(2 * time.Minute))
		}
	}
}

func writeResponses(e *emitter.Emitter, conn net.Conn, id uuid.UUID) {
	logger, closeLog := logging.Logger("server.writeResponses")
	defer closeLog()

	logger.Println("awaiting responses to write for:", id.String())

	for event := range e.On(id.String()) {
		if len(event.Args) == 1 {
			logger.Println("receiving response to write")
			resp := event.Args[0].(responses.Response)
			logger.Println("mashralling response:", reflect.TypeOf(resp))

			bites, err := responses.Marshal(resp)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
			}

			logger.Println("sending response:", reflect.TypeOf(resp))
			conn.Write(bites)
			conn.Write([]byte{delimiter})
			conn.SetDeadline(time.Now().Add(2 * time.Minute))
			logger.Println("sent response:", reflect.TypeOf(resp))
		}
	}
}

// Process requests against the game world.
func (s Server) Process() {
	logger, closeLog := logging.Logger("server.Server.Process")
	defer closeLog()

	logger.Println("loading game world")

	gameData, err := ioutil.ReadFile("game.save")
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	world := entities.MakeWorld()
	err = json.Unmarshal(gameData, &world)
	if err != nil {
		stack := errors.New(err).ErrorStack()
		logger.Printf("%s\n", stack)
		panic(stack)
	}

	logger.Println("game world loaded")
	logger.Println("await requests to process")

	for event := range s.Emitter.On("request") {
		if len(event.Args) == 1 {

			var (
				resp responses.Response
				err  error
			)
			req := event.Args[0].(requests.Request)
			logger.Println("processing request:", reflect.TypeOf(req))
			world, resp, err = req.Execute(world)
			if err != nil {
				stack := errors.New(err).ErrorStack()
				logger.Printf("%s\n", stack)
				panic(stack)
				// close connections?
			}
			ids := resp.IDs()
			for _, id := range ids {
				<-s.Emitter.Emit(id.String(), resp)
			}
		}
	}
}

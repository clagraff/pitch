package requests

import (
	"encoding/json"
	"reflect"

	"github.com/go-errors/errors"

	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/logging"
)

// Direction represents a 2D dimensional axis.
type Direction int

// The 4 valid directions.
// TODO: support diagnol directions.
const (
	North Direction = iota
	East
	South
	West
)

// Request is an interface used to specify desired actions from clients to the
// main server.
type Request interface {
	Execute(entities.World) (entities.World, responses.Response, error)
}

type payload struct {
	Type    string          `json:"type"`
	Request json.RawMessage `json:"request"`
}

// Unmarshal a bites into a payload which can be used to return the
// embedded request.
func Unmarshal(bites []byte) (Request, error) {
	logger, closeLog := logging.Logger("comms.requests.Unmarshal")
	defer closeLog()

	logger.Println("received request json to unmarshal")

	p := payload{}
	err := json.Unmarshal(bites, &p)
	if err != nil {
		return nil, errors.New(err)
	}

	logger.Println("going to unmarshal request:", p.Type)

	var req Request

	switch p.Type {
	case reflect.TypeOf(MeleeAttackRequest{}).Name():
		r := MeleeAttackRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r
	case reflect.TypeOf(RangeAttackRequest{}).Name():
		r := RangeAttackRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r

	case reflect.TypeOf(CloseRequest{}).Name():
		r := CloseRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r

	case reflect.TypeOf(MoveRequest{}).Name():
		r := MoveRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r

	case reflect.TypeOf(OpenRequest{}).Name():
		r := OpenRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r

	case reflect.TypeOf(ViewRequest{}).Name():
		r := ViewRequest{}
		err = json.Unmarshal(p.Request, &r)
		if err != nil {
			return nil, errors.New(err)
		}

		req = r

	default:
		logger.Println("invalid request type:", p.Type)
		return nil, errors.Errorf("invalid request type: %s", p.Type)
	}

	logger.Println("successfully unmashalled request:", p.Type)
	return req, nil
}

// Marshal a request into bites representing a payload instance.
func Marshal(req Request) ([]byte, error) {
	bites, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New(err)
	}

	p := payload{
		Type:    reflect.TypeOf(req).Name(),
		Request: json.RawMessage(bites),
	}

	bites, err = json.Marshal(p)
	if err != nil {
		return nil, errors.New(err)
	}
	return bites, nil
}

package logging

import (
	"fmt"
	"log"
	"os"
)

// Logger will setup and return a logger. It will automatically log to a file
// based on the name provided to the function.
func Logger(name string) (*log.Logger, func()) {
	fileName := fmt.Sprintf("logs/%s.log", name)

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	close := func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}

	return log.New(f, "", log.Lshortfile|log.Ltime), close
}

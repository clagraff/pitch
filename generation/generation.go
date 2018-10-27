package generation

import (
	"github.com/clagraff/pitch/entities/objects"
	uuid "github.com/satori/go.uuid"
)

// Square will return a list of entities which match the entity description
// provided, and whose position forms a square matching the size specified.
func Square(source objects.Entity, size int) []objects.Entity {
	entityList := make([]objects.Entity, (size*4)+4)
	i := 0

	// Make top
	for _, y := range []int{0, size + 1} {
		for x := 0; x < size+2; x++ {
			var err error
			e := source

			e.ID, err = uuid.NewV4()
			if err != nil {
				panic(err)
			}
			e.Position.X = x
			e.Position.Y = y

			entityList[i] = e
			i++
		}
	}

	for _, x := range []int{0, size + 1} {
		for y := 1; y < size+1; y++ {
			var err error
			e := source

			e.ID, err = uuid.NewV4()
			if err != nil {
				panic(err)
			}
			e.Position.X = x
			e.Position.Y = y

			entityList[i] = e
			i++
		}
	}

	return entityList
}

// Fill will return a list of entities which match the entity description
// provided, and whose position forms fills the rectangle specified by the
// height and width.
func Fill(source objects.Entity, height, width int) []objects.Entity {
	entityList := make([]objects.Entity, height*width)
	i := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var err error
			e := source

			e.ID, err = uuid.NewV4()
			if err != nil {
				panic(err)
			}
			e.Position.X = x
			e.Position.Y = y

			entityList[i] = e
			i++
		}
	}

	return entityList
}

// Offset will respotion all of the provided entities by the specified X and
// Y amounts.
func Offset(entityList []objects.Entity, xOffset, yOffset int) []objects.Entity {
	newEntities := make([]objects.Entity, len(entityList))

	for i, e := range entityList {
		e.Position.X = e.Position.X + xOffset
		e.Position.Y = e.Position.Y + yOffset
		newEntities[i] = e
	}

	return newEntities
}

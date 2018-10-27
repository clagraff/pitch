package utils

import "math/rand"

// Roll will keep a running total of generated a random numbers (from zero to
// the specified upper range) for `num` number of times, and return the total.
func Roll(num, upperRange int) int {
	total := 0
	for i := 0; i < num; i++ {
		r := rand.Intn(upperRange + 1)
		if r <= 0 {
			r = 1
		}

		total = total + r
	}

	return total
}

// RollWithAdvantage will perform two Roll calls, and return the higher of the
// two.
func RollWithAdvantage(num, upperRange int) int {
	firstRoll := Roll(num, upperRange)
	secondRoll := Roll(num, upperRange)
	if firstRoll > secondRoll {
		return firstRoll
	}
	return secondRoll
}

// RollWithDisadvantage will perform two Roll calls, and return the lower of the
// two.
func RollWithDisadvantage(num, upperRange int) int {
	firstRoll := Roll(num, upperRange)
	secondRoll := Roll(num, upperRange)
	if firstRoll > secondRoll {
		return secondRoll
	}
	return firstRoll
}

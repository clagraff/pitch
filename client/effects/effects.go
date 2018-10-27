package effects

import (
	"time"

	termbox "github.com/nsf/termbox-go"
)

// Cell is used to represent a single displayable tile.
type Cell struct {
	Left int
	Top  int
	Ch   rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

// Render will use termbox to render the current Cell to the screen.
func (c Cell) Render() {
	termbox.SetCell(c.Left, c.Top, c.Ch, c.Fg, c.Bg)
}

// Animation is used to iterate between multiple cells.
type Animation struct {
	Cells []Cell
	Delay time.Duration
	Index int
	Next  time.Time
}

// Render will render the current state of the animation, and handle
// determining when to swap to the next state.
func (a *Animation) Render() {
	now := time.Now()

	if now.After(a.Next) {
		a.Index++
		a.Next = now.Add(a.Delay)
		if a.Index >= len(a.Cells) {
			a.Index = 0
		}
	}

	a.Cells[a.Index].Render()
}

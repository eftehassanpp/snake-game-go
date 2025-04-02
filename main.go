package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Field struct {
	Size        int
	Icons       map[int]string
	SnakeCoords [][]int
	Field       [][]int
}

func NewField(size int) *Field {
	f := &Field{
		Size: size,
		Icons: map[int]string{
			0: " . ",
			1: " * ",
			2: " # ",
			3: " & ",
		},
		SnakeCoords: [][]int{},
	}
	f.generateField()
	f.addEntity()
	return f
}

func (f *Field) addEntity() {
	for {
		i := rand.Intn(f.Size)
		j := rand.Intn(f.Size)
		entity := []int{i, j}

		found := false
		for _, coord := range f.SnakeCoords {
			if coord[0] == entity[0] && coord[1] == entity[1] {
				found = true
				break
			}
		}

		if !found {
			f.Field[i][j] = 3
			break
		}
	}
}

func (f *Field) generateField() {
	f.Field = make([][]int, f.Size)
	for i := range f.Field {
		f.Field[i] = make([]int, f.Size)
	}
}

func (f *Field) clearField() {
	for i := range f.Field {
		for j := range f.Field[i] {
			if f.Field[i][j] == 1 || f.Field[i][j] == 2 {
				f.Field[i][j] = 0
			}
		}
	}
}

func (f *Field) render(screen tcell.Screen) {
	f.clearField()

	for _, coord := range f.SnakeCoords {
		f.Field[coord[0]][coord[1]] = 1
	}

	if len(f.SnakeCoords) > 0 {
		head := f.SnakeCoords[len(f.SnakeCoords)-1]
		f.Field[head[0]][head[1]] = 2
	}

	for i := 0; i < f.Size; i++ {
		row := ""
		for j := 0; j < f.Size; j++ {
			row += f.Icons[f.Field[i][j]]
		}
		for k, r := range row {
			screen.SetContent(k, i, r, nil, tcell.StyleDefault)
		}

	}
}

func (f *Field) getEntityPos() []int {
	for i := 0; i < f.Size; i++ {
		for j := 0; j < f.Size; j++ {
			if f.Field[i][j] == 3 {
				return []int{i, j}
			}
		}
	}
	return []int{-1, -1}
}

func (f *Field) isSnakeEatEntity() bool {
	entity := f.getEntityPos()
	if len(f.SnakeCoords) == 0 {
		return false
	}
	head := f.SnakeCoords[len(f.SnakeCoords)-1]
	return entity[0] == head[0] && entity[1] == head[1]
}

type Snake struct {
	Name      string
	Direction tcell.Key
	Coords    [][]int
	Field     *Field
}

func NewSnake(name string) *Snake {
	return &Snake{
		Name:      name,
		Direction: tcell.KeyRight,
		Coords:    [][]int{{0, 0}, {0, 1}, {0, 2}, {0, 3}},
	}
}

func (s *Snake) setDirection(ch tcell.Key) {
	if ch == tcell.KeyLeft && s.Direction == tcell.KeyRight {
		return
	}
	if ch == tcell.KeyRight && s.Direction == tcell.KeyLeft {
		return
	}
	if ch == tcell.KeyUp && s.Direction == tcell.KeyDown {
		return
	}
	if ch == tcell.KeyDown && s.Direction == tcell.KeyUp {
		return
	}
	s.Direction = ch
}

func (s *Snake) levelUp() {
	a := s.Coords[0]
	b := s.Coords[1]

	tail := make([]int, 2)
	copy(tail, a)

	if a[0] < b[0] {
		tail[0]--
	} else if a[1] < b[1] {
		tail[1]--
	} else if a[0] > b[0] {
		tail[0]++
	} else if a[1] > b[1] {
		tail[1]++
	}

	tail = s.checkLimit(tail)
	s.Coords = append([][]int{tail}, s.Coords...)
}

func (s *Snake) isAlive() bool {
	if len(s.Coords) == 0 {
		return true
	}
	head := s.Coords[len(s.Coords)-1]
	snakeBody := s.Coords[:len(s.Coords)-1]

	for _, coord := range snakeBody {
		if coord[0] == head[0] && coord[1] == head[1] {
			return false
		}
	}
	return true
}

func (s *Snake) checkLimit(point []int) []int {
	if point[0] > s.Field.Size-1 {
		point[0] = 0
	} else if point[0] < 0 {
		point[0] = s.Field.Size - 1
	} else if point[1] < 0 {
		point[1] = s.Field.Size - 1
	} else if point[1] > s.Field.Size-1 {
		point[1] = 0
	}
	return point
}

func (s *Snake) move() {
	head := make([]int, 2)
	copy(head, s.Coords[len(s.Coords)-1])

	if s.Direction == tcell.KeyUp {
		head[0]--
	} else if s.Direction == tcell.KeyDown {
		head[0]++
	} else if s.Direction == tcell.KeyRight {
		head[1]++
	} else if s.Direction == tcell.KeyLeft {
		head[1]--
	}

	head = s.checkLimit(head)

	s.Coords = s.Coords[1:]
	s.Coords = append(s.Coords, head)
	s.Field.SnakeCoords = s.Coords

	if !s.isAlive() {
		fmt.Println("Game Over")
		os.Exit(0)
	}

	if s.Field.isSnakeEatEntity() {
		fmt.Print("\a") // Beep
		s.levelUp()
		s.Field.addEntity()
	}
}

func (s *Snake) setField(field *Field) {
	s.Field = field
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defer screen.Fini()

	screen.Clear()

	field := NewField(30)
	snake := NewSnake("Joe")
	snake.setField(field)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Goroutine for snake movement and rendering
	go func() {
		for range ticker.C {
			snake.move()
			field.render(screen)
			screen.Show()
		}
	}()

	// Main loop for event handling
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			}
			if ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyRight {
				snake.setDirection(ev.Key())
			}
		}
	}
}

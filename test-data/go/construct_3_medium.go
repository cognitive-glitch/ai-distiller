package medium

import "fmt"

type Mover interface {
	Move(x, y int)
}

// Point is a basic 2D point.
type Point struct {
	X, Y int
}

// Move implements the Mover interface for Point.
func (p *Point) Move(x, y int) {
	p.X += x
	p.Y += y
}

// Shape is an embedded struct.
type Shape struct {
	Name string
	Point // Anonymous field (embedding)
}

// ProcessShapes takes a slice of Movers and identifies their underlying type.
func ProcessShapes(shapes []Mover) (names []string) {
	for _, s := range shapes {
		// Type switch is a major parsing challenge.
		switch v := s.(type) {
		case *Shape:
			fmt.Printf("Shape '%s' at (%d, %d)\n", v.Name, v.X, v.Y)
			// Accessing fields from both embedded and outer struct.
			v.Move(10, 10)
			names = append(names, v.Name)
		case *Point:
			fmt.Printf("Just a Point at (%d, %d)\n", v.X, v.Y)
		default:
			// Type assertion inside the default case.
			if p, ok := s.(*Point); ok {
				fmt.Printf("Assertion found a point: %v\n", p)
			}
		}
	}
	return // Naked return
}
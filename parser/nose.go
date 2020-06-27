
package parser

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type Nose struct {

	links []Link
	properties map[string]float32
	position Vector
	rotate Matrix
}

func (nose *Nose) Links() []Link {
	return nose.links
}

func (nose *Nose) Properties() map[string]float32 {
	return nose.properties
}

func (nose *Nose) Position() *Vector {
	return &nose.position
}

func (nose *Nose) Rotate() *Matrix {
	return &nose.rotate
}

func (nose *Nose) AddToModel(model *solver.Model) {

	segments := int(nose.properties["Segments"])
	steps := int(nose.properties["Steps"])
	radius := nose.properties["Diameter"]*0.5
	length := nose.properties["Length"]
	lenStep := length/float32(steps)

	for i:=0; i<segments; i++ {
		a0 := float32(i)*2*3.1415926/float32(segments)
		a1 := float32(i+1)*2*3.1415926/float32(segments)
		y := length
		p0 := Point { 0, y, 0}
		p1 := Point { 0, y, 0}

		for j:=0; j<steps; j++ {

			frac := 1 - float32(j+1)/float32(steps)
			//r := radius*Sqrt(1 - frac*frac)
			r := radius*(1 - frac)
			y = y - lenStep

			p2 := Point { r*Cos(a0), y, r*Sin(a0) }
			p3 := Point { r*Cos(a1), y, r*Sin(a1) }

			var panel *solver.Panel
			if length > 0 {
				panel = &solver.Panel {
					Points: [4]Point{
						p0.Add(nose.position),
						p1.Add(nose.position),
						p3.Add(nose.position),
						p2.Add(nose.position),
					},
					Count: 4,
					Strength: 1,
				}
			} else {
				panel = &solver.Panel {
					Points: [4]Point{
						p2.Add(nose.position),
						p3.Add(nose.position),
						p1.Add(nose.position),
						p0.Add(nose.position),
					},
					Count: 4,
					Strength: 1,
				}
			}

			model.Panels = append(model.Panels, panel)

			p0 = p2
			p1 = p3
		}
	}
}

func MakeFakeNose(diameter float32, length float32) *Nose {

	return &Nose{
		links: nil,
		properties: map[string]float32 {
			"Segments": 12,
			"Steps": 5,
			"Diameter": diameter,
			"Length": length,
		},
		position: Vector{0, 0, 0},
		rotate: IdentityMatrix,
	}
}


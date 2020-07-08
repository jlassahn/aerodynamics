
package parser

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type slot struct {
	top float32
	bottom float32
}

type Tube struct {

	tag string
	links []Link
	properties map[string]float32
	position Vector
	rotate Matrix

	slots [][]slot // [segment][top to bottom]
}

func (tube *Tube) Links() []Link {
	return tube.links
}

func (tube *Tube) Properties() map[string]float32 {
	return tube.properties
}

func (tube *Tube) Position() *Vector {
	return &tube.position
}

func (tube *Tube) Rotate() *Matrix {
	return &tube.rotate
}

func (tube *Tube) AddToModel(model *solver.Model) {

	segments := int(tube.properties["Segments"])
	steps := int(tube.properties["Steps"])
	radius := tube.properties["Diameter"]*0.5
	length := tube.properties["Length"]
	lenStep := length/float32(steps)

	for i:=0; i<segments; i++ {
		a0 := float32(i)*2*3.1415926/float32(segments)
		a1 := float32(i+1)*2*3.1415926/float32(segments)
		p0 := Vector { radius*Cos(a0), length*0.5, radius*Sin(a0) }
		p1 := Vector { radius*Cos(a1), length*0.5, radius*Sin(a1) }
		for j:=0; j<steps; j++ {
			p2 := p0.Add(Vector{0, -lenStep, 0})
			p3 := p1.Add(Vector{0, -lenStep, 0})

			panel := &solver.Panel {
				Tag: tube.tag,
				IX: i,
				IY: j,
				Points: [4]Point{
					Point(tube.rotate.Transform(p0).Add(tube.position)),
					Point(tube.rotate.Transform(p1).Add(tube.position)),
					Point(tube.rotate.Transform(p3).Add(tube.position)),
					Point(tube.rotate.Transform(p2).Add(tube.position)),
				},
				Count: 4,
				Strength: 1,
			}

			model.Panels = append(model.Panels, panel)

			p0 = p2
			p1 = p3
		}
	}
}

func MakeFakeTube(tag string, diameter float32, length float32) *Tube {

	return &Tube{
		tag: tag,
		links: nil,
		properties: map[string]float32 {
			"Segments": 12,
			"Steps": 6,
			"Diameter": diameter,
			"Length": length,
		},
		position: Vector{0, 0, 0},
		rotate: IdentityMatrix,
		slots: nil,
	}
}



package parser

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type Nose struct {

	tag string
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

	curve := nose.properties["Style"]

	for i:=0; i<segments; i++ {
		a0 := float32(i)*2*3.1415926/float32(segments)
		a1 := float32(i+1)*2*3.1415926/float32(segments)
		y := length
		p0 := Vector { 0, y, 0}
		p1 := Vector { 0, y, 0}

		for j:=0; j<steps; j++ {

			frac := 1 - float32(j+1)/float32(steps)
			var r float32
			switch curve {
				case 0: r = Sqrt(1 - frac*frac)
				case 1: r = (1 - frac)
				default: r = (Sqrt(2-frac*frac)-1)/(1.414213562 - 1)
			}
			r = r*radius

			y = y - lenStep

			p2 := Vector { r*Cos(a0), y, r*Sin(a0) }
			p3 := Vector { r*Cos(a1), y, r*Sin(a1) }

			var panel *solver.Panel
			if j == 0 {
				panel = &solver.Panel {
					Tag: nose.tag,
					IX: i,
					IY: j,
					Points: [4]Point{
						Point(nose.rotate.Transform(p1).Add(nose.position)),
						Point(nose.rotate.Transform(p3).Add(nose.position)),
						Point(nose.rotate.Transform(p2).Add(nose.position)),
						Point{0,0,0},
					},
					Count: 3,
					Strength: 1,
				}
			} else {
				panel = &solver.Panel {
					Tag: nose.tag,
					IX: i,
					IY: j,
					Points: [4]Point{
						Point(nose.rotate.Transform(p0).Add(nose.position)),
						Point(nose.rotate.Transform(p1).Add(nose.position)),
						Point(nose.rotate.Transform(p3).Add(nose.position)),
						Point(nose.rotate.Transform(p2).Add(nose.position)),
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

func MakeFakeNose(tag string, diameter float32, length float32) *Nose {

	return &Nose{
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
	}
}

func MakeFakeTail(tag string, diameter float32, length float32) *Nose {
	ret := MakeFakeNose(tag, diameter, length)
	ret.rotate = Matrix{ [9]float32 {
		-1,  0,  0,
		 0, -1,  0,
		 0,  0,  1,
	}}
	return ret
}


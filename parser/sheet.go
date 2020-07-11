
package parser



import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type Sheet struct {
	tag string
	links []Link
	properties map[string]float32
	position Vector
	rotate Matrix
}

func (sheet *Sheet) Links() []Link {
	return sheet.links
}

func (sheet *Sheet) Properties() map[string]float32 {
	return sheet.properties
}

func (sheet *Sheet) Position() *Vector {
	return &sheet.position
}

func (sheet *Sheet) Rotate() *Matrix {
	return &sheet.rotate
}

func (sheet *Sheet) AddToModel(model *solver.Model) {

	wsteps := 5
	lsteps := 6
	width := sheet.properties["Width"]
	length := sheet.properties["Length"]
	sweep := sheet.properties["Sweep"]
	taper := sheet.properties["Taper"]
	thick := sheet.properties["Thick"]

	yt0 := length/2
	yb0 := -length/2
	x0 := float32(0)
	for i:=0; i<wsteps; i++ {
		yt1 := yt0 - sweep/float32(wsteps) - taper/float32(wsteps)
		yb1 := yb0 - sweep/float32(wsteps) + taper/float32(wsteps)
		x1 := x0 + width/float32(wsteps)
		span := float32(-1)

		z0 := float32(0)
		p0 := Vector{x0, yt0, 0}
		p1 := Vector{x1, yt1, 0}

		for j:=0; j<lsteps; j++ {
			span += 2/float32(lsteps)
			z1 := (1 - span*span)*thick
			p2 := p0
			p3 := p1
			p2.Y += (yb0 - yt0)/float32(lsteps)
			p3.Y += (yb1 - yt1)/float32(lsteps)

			p0.Z = -z0
			p1.Z = -z0
			p2.Z = -z1
			p3.Z = -z1
			panel := &solver.Panel {
				IX: i,
				IY: j,
				Tag: sheet.tag,
				Points: [4]Point{
					Point(sheet.rotate.Transform(p0).Add(sheet.position)),
					Point(sheet.rotate.Transform(p1).Add(sheet.position)),
					Point(sheet.rotate.Transform(p3).Add(sheet.position)),
					Point(sheet.rotate.Transform(p2).Add(sheet.position)),
				},
				Count: 4,
				Strength: 1,
			}
			model.Panels = append(model.Panels, panel)

			p0.Z = z0
			p1.Z = z0
			p2.Z = z1
			p3.Z = z1
			panel = &solver.Panel {
				IX: i,
				IY: j,
				Tag: sheet.tag,
				Points: [4]Point{
					Point(sheet.rotate.Transform(p1).Add(sheet.position)),
					Point(sheet.rotate.Transform(p0).Add(sheet.position)),
					Point(sheet.rotate.Transform(p2).Add(sheet.position)),
					Point(sheet.rotate.Transform(p3).Add(sheet.position)),
				},
				Count: 4,
				Strength: 1,
			}
			model.Panels = append(model.Panels, panel)

			p0 = p2
			p1 = p3
			z0 = z1
		}

		p0 = Vector{x1, -100, 0}
		p1 = Vector{x1, 0, 0}
		p2 := Vector{0,0,0}

		wake := &solver.Wake {
			Points: []Point {
				Point(sheet.rotate.Transform(p0).Add(sheet.position)),
				Point(sheet.rotate.Transform(p1).Add(sheet.position)),
				},
				TreePath: []Point {
					Point(sheet.rotate.Transform(p2).Add(sheet.position)),
				},
				BlurInternal: 0.0001,
				BlurWake: width/(4*float32(wsteps)),
				Strength: 0,
		}
		model.Wakes = append(model.Wakes, wake)

		p0 = Vector{(x0 + x1)/2, (yb0 + yb1)/2 - 3*solver.EPSILON, 0}
		p1 = Vector{0, 0, 0.1}
		edge := solver.Edge {
			Center: Point(sheet.rotate.Transform(p0).Add(sheet.position)),
			Normal: sheet.rotate.Transform(p1),
		}
		model.Edges = append(model.Edges, &edge)

		yt0 = yt1
		yb0 = yb1
		x0 = x1
	}
}

func MakeFakeSheet(tag string, width float32, length float32) *Sheet {

	return &Sheet{
		tag: tag,
		links: nil,
		properties: map[string]float32 {
			"Width": width,
			"Length": length,
			"Sweep": length*0,
			"Taper": length*0.25,
			"Thick": 0.05,
		},
		position: Vector{0, 0, 0},
		rotate: IdentityMatrix,
	}
}


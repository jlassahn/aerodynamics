
package solver

import (
	"math"

	. "github.com/jlassahn/aerodynamics/geometry"
)

const (
	EPSILON = 0.0001
	MIN_RANGE = 3
)

type Panel struct {
	Points [4]Point
	Count int
	Normal Vector
	Area float32
	MaxR float32

	Strength float32
}

func (panel *Panel) Center() Point {

	var p Point
	for i := 0; i < panel.Count; i++ {
		p.X += panel.Points[i].X
		p.Y += panel.Points[i].Y
		p.Z += panel.Points[i].Z
	}
	p.X = p.X/float32(panel.Count)
	p.Y = p.Y/float32(panel.Count)
	p.Z = p.Z/float32(panel.Count)

	return p
}

func (panel *Panel) InitStats() {

	var n Vector
	center := panel.Center()

	if panel.Count == 4 {
		n = n.Add(center.Sub(panel.Points[0]).Cross(center.Sub(panel.Points[1])))
		n = n.Add(center.Sub(panel.Points[1]).Cross(center.Sub(panel.Points[2])))
		n = n.Add(center.Sub(panel.Points[2]).Cross(center.Sub(panel.Points[3])))
		n = n.Add(center.Sub(panel.Points[3]).Cross(center.Sub(panel.Points[0])))
		n = n.Scale(0.5)
	} else {
		n = panel.Points[3].Sub(panel.Points[0]).Cross(panel.Points[3].Sub(panel.Points[1]))
	}

	panel.Area = Sqrt(n.Dot(n))
	panel.Normal = n.Scale(1/panel.Area)

	for i:=0; i<panel.Count; i++ {
		v := center.Sub(panel.Points[i])
		ln2 := Sqrt(v.Dot(v))
		if ln2 > panel.MaxR {
			panel.MaxR = ln2
		}
	}
}

func (panel *Panel) Velocity(pt Point) Vector {

	v := pt.Sub(panel.Center())
	ln := Sqrt(v.Dot(v))

	if ln < EPSILON {
		//FIXME can end up with a few times strength in the final vector
		//      because several subdivisions are added to it
		return panel.Normal.Scale(panel.Strength)
	}

	if ln < panel.MaxR*MIN_RANGE {
		return panel.Subdivide(pt)
	}

	scale := panel.Area*panel.Strength/(ln*ln*ln*2*math.Pi)
	return v.Scale(scale)
}

func (panel *Panel) Subdivide(pt Point) Vector {
	var v Vector
	var p Panel

	p.Count = panel.Count
	p.Normal = panel.Normal
	p.Area = panel.Area/4
	p.MaxR = panel.MaxR/2
	p.Strength = panel.Strength

	// FIXME all subpanels aren't actually the same area!
	if panel.Count == 4 {
		p1 := panel.Points[0].Average(panel.Points[1])
		p2 := panel.Points[1].Average(panel.Points[2])
		p3 := panel.Points[2].Average(panel.Points[3])
		p4 := panel.Points[3].Average(panel.Points[0])
		pc := p1.Average(p3)

		p.Points[0] = panel.Points[0]
		p.Points[1] = p1
		p.Points[2] = pc
		p.Points[3] = p4

		dv := p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = panel.Points[1]
		p.Points[1] = p2
		p.Points[2] = pc
		p.Points[3] = p1

		dv = p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = panel.Points[2]
		p.Points[1] = p3
		p.Points[2] = pc
		p.Points[3] = p2

		dv = p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = panel.Points[3]
		p.Points[1] = p4
		p.Points[2] = pc
		p.Points[3] = p3

		dv = p.Velocity(pt)
		v = v.Add(dv)

	} else {

		p1 := panel.Points[0].Average(panel.Points[1])
		p2 := panel.Points[1].Average(panel.Points[2])
		p3 := panel.Points[2].Average(panel.Points[0])

		p.Points[0] = panel.Points[0]
		p.Points[1] = p1
		p.Points[2] = p3

		dv := p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = panel.Points[1]
		p.Points[1] = p2
		p.Points[2] = p1

		dv = p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = panel.Points[2]
		p.Points[1] = p3
		p.Points[2] = p2

		dv = p.Velocity(pt)
		v = v.Add(dv)

		p.Points[0] = p1
		p.Points[1] = p2
		p.Points[2] = p3

		dv = p.Velocity(pt)
		v = v.Add(dv)
	}

	return v
}


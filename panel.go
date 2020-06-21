
package main

/* Panel Method Solver

Model is an arbitrary mesh of triangles and quads enclosing a volume.
Some points on the model are wake vortex sources, these include a path
to a tree that connects them internally.  There must be at least two wake
vortices.
Some edges of the model have vorticity shedding parameters, a perpendicular
velocity vector (which the solver tries to minimize) and a strength roughly
proportional to the curvature at the edge.  These can be calculated from
the geometry, but we might want to specify an edge sharpness adjustment.

Each polygon has a normal and an area (compute these).

Each polygon has a control point at the centroid.

The wake vortices start out as straight lines going downstream parallel to
the free stream velocity.  We might want to solve iteratively to make the
wakes follow the actual fluid flow.  The exact positions of the internal
tree of wake paths shouldn't matter as long as they connect.

Each polygon acts as a source of constant strength spread over the surface.
Each wake (and the internal paths connecting them) act as a vortex of constant
strength over the length of the path.

Transform the wake lines into N-1 combinations, each consisting of one wake
at strength +1, and all the others at strength -1/(N-1).  To get N-1 of these
you have to leave one wake out, it doesn't matter which one.  This transform
enforces the constraint that vortex lines can't end.

Solving the model consists of optimizing for these constraints:
For each polygon the normal velocity is zero.
Minimize the weighted sum of the squares of the vorticity shedding edges over
variation in the wake lines.

This produces a set of N(polygon) + N(wake vortex)-1 linear equations, which
can be solved directly.
*/

import (
	"fmt"
	"math"
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

type Wake struct {

	Points []Point
	TreePath []Point // path from model surface to single central Wake tie point

	BlurInternal float32
	BlurWake float32
	Strength float32
}

func (wake *Wake) Velocity(pt Point) Vector {

	if len(wake.Points) == 0 {
		return Vector{0, 0, 0}
	}

	var ret Vector

	pt1 := wake.Points[0]
	for i:=1; i<len(wake.Points); i++ {
		pt2 := wake.Points[i]
		ret = ret.Add(WakeSegmentVelocity(pt, pt1, pt2, wake.Strength, wake.BlurWake))
		pt1 = pt2
	}
	for i:=0; i<len(wake.TreePath); i++ {
		pt2 := wake.TreePath[i]
		ret = ret.Add(WakeSegmentVelocity(pt, pt1, pt2, wake.Strength, wake.BlurInternal))
		pt1 = pt2
	}
	return ret
}

func WakeSegmentVelocity(pt Point, pt1 Point, pt2 Point, strength float32, blur float32) Vector {

	center := pt1.Average(pt2)
	v := pt.Sub(center)
	ln := Sqrt(v.Dot(v))

	if ln < EPSILON {
		return Vector{0,0,0}
	}

	dl := pt2.Sub(pt1)

	if ln*ln < dl.Dot(dl)*(MIN_RANGE*MIN_RANGE/4) {
		ret := WakeSegmentVelocity(pt, pt1, center, strength, blur)
		ret = ret.Add(WakeSegmentVelocity(pt, center, pt2, strength, blur))
		return ret
	}

	scale := strength/(ln*(ln*ln + blur*blur))
	return dl.Cross(v).Scale(scale)
}

// FIXME redo
func (wake *Wake) Draw(glctx *DrawGL, color int, width float32) {

	if len(wake.Points) == 0 {
		return
	}

	pt := wake.Points[0]
	glctx.StartLine(pt)
	for i:=1; i<len(wake.Points); i++ {
		pt2 := wake.Points[i]
		glctx.LineTo(pt2);
		pt = pt2
	}
	for i:=0; i<len(wake.TreePath); i++ {
		pt2 := wake.TreePath[i]
		glctx.LineTo(pt2);
		pt = pt2
	}

	glctx.EndLine(Color{1,0,0,1})
}

type Edge struct {
	Center Point
	Normal Vector //length encodes strength
}

type Model struct {
	Panels []*Panel
	Wakes []*Wake
	Edges []*Edge // solutions are only unique if len(Edges) >= len(wakes)
}


func (model *Model) Velocity(pt Point, vStream Vector) Vector {
		v := vStream
		for _,p := range model.Panels {
			v = v.Add(p.Velocity(pt))
		}
		for _,w := range model.Wakes {
			v = v.Add(w.Velocity(pt))
		}
		return v
}

/*
Eqns to solve:

	Velocity(i, point) = sum(Panel[i]) + sum(Wake[i])

	Sum_i(Velocity(i, Center[j]) . Normal[j] ) = -V0.Normal[j]
	dCost/dW = sum_i_j ( (EdgeNormal[j] . Velocity(i, Center[j])) * (EdgeNormal[j] . d.Velocity(i, Center[j])/dW) ) = 0
*/

func DrawVector(glctx *DrawGL, pt Point, v Vector) {
	var scale float32 = 0.1
	pt2 := Point{ pt.X + v.X*scale, pt.Y + v.Y*scale, pt.Z + v.Z*scale }
	glctx.StartLine(pt)
	glctx.LineTo(pt2)
	glctx.EndLine(Color{0,0,1,1})
}

func DrawStreamLine(glctx *DrawGL,  model *Model, vStream Vector, pt Point) {

	glctx.StartLine(pt)
	for i:=0; i<500; i++ {

		pt2 := pt
		for j:=0; j<10; j++ {
			v := model.Velocity(pt2, vStream)
			v = v.Scale(0.002)
			pt2 = pt2.Add(v)
		}

		//ctx.Line(pt, pt2, 0x0000FF, 1)
		glctx.LineTo(pt2);
		pt = pt2
	}
	glctx.EndLine(Color{0,0,1,1})
}


func CreateModel() *Model {
	ret := Model{}

	AddTestFlat(&ret, 2, 0.4, 5)
	for _,p := range ret.Panels {
		p.InitStats()
	}

	return &ret
}

func main() {

	glctx, err := CreateDrawGL("aerodynamics/webgl/data.js")

	/*
	ctx, err := CreateDraw3D("out.svg",
		Vector{100, 0, 0},
		Vector{0, 100, 0},
		//Vector{0, 0, 0})
		Vector{22, 13, 0})
	ctx, err := CreateDraw3D("out.svg",
		Vector{0, 100, 0},
		Vector{0, 0, 0},
		Vector{100, 0, 0})
	*/
	/*
	ctx, err := CreateDraw3D("out.svg",
		Vector{0, 0, 0},
		Vector{0, 100, 0},
		Vector{100, 0, 0})
	*/

	if (err != nil) {
		fmt.Println(err)
		return
	}
	//defer ctx.Finalize()
	defer glctx.Finalize()

	model := CreateModel()

	var angle float32 = 20*3.1415926/180

	//vStream := Vector{0, 0, 0}
	vStream := Vector{-Cos(angle), -Sin(angle), 0}

	fmt.Printf("solving %v panels\n", len(model.Panels))
	Solve(model, vStream)
	fmt.Println("solving done")

	parPos := Point{0,0,0}.Add(vStream.Scale(-3))
	perpStep1 := vStream.Cross(Vector{0,0,0.1})
	perpStep2 := vStream.Cross(perpStep1)

	for i:=-4; i<=4; i++ {
		for j:=0; j<1; j++ {
			pt := parPos
			pt = pt.Add(perpStep1.Scale(float32(i)))
			pt = pt.Add(perpStep2.Scale(float32(j)*0.5))

			DrawStreamLine(glctx, model, vStream, pt)
		}
	}

	for _,w := range model.Wakes {
		w.Draw(glctx, 0xFF0000, 1)
	}

	for z:=-4.5; z<=4.5; z += 1 {
		pt := Point{1.0003, 0, float32(z/2)}
		v := model.Velocity(pt, vStream)
		DrawVector(glctx, pt, v)

		pt = Point{-1.0003, 0, float32(z/2)}
		v = model.Velocity(pt, vStream)
		DrawVector(glctx, pt, v)
	}

	/*
	for y := -20; y < 20; y++ {
		for x := -20; x < 20; x++ {
			pt := Point{float32(x)*0.1 + 0.05, float32(y)*0.1 + 0.05, 0}
			v := model.Velocity(pt, vStream)
			DrawVector(glctx, pt, v)
		}
	}
	*/

	for _,p := range model.Panels {

		v := model.Velocity(p.Center(), vStream)
		color := ColorFromValue(v.Dot(v)/vStream.Dot(vStream))
		if p.Count == 4 {
			glctx.DrawQuad(
				p.Points[0],
				p.Points[1],
				p.Points[2],
				p.Points[3],
				p.Normal,
				color)
		}
	}

	force := Vector{0,0,0}
	for _,p := range model.Panels {
		v := model.Velocity(p.Center(), vStream)
		cp := 1 - v.Dot(v)/vStream.Dot(vStream)
		cp = cp*p.Area
		df := p.Normal.Scale(-cp)
		force = force.Add(df)
	}

	fmt.Printf("force = %v\n", force)
	fmt.Printf("par = %v\n", vStream.Scale(force.Dot(vStream)))
	fmt.Printf("perp = %v\n", force.Sub(vStream.Scale(force.Dot(vStream))))
}


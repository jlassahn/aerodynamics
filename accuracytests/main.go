
package main

import (
	"fmt"
	"os"

	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/parser"
	"github.com/jlassahn/aerodynamics/solver"
	"github.com/jlassahn/aerodynamics/draw"
)

const (
	UNIT_DIAMETER = 1.128379167
)

var DragTests = []DragTest {
	{ "Sphere", BuildSphere, 0.2 },
	{ "Sphere2", BuildSphere2, 0.2 },
	{ "Plate", BuildPlate, 1.0 },
	{ "Streamline", BuildStreamline, 0.1 },
}


func main() {
	fmt.Println("HELLO")

	RunDragTests()
}

func RunDragTests() {

	for _,t := range DragTests {

		model := t.Builder()

		// FIXME put this somewhere
		for _,p := range model.Panels {
			p.InitStats()
		}

		vStream := Vector{0, -1, 0}
		solver.Solve(model, vStream)

		path := "./testresults/drag/"+t.Name
		os.MkdirAll(path, os.ModePerm)
		os.MkdirAll(path+"/webgl", os.ModePerm)

		fmt.Printf("%v: %v\n", t.Name, ComputeForce(model, vStream))

		glctx,_ := draw.CreateDrawGL(path+"/webgl/data.js")
		DrawPressureMap(glctx, model, vStream)
		glctx.Finalize()
	}
}

func ComputeForce(model *solver.Model, vStream Vector) Vector {

	force := Vector{0,0,0}

	for _,p := range model.Panels {

		// FIXME make pressure compute function
		v := model.Velocity(p.Center(), vStream)
		cp := 1 - v.Dot(v)/vStream.Dot(vStream)
		cp = LimitP(cp, p.Normal.Dot(vStream))

		df := p.Normal.Scale(cp*p.Area)
		force = force.Add(df)
	}

	return force
}

// FIXME better drawing primitives
func DrawPressureMap(glctx *draw.DrawGL, model *solver.Model, vStream Vector) {

	for _,p := range model.Panels {

		// pressure map
		// FIXME make pressure compute function
		v := model.Velocity(p.Center(), vStream)
		cp := 1 - v.Dot(v)/vStream.Dot(vStream)
		cp = LimitP(cp, p.Normal.Dot(vStream))
		color := draw.ColorFromValue(1 - cp)
		/*
		// grid
		color := draw.Color{0.4,0.4,0.4,1}
		if ((p.IX ^ p.IY) & 1) == 1 {
			color = draw.Color{.6,.6,.6,1}
		}
		*/

		if p.Count == 4 {
			glctx.DrawQuad(
				p.Points[0],
				p.Points[1],
				p.Points[2],
				p.Points[3],
				p.Normal,
				color)
		} else {
			glctx.DrawQuad(
				p.Points[0],
				p.Points[1],
				p.Points[2],
				p.Points[2],
				p.Normal,
				color)
		}
	}
}

type DragTest struct {
	Name string
	Builder func()*solver.Model
	ExpectedCD float32
}

func BuildSphere() *solver.Model {
	model := solver.Model{}

	el := parser.MakeFakeNose("Front", UNIT_DIAMETER, UNIT_DIAMETER/2)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  1,  0,  0,
		  0,  1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el = parser.MakeFakeNose("Back", UNIT_DIAMETER, UNIT_DIAMETER/2)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		 -1,  0,  0,
		  0, -1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	return &model
}

func BuildSphere2() *solver.Model {
	model := solver.Model{}

	el := parser.MakeFakeNose("Left", UNIT_DIAMETER, UNIT_DIAMETER/2)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  0,  1,  0,
		 -1,  0,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el = parser.MakeFakeNose("Right", UNIT_DIAMETER, UNIT_DIAMETER/2)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  0, -1,  0,
		  1,  0,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	return &model
}

func BuildPlate() *solver.Model {
	model := solver.Model{}

	el := parser.MakeFakeNose("Front", UNIT_DIAMETER, 0)
	el.Properties()["Style"] = 1
	*el.Position() = Vector{0, 0.05, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  1,  0,  0,
		  0,  1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el = parser.MakeFakeNose("Back", UNIT_DIAMETER, 0)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, -0.05, 0}
	*el.Rotate() = Matrix{ [9]float32{
		 -1,  0,  0,
		  0, -1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el2 := parser.MakeFakeTube("Edge", UNIT_DIAMETER, 0.1)
	el2.AddToModel(&model)

	return &model
}

func BuildStreamline() *solver.Model {
	model := solver.Model{}

	el := parser.MakeFakeNose("Front", UNIT_DIAMETER, UNIT_DIAMETER)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  1,  0,  0,
		  0,  1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el = parser.MakeFakeNose("Back", UNIT_DIAMETER, UNIT_DIAMETER)
	el.Properties()["Style"] = 2
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		 -1,  0,  0,
		  0, -1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	return &model
}

// FIXME put this elsewhere
func LimitP(p float32, dir float32) float32 {

	/*
	mx :=  -2*dir
	if p > mx {
		p = mx
	}
	*/

	if (dir > 0) && (p > 0) {
		p = 0
	}

	/*
	if p < -1 {
		p = -1
	}
	*/

	return p
}


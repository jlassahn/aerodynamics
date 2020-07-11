
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
	REYNOLDS = 10000
)

// FIXME need laminar nd turbulent drag models
var DragTests = []DragTest {
	{ "Sphere", BuildSphere, 0.2 },
	{ "Sphere2", BuildSphere2, 0.2 },
	{ "Plate", BuildPlate, 1.0 },
	{ "Streamline", BuildStreamline, 0.1 },
}


func main() {
	RunDragTests()
}

func RunDragTests() {

	for _,t := range DragTests {

		model := t.Builder()
		model.InitStats()

		vStream := Vector{0, -1, 0}
		solver.Solve(model, vStream)

		path := "./testresults/drag/"+t.Name
		os.MkdirAll(path, os.ModePerm)
		os.MkdirAll(path+"/webgl", os.ModePerm)

		fmt.Printf("%v: %v\n", t.Name, ComputeForce(model, vStream))

		glctx,_ := draw.CreateDrawGLDirectory(path+"/webgl")
		draw.DrawPressureMap(glctx, model, vStream, REYNOLDS)
		glctx.Finalize()
	}
}

func ComputeForce(model *solver.Model, vStream Vector) Vector {

	force := Vector{0,0,0}

	for _,p := range model.Panels {
		cp := solver.PressureCoefficient(model, p, vStream, REYNOLDS)
		df := p.Normal.Scale(-cp*p.Area)
		force = force.Add(df)
	}

	return force
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
	el.Properties()["Style"] = 1
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

	el := parser.MakeFakeNose("Front", UNIT_DIAMETER, UNIT_DIAMETER*1)
	el.Properties()["Style"] = 0
	*el.Position() = Vector{0, 0, 0}
	*el.Rotate() = Matrix{ [9]float32{
		  1,  0,  0,
		  0,  1,  0,
		  0,  0,  1,
	}}
	el.AddToModel(&model)

	el = parser.MakeFakeNose("Back", UNIT_DIAMETER, UNIT_DIAMETER*3)
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


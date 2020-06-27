
package main

import (
	"fmt"

	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/parser"
	"github.com/jlassahn/aerodynamics/solver"
	"github.com/jlassahn/aerodynamics/draw"
)


func CreateModel() *solver.Model {
	ret := solver.Model{}

	//solver.AddTestFlat(&ret, 2, 0.2, 5)

	tube := parser.MakeFakeTube(0.5, 2)
	*tube.Position() = Vector{0, 0, 0}
	tube.AddToModel(&ret)

	nose := parser.MakeFakeNose(0.5, 1)
	*nose.Position() = Vector{0, 1, 0}
	nose.AddToModel(&ret)

	tail := parser.MakeFakeNose(0.5, -0.1)
	*tail.Position() = Vector{0, -1, 0}
	tail.AddToModel(&ret)

	for _,p := range ret.Panels {
		p.InitStats()
	}

	return &ret
}

func main() {

	model := CreateModel()

	var angle float32 = 90
	vStream := Vector{0, 0, 0}

	rads := angle*3.1415926/180

	vStream = Vector{-Cos(rads), -Sin(rads), 0}

	fmt.Printf("solving %v panels\n", len(model.Panels))
	solver.Solve(model, vStream)
	fmt.Println("solving done")

	ComputeForces(model, vStream)
	glctx, err := draw.CreateDrawGL("aerodynamics/webgl/data.js")

	if (err != nil) {
		fmt.Println(err)
		return
	}
	defer glctx.Finalize()

	parPos := Point{0,0,0}.Add(vStream.Scale(-3))
	perpStep1 := vStream.Cross(Vector{0,0,0.1})
	perpStep2 := vStream.Cross(perpStep1)

	for i:=-4; i<=4; i++ {
		for j:=0; j<1; j++ {
			pt := parPos
			pt = pt.Add(perpStep1.Scale(float32(i)))
			pt = pt.Add(perpStep2.Scale(float32(j)*0.5))

			draw.DrawStreamLine(glctx, model, vStream, pt)
		}
	}

	for _,w := range model.Wakes {
		draw.DrawWake(glctx, w, 0xFF0000, 1)
	}

	/*
	for z:=-4.5; z<=4.5; z += 1 {
		pt := Point{1.0003, 0, float32(z/2)}
		v := model.Velocity(pt, vStream)
		draw.DrawVector(glctx, pt, v)

		pt = Point{-1.0003, 0, float32(z/2)}
		v = model.Velocity(pt, vStream)
		draw.DrawVector(glctx, pt, v)
	}
	*/

	/*
	for y := -20; y < 20; y++ {
		for x := -20; x < 20; x++ {
			pt := Point{float32(x)*0.1 + 0.05, float32(y)*0.1 + 0.05, 0}
			v := model.Velocity(pt, vStream)
			draw.DrawVector(glctx, pt, v)
		}
	}
	*/

	for _,p := range model.Panels {

		v := model.Velocity(p.Center(), vStream)
		color := draw.ColorFromValue(v.Dot(v)/vStream.Dot(vStream))
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

	//parser.ParseTest()
}

func ComputeForces(model *solver.Model, vStream Vector) {

	var torque float32
	force := Vector{0,0,0}
	for _,p := range model.Panels {
		v := model.Velocity(p.Center(), vStream)
		cp := 1 - v.Dot(v)/vStream.Dot(vStream)
		cp = cp*p.Area
		df := p.Normal.Scale(-cp)
		force = force.Add(df)
		torque += (p.Center().X - 1)*df.Y
	}

	fPar := vStream.Scale(force.Dot(vStream))
	fPerp := force.Sub(fPar)

	lift := Sqrt(fPerp.Dot(fPerp))
	drag := Sqrt(fPar.Dot(fPar))

	fmt.Printf("lift = %.3f drag = %.3f\n", lift, drag)
	fmt.Printf("force = %v\n", force)
	fmt.Printf("torque = %v\n", torque)
	fmt.Printf("par = %v\n", fPar)
	fmt.Printf("perp = %v\n", fPerp)
}

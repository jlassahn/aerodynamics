
package main

import (
	"fmt"

	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/parser"
	"github.com/jlassahn/aerodynamics/solver"
	"github.com/jlassahn/aerodynamics/draw"
)

const (
	REYNOLDS = 10000
)


func CreateModel() *solver.Model {
	ret := solver.Model{}

	//solver.AddTestFlat(&ret, 2, 0.2, 5)

	tube := parser.MakeFakeTube("tube", 0.5, 3)
	*tube.Position() = Vector{0, 0, 0}
	tube.AddToModel(&ret)

	nose := parser.MakeFakeNose("nose", 0.5, 1)
	nose.Properties()["Style"] = 0
	*nose.Position() = Vector{0, 1.5, 0}
	nose.AddToModel(&ret)

	tail := parser.MakeFakeTail("tail", 0.5, 0)
	tail.Properties()["Style"] = 1
	*tail.Position() = Vector{0, -1.5, 0}
	tail.AddToModel(&ret)

	f1 := parser.MakeFakeSheet("fin1", 1, 1)
	*f1.Position() = Vector{0.25, -1, 0}
	f1.AddToModel(&ret)

	f1 = parser.MakeFakeSheet("fin2", 1, 1)
	*f1.Position() = Vector{0, -1, 0.25}
	*f1.Rotate() = Matrix{ [9]float32{
		  0,  0, -1,
		  0,  1,  0,
		  1,  0,  0,
	}}
	f1.AddToModel(&ret)

	f1 = parser.MakeFakeSheet("fin3", 1, 1)
	*f1.Position() = Vector{0, -1, -0.25}
	*f1.Rotate() = Matrix{ [9]float32{
		  0,  0,  1,
		  0,  1,  0,
		 -1,  0,  0,
	}}
	f1.AddToModel(&ret)

	f1 = parser.MakeFakeSheet("fin4", 1, 1)
	*f1.Position() = Vector{-0.25, -1, 0}
	*f1.Rotate() = Matrix{ [9]float32{
		 -1,  0,  0,
		  0,  1,  0,
		  0,  0, -1,
	}}
	f1.AddToModel(&ret)

	ret.InitStats()

	return &ret
}

func main() {

	//model := CreateModel()
	model := parser.ParseTest()

	var angle float32 = 90
	vStream := Vector{0, 0, 0}

	rads := angle*3.1415926/180

	vStream = Vector{-Cos(rads), -Sin(rads), 0}

	fmt.Printf("solving %v panels\n", len(model.Panels))
	solver.Solve(model, vStream)
	fmt.Println("solving done")

	ComputeForces(model, vStream)
	glctx,err := draw.CreateDrawGLDirectory("aerodynamics/webgl")
	if (err != nil) {
		fmt.Println(err)
		return
	}
	defer glctx.Finalize()

	parPos := Point{0,0,0.01}.Add(vStream.Scale(-3))
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

	draw.DrawPressureMap(glctx, model, vStream, REYNOLDS)

}

func ComputeForces(model *solver.Model, vStream Vector) {

	torque := Vector{0,0,0}
	force := Vector{0,0,0}
	forceSet := map[string]Vector{}
	cg := Point{0,0,0} // FIXME fake

	strength := float32(0)
	for _,p := range model.Panels {

		cp := solver.PressureCoefficient(model, p, vStream, REYNOLDS)
		df := p.Normal.Scale(-cp*p.Area)

		force = force.Add(df)
		torque = torque.Add(p.Center().Sub(cg).Cross(df))
		strength += p.Strength*p.Area

		forceSet[p.Tag] = forceSet[p.Tag].Add(df)
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
	fmt.Printf("strength = %v\n", strength)
	fmt.Println(forceSet)

	forcePerp := force.Sub(torque.Scale(force.Dot(torque)*1/(torque.Dot(torque))))
	fmt.Printf("residual force = %v\n", force.Sub(forcePerp))

	fmt.Printf("Force Perp = %v\n", forcePerp)

	cp := forcePerp.Cross(torque).Scale(1/forcePerp.Dot(forcePerp))
	fmt.Printf("Cp offset = %v\n", cp)
	fmt.Printf("Cp = %v\n", cg.Add(cp))
	fmt.Printf("Cd = %v\n", force.Dot(vStream)/(3.14159*0.25*0.25))
}


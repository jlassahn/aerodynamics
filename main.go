
package main

import (
	"fmt"
)


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

	if (err != nil) {
		fmt.Println(err)
		return
	}
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


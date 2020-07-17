
package draw

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

func DrawVector(glctx *DrawGL, pt Point, v Vector) {
	var scale float32 = 0.1
	pt2 := Point{ pt.X + v.X*scale, pt.Y + v.Y*scale, pt.Z + v.Z*scale }
	glctx.DrawLine(pt, pt2, Color{0,0,1,1})
}

func DrawStreamLine(glctx *DrawGL,  model *solver.Model, vStream Vector, pt Point) {


	for i:=0; i<500; i++ {

		pt2 := pt
		for j:=0; j<10; j++ {
			v := model.Velocity(pt2, vStream)
			v = v.Scale(0.002)
			pt2 = pt2.Add(v)
		}

		glctx.DrawLine(pt, pt2, Color{0,0,1,1})
		pt = pt2
	}
}

// FIXME redo
func DrawWake(glctx *DrawGL, wake *solver.Wake, color int, width float32) {

	if len(wake.Points) == 0 {
		return
	}

	r := wake.Strength*200
	if r < 0 { r = -r }
	if r > 1 { r = 1 }

	pt := wake.Points[0]
	for i:=1; i<len(wake.Points); i++ {
		pt2 := wake.Points[i]
		glctx.DrawLine(pt, pt2, Color{1,0,0,r})
		pt = pt2
	}
	for i:=0; i<len(wake.TreePath); i++ {
		pt2 := wake.TreePath[i]
		glctx.DrawLine(pt, pt2, Color{1,0,0,r})
		pt = pt2
	}
}

func DrawPressureMap(glctx *DrawGL, model *solver.Model, vStream Vector, reynolds float32) {

	glctx.StartObject("Main", []string{"Default"})
	defer glctx.EndObject()

	for _,p := range model.Panels {

		// pressure map
		cp := solver.PressureCoefficient(model, p, vStream, reynolds)
		color := ColorFromValue(1 - cp)
		/*
		// grid
		color := draw.Color{0.4,0.4,0.4,1}
		if ((p.IX ^ p.IY) & 1) == 1 {
			color = draw.Color{.6,.6,.6,1}
		}
		*/

		glctx.DrawTriangle(
			p.Points[0],
			p.Points[1],
			p.Points[2],
			p.Normal,
			[]Color{color})
		if p.Count == 4 {
			glctx.DrawTriangle(
				p.Points[0],
				p.Points[2],
				p.Points[3],
				p.Normal,
				[]Color{color})
		}
	}
}


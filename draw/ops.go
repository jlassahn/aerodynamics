
package draw

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

func DrawVector(glctx *DrawGL, pt Point, v Vector) {
	var scale float32 = 0.1
	pt2 := Point{ pt.X + v.X*scale, pt.Y + v.Y*scale, pt.Z + v.Z*scale }
	glctx.StartLine(pt)
	glctx.LineTo(pt2)
	glctx.EndLine(Color{0,0,1,1})
}

func DrawStreamLine(glctx *DrawGL,  model *solver.Model, vStream Vector, pt Point) {

	glctx.StartLine(pt)
	for i:=0; i<500; i++ {

		pt2 := pt
		for j:=0; j<10; j++ {
			v := model.Velocity(pt2, vStream)
			v = v.Scale(0.002)
			pt2 = pt2.Add(v)
		}

		glctx.LineTo(pt2);
		pt = pt2
	}
	glctx.EndLine(Color{0,0,1,1})
}

// FIXME redo
func DrawWake(glctx *DrawGL, wake *solver.Wake, color int, width float32) {

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

	r := wake.Strength*200
	if r < 0 { r = -r }
	if r > 1 { r = 1 }

	glctx.EndLine(Color{1,0,0,r})
}



package main

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

		glctx.LineTo(pt2);
		pt = pt2
	}
	glctx.EndLine(Color{0,0,1,1})
}


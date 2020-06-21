
package main

type Wake struct {

	Points []Point
	TreePath []Point // path from model surface to single central Wake tie point

	BlurInternal float32
	BlurWake float32
	Strength float32
}

type Edge struct {
	Center Point
	Normal Vector //length encodes strength
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


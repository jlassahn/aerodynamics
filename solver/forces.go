
package solver

import (
	. "github.com/jlassahn/aerodynamics/geometry"
)

func PressureCoefficient(model *Model, panel *Panel, vStream Vector, reynolds float32) float32 {

	v := model.Velocity(panel.Center(), vStream)
	cp := 1 - v.Dot(v)/vStream.Dot(vStream)
	cp = LimitP(cp, panel.Normal.Dot(vStream), reynolds)
	return cp
}

func LimitP(p float32, dir float32, reynolds float32) float32 {

	// FIXME probably slightly underestimates drag at high Reynolds numbers
	if (dir > 0) && (p > 0) {
		p = 0
	}

	/*
	// FIXME probably slightly overestimates form drag at medium reynolds numbers
	mx :=  -2*dir
	if mx > 1 { mx = 1 }
	if mx < -0.5 { mx = - 0.5 }

	if p > mx {
		p = mx
	}
	*/

	return p
}


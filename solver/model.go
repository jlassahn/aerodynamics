
package solver

import (
	. "github.com/jlassahn/aerodynamics/geometry"
)

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



package main

func Solve(model *Model, vStream Vector) {

	nP := len(model.Panels)
	//nE := len(model.Edges)
	//nW := len(model.Wakes)
	n := nP

	vCoupling := make([]float32, n*n)
	kStream := make([]float32, n)

	for i:=0; i<nP; i++ {
		model.Panels[i].Strength = 1
	}

	for j:=0; j<nP; j++ {
		for i:=0; i<nP; i++ {
			if i==j {
				vCoupling[i + n*j] = 1
			} else {
				src := model.Panels[i]
				dst := model.Panels[j]
				v := src.Velocity(dst.Center())
				vCoupling[i + n*j] = v.Dot(dst.Normal)
			}
		}

		kStream[j] = -vStream.Dot(model.Panels[j].Normal)
	}


	// solve equations
	for i:=0; i<n; i++ {

		pivot := vCoupling[i + n*i]
		if pivot < 0.01 && pivot > -0.01 {
			panic("solver fail") // FIXME handle solver errors
		}

		scale := 1/pivot
		for k:=i; k<n; k++ {
			vCoupling[k + n*i] *= scale
		}
		kStream[i] *= scale

		for j:=0; j<n; j++ {
			if i == j {
				continue
			}

			scale = vCoupling[i + n*j]
			for k:=i; k<n; k++ {
				vCoupling[k + n*j] -= scale*vCoupling[k + n*i]
			}
			kStream[j] -= scale*kStream[i]
		}
	}

	for i:=0; i<n; i++ {
		model.Panels[i].Strength = kStream[i]
	}
}


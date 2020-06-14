
package main

import (
	"fmt"
)

func Solve(model *Model, vStream Vector) {

	nP := len(model.Panels)
	nE := len(model.Edges)
	nW := len(model.Wakes)
	n := nP + nW

	vCoupling := make([]float32, n*n)
	kStream := make([]float32, n)

	for i:=0; i<nP; i++ {
		model.Panels[i].Strength = 1
	}

	for i :=0; i<nW; i++ {
		model.Wakes[i].Strength = 1
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

		for i:=0; i<nW; i++ {
			src := model.Wakes[i]
			dst := model.Panels[j]
			v := src.Velocity(dst.Center())
			vCoupling[i+nP + n*j] = v.Dot(dst.Normal)
		}

		kStream[j] = -vStream.Dot(model.Panels[j].Normal)
	}

	for j:=0; j<nW-1; j++ {
		for k:=0; k<nE; k++ {

			dst := model.Edges[k]
			pe := dst.Center
			wakeV := model.Wakes[j].Velocity(pe)
			// FIXME can be hoisted out of the loop...
			for i:=0; i<nW; i++ {
				src := model.Wakes[i]
				wakeV.Add(src.Velocity(pe).Scale(-1.0/float32(nW)))
			}
			edgeWeight := dst.Normal.Dot(wakeV)

			for i:=0; i<nP; i++ {
				src := model.Panels[i]
				v := src.Velocity(pe)
				vCoupling[i + n*(j+nP)] += v.Dot(dst.Normal)*edgeWeight
			}

			for i:=0; i<nW; i++ {
				src := model.Wakes[i]
				v := src.Velocity(pe)
				vCoupling[(i+nP) + n*(j+nP)] += v.Dot(dst.Normal)*edgeWeight
			}

			kStream[j+nP] -= vStream.Dot(dst.Normal)*edgeWeight
		}
	}

	for i:=0; i<nW; i++ {
		vCoupling[i+nP + n*(n-1)] = 1
	}

	count := 0
	for i:=0; i<n*n; i++ {
		if (vCoupling[i] < 0.0001) && (vCoupling[i] > -0.0001) {
			count++
		}
	}
	fmt.Printf("zeros = %v / %v\n", count, n*n)

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

	for i:=0; i<nP; i++ {
		model.Panels[i].Strength = kStream[i]
	}

	for i:=0; i<nW; i++ {
		model.Wakes[i].Strength = kStream[i+nP]
	}
}


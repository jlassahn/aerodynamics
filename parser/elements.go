
package parser

import (
	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type Link interface {
	Parent() Element
	Pair() Link
	Properties() map[string]float32
	Offset() *Vector
	Rotate() *Matrix
}

type Element interface {
	Links() []Link
	Properties() map[string]float32

	Position() *Vector
	Rotate() *Matrix

	AddToModel(model *solver.Model)
}


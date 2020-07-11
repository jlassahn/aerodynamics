
package parser

import (
	"fmt"

	. "github.com/jlassahn/aerodynamics/geometry"
	"github.com/jlassahn/aerodynamics/solver"
)

type genState struct {
	Rotate Matrix
	Offset Vector
}

func GenerateModel(root *ParseObject) *solver.Model {

	model := &solver.Model{}

	state := genState{}
	state.Rotate = IdentityMatrix

	AddTreeToModel(model, &state,  root)

	return model
}

func AddTreeToModel(model *solver.Model, stateIn *genState, obj *ParseObject) {

	state := *stateIn

	localOffset := Vector{0,0,0}
	localRotate := IdentityMatrix

	baseOffset := Vector{
		getNumber(obj, "RadialPosition"),
		getNumber(obj, "Position"),
		0}

	baseRotate := RollMatrix(getNumber(obj, "Rotation"))

	/*
		Position: -90
		Rotation: i*120
		RadialPosition: 12

		XOffset: 0
		YOffset: 0
		ZOffset: 0
		Roll: 0
		Pitch: 0
		Yaw: 0
	*/

	// transform pipeline
	// Vout = Vroot + Mroot*(Vmid + Mmid*Vtip)
	// Vout = Vroot + Mroot*Vmid + Mroot*Mmid*Vtip
	
	state.Rotate = state.Rotate.Mult(baseRotate)
	state.Offset = state.Offset.Add(state.Rotate.Transform(baseOffset))
	state.Offset = state.Offset.Add(state.Rotate.Transform(localOffset))
	state.Rotate = state.Rotate.Mult(localRotate)

	fmt.Printf("Name = %v Index=%v Offset=%v\n", obj.Name, obj.Index, state.Offset)
	switch obj.ObjectType {
	case "Tube":
		radius := getNumber(obj, "Radius")
		length := getNumber(obj, "Length")
		tube := MakeFakeTube(obj.Name, radius*2, length)
		*tube.Position() = state.Offset
		*tube.Rotate() = state.Rotate
		tube.AddToModel(model)

	case "Sheet":
		root := getNumber(obj, "RootChord")
		span := getNumber(obj, "Span")
		fin := MakeFakeSheet(obj.Name, span, root)
		fin.Properties()["Sweep"] = getNumber(obj, "Sweep")
		fin.Properties()["Taper"] = (root - getNumber(obj, "TipChord"))/2
		fin.Properties()["Thick"] = getNumber(obj, "Thickness")
		*fin.Position() = state.Offset
		*fin.Rotate() = state.Rotate
		fin.AddToModel(model)

	case "NoseCone":
		radius := getNumber(obj, "Radius")
		length := getNumber(obj, "Length")
		style := getNumber(obj, "Shape")
		nose := MakeFakeNose(obj.Name, radius*2, length)
		nose.Properties()["Style"] = style
		*nose.Position() = state.Offset
		*nose.Rotate() = state.Rotate
		nose.AddToModel(model)

	case "TailCone":
		radius := getNumber(obj, "Radius")
		length := getNumber(obj, "Length")
		style := getNumber(obj, "Shape")
		tail := MakeFakeTail(obj.Name, radius*2, length)
		tail.Properties()["Style"] = style
		*tail.Position() = state.Offset
		*tail.Rotate() = state.Rotate.Mult(*tail.Rotate())
		tail.AddToModel(model)

	default:
		fmt.Printf("FIXME skipping %v\n", obj.ObjectType)
	}

	for _,childArray := range obj.Definitions {
		for i := range childArray.Children {
			AddTreeToModel(model, &state, &childArray.Children[i])
		}
	}
}

func RollMatrix(angle float32) Matrix {
	a := angle*3.1415926/180

	return Matrix { [9]float32 {
		Cos(a),   0, -Sin(a),
		     0,   1,       0,
		Sin(a),   0,  Cos(a) } }
}

func getNumber(obj *ParseObject, name string) float32 {
	v := obj.Values[name]
	if v == nil {
		return 0
	}
	return v.Number()
}



package geometry

import (
	"math"
)

type Point struct {
	X float32
	Y float32
	Z float32
}

type Vector Point

func (a Point) Sub(b Point) Vector {
	return Vector{
		X: a.X - b.X,
		Y: a.Y - b.Y,
		Z: a.Z - b.Z,
	}
}

func (a Point) Add(b Vector) Point {
	return Point{
		X: a.X + b.X,
		Y: a.Y + b.Y,
		Z: a.Z + b.Z,
	}
}

func (a Vector) Add(b Vector) Vector {
	return Vector{
		X: a.X + b.X,
		Y: a.Y + b.Y,
		Z: a.Z + b.Z,
	}
}

func (a Vector) Sub(b Vector) Vector {
	return Vector{
		X: a.X - b.X,
		Y: a.Y - b.Y,
		Z: a.Z - b.Z,
	}
}

func (a Point) Average(b Point) Point {
	return Point{
		X: (a.X + b.X)/2,
		Y: (a.Y + b.Y)/2,
		Z: (a.Z + b.Z)/2,
	}
}

func (a Vector) Average(b Vector) Vector {
	return Vector{
		X: (a.X + b.X)/2,
		Y: (a.Y + b.Y)/2,
		Z: (a.Z + b.Z)/2,
	}
}

func (a Vector) Scale(n float32) Vector {
	return Vector{
		X: a.X*n,
		Y: a.Y*n,
		Z: a.Z*n,
	}
}

func (a Vector) Dot(b Vector) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vector) Cross(b Vector) Vector {
	return Vector{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

type Matrix struct {
	M [9]float32
}

func (a Matrix) Mult(b Matrix) Matrix {

	ret := Matrix{}
	for i:=0; i<3; i++ {
		for j:=0; j<3; j++ {
			for k:=0; k<3; k++ {
				ret.M[i+3*j] += a.M[k+3*j]*b.M[i+3*k]
			}
		}
	}
	return ret
}

func (a Matrix) Transform(b Vector) Vector {

	ret := Vector{}

	ret.X = a.M[0]*b.X + a.M[1]*b.Y + a.M[2]*b.Z
	ret.Y = a.M[3]*b.X + a.M[4]*b.Y + a.M[5]*b.Z
	ret.Z = a.M[6]*b.X + a.M[7]*b.Y + a.M[8]*b.Z

	return ret
}

func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

var IdentityMatrix = Matrix { [9]float32 {
	1, 0, 0,
	0, 1, 0,
	0, 0, 1 } }


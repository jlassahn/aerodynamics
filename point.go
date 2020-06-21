
package main

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

func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}


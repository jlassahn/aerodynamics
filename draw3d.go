
package main

// FIXME change SVG interface to use float coords

type Draw3D struct {
	Dx Vector
	Dy Vector
	Dz Vector
	SVG SVGFile
}

func CreateDraw3D(filename string, dx Vector, dy Vector, dz Vector) (*Draw3D, error) {

	svg, err := CreateSVGFile(filename, 1000, 1000)
	if (err != nil) {
		return nil, err
	}


	ret := Draw3D{}
	ret.Dx = dx
	ret.Dy = dy
	ret.Dz = dz
	ret.SVG = svg

	return &ret, nil
}

func (ctx *Draw3D) Line(v0 Point, v1 Point, color int, width float32) {
	x0 := 500 + v0.X*ctx.Dx.X + v0.Y*ctx.Dy.X + v0.Z*ctx.Dz.X
	y0 := 500 - v0.X*ctx.Dx.Y - v0.Y*ctx.Dy.Y - v0.Z*ctx.Dz.Y
	x1 := 500 + v1.X*ctx.Dx.X + v1.Y*ctx.Dy.X + v1.Z*ctx.Dz.X
	y1 := 500 - v1.X*ctx.Dx.Y - v1.Y*ctx.Dy.Y - v1.Z*ctx.Dz.Y
	ctx.SVG.Line(x0, y0, x1, y1, color, width)
}

func (ctx *Draw3D) Finalize() {
	ctx.SVG.Finalize()
}


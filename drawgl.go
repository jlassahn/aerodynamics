
package main

import (
	"fmt"
	"io"
	"os"
)

type Color struct {
	R, G, B, A float32
}

type DrawGL struct {
	fp io.WriteCloser
	quads []quad
	lines []line
}

func ColorFromValue(x float32) Color {
	if x < 0 {
		return Color{1,0,0,1}
	}

	if x < 1 {
		return Color{1-x,x,0,1}
	}

	if x < 2 {
		return Color{0, 2-x, x-1, 1}
	}
	if x < 2.5 {
		return Color{x-2, 0, 3-x, 1}
	}
	return Color{.5,0,.5,1}
}

func CreateDrawGL(name string) (*DrawGL, error) {

	ret := DrawGL{}

	fp, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	ret.fp = fp

	return &ret, nil
}

func (ctx *DrawGL) DrawQuad(p1, p2, p3, p4 Point, n Vector, color Color) {

	q := quad{p1, p2, p3, p4, n, color}
	ctx.quads = append(ctx.quads, q)
}

func (ctx *DrawGL) StartLine(p Point) {

	ln := line{}
	ln.points = []Point{p}
	ctx.lines = append(ctx.lines, ln)
}

func (ctx *DrawGL) LineTo(p Point) {
	ln := &ctx.lines[len(ctx.lines)-1]
	ln.points = append(ln.points, p)
}

func (ctx DrawGL) EndLine(c Color) {
	ctx.lines[len(ctx.lines)-1].color = c
}

func (ctx *DrawGL) Finalize() {

	fmt.Fprintf(ctx.fp, "DATA_quads = new Float32Array([\n")
	for _,q := range ctx.quads {
		fmt.Fprintf(ctx.fp, 
			"\t %6.3f, %6.3f, %6.3f,   %6.3f, %6.3f, %6.3f,   %6.3f, %6.3f, %6.3f,   %6.3f, %6.3f, %6.3f,\n",
			q.p1.X, q.p1.Y, q.p1.Z,
			q.p2.X, q.p2.Y, q.p2.Z,
			q.p3.X, q.p3.Y, q.p3.Z,
			q.p4.X, q.p4.Y, q.p4.Z)
	}
	fmt.Fprintf(ctx.fp, "]);\n\n")

	fmt.Fprintf(ctx.fp, "DATA_quadcolors = new Float32Array([\n")
	for _,q := range ctx.quads {
		fmt.Fprintf(ctx.fp, 
			"\t %6.3f, %6.3f, %6.3f, %6.3f,\n",
			q.color.R, q.color.G, q.color.B, q.color.A)
	}
	fmt.Fprintf(ctx.fp, "]);\n\n")

	fmt.Fprintf(ctx.fp, "DATA_quadnorms = new Float32Array([\n")
	for _,q := range ctx.quads {
		fmt.Fprintf(ctx.fp, 
			"\t %6.3f, %6.3f, %6.3f,\n",
			q.normal.X, q.normal.Y, q.normal.Z)
	}
	fmt.Fprintf(ctx.fp, "]);\n\n")

	fmt.Fprintf(ctx.fp, "DATA_lines = [\n")
	for _,ln := range ctx.lines {
		fmt.Fprintf(ctx.fp, "\tnew Float32Array([\n")
		for _,p := range ln.points {
			fmt.Fprintf(ctx.fp,
				"\t\t %6.3f, %6.3f, %6.3f,\n",
				p.X, p.Y, p.Z)
		}
		fmt.Fprintf(ctx.fp, "]),\n")
	}
	fmt.Fprintf(ctx.fp, "]\n\n")

	fmt.Fprintf(ctx.fp, "DATA_linecolors = new Float32Array([\n")
	for _,ln := range ctx.lines {
		fmt.Fprintf(ctx.fp, 
			"\t %6.3f, %6.3f, %6.3f, %6.3f,\n",
			ln.color.R, ln.color.G, ln.color.B, ln.color.A)
	}
	fmt.Fprintf(ctx.fp, "])\n\n")

	ctx.fp.Close()
	ctx.fp = nil
}

type quad struct {
	p1, p2, p3, p4 Point
	normal Vector
	color Color
}

type line struct {
	points []Point
	color Color
}


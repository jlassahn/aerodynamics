
package draw

import (
	"fmt"
	"io"
	"os"
	"path"

	. "github.com/jlassahn/aerodynamics/geometry"
)

type Color struct {
	R, G, B, A float32
}

type DrawGL struct {
	fp io.WriteCloser

	objects []object3d
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

func CreateDrawGLDirectory(base string) (*DrawGL, error) {

	fp,_ := os.Create(path.Join(base, "graph3d.js"))
	fp.Write([]byte(Graph3Djs))
	fp.Close()

	fp,_ = os.Create(path.Join(base, "style.css"))
	fp.Write([]byte(Stylecss))
	fp.Close()

	fp,_ = os.Create(path.Join(base, "test.html"))
	fp.Write([]byte(MainHTML))
	fp.Close()

	name := path.Join(base, "data.js")
	return CreateDrawGL(name)

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

func (ctx *DrawGL) StartObject(name string, colors []string) {

	ctx.objects = append(ctx.objects, object3d{})
	obj := &ctx.objects[len(ctx.objects)-1]
	obj.Name = name
	obj.TriangleColors = make([][][3]Color, len(colors))
	obj.ColorNames = colors
}

func (ctx *DrawGL) EndObject() {
}

func (ctx *DrawGL) DrawTriangle(p1, p2, p3 Point, n Vector, colors []Color) {

	obj := &ctx.objects[len(ctx.objects)-1]

	obj.TriangleCoords = append(obj.TriangleCoords, [3]Point{p1, p2, p3})
	obj.TriangleNorms = append(obj.TriangleNorms,  [3]Vector{n, n, n})
	for i, col := range colors {
		obj.TriangleColors[i] = append(obj.TriangleColors[i],
			[3]Color{col, col, col})
	}
}

func (ctx *DrawGL) DrawLine(p1 Point, p2 Point, color Color) {

	obj := &ctx.objects[len(ctx.objects)-1]

	obj.LineCoords = append(obj.LineCoords, [2]Point{p1, p2})
	obj.LineColors = append(obj.LineColors,  [2]Color{color, color})
}

func (ctx *DrawGL) Finalize() {

	fmt.Fprintf(ctx.fp, "DATA_Objects = [\n")
	for _,obj := range ctx.objects {

		fmt.Fprintf(ctx.fp, "{\n")
		fmt.Fprintf(ctx.fp, "\tName: \"%s\",\n", obj.Name)

		fmt.Fprintf(ctx.fp, "\tTriangleCoords: new Float32Array([\n")
		for _,pts := range obj.TriangleCoords {
			fmt.Fprintf(ctx.fp, "\t\t %6.3f, %6.3f, %6.3f,    %6.3f, %6.3f, %6.3f,    %6.3f, %6.3f, %6.3f,\n",
				pts[0].X, pts[0].Y, pts[0].Z,
				pts[1].X, pts[1].Y, pts[1].Z,
				pts[2].X, pts[2].Y, pts[2].Z)

		}
		fmt.Fprintf(ctx.fp, "\t]),\n")

		fmt.Fprintf(ctx.fp, "\tTriangleNorms: new Float32Array([\n")
		for _,pts := range obj.TriangleNorms {
			fmt.Fprintf(ctx.fp, "\t\t %6.3f, %6.3f, %6.3f,    %6.3f, %6.3f, %6.3f,    %6.3f, %6.3f, %6.3f,\n",
				pts[0].X, pts[0].Y, pts[0].Z,
				pts[1].X, pts[1].Y, pts[1].Z,
				pts[2].X, pts[2].Y, pts[2].Z)

		}
		fmt.Fprintf(ctx.fp, "\t]),\n")

		fmt.Fprintf(ctx.fp, "\tTriangleColors: {\n")
		for i,cols := range obj.TriangleColors {
			fmt.Fprintf(ctx.fp, "\t\t\"%s\": new Float32Array([\n", obj.ColorNames[i])
			for _,col := range cols {
				fmt.Fprintf(ctx.fp, "\t\t\t %1.3f, %1.3f, %1.3f, %1.3f,    %1.3f, %1.3f, %1.3f, %1.3f,    %1.3f, %1.3f, %1.3f, %1.3f,\n",
					col[0].R, col[0].G, col[0].B, col[0].A,
					col[1].R, col[1].G, col[1].B, col[1].A,
					col[2].R, col[2].G, col[2].B, col[2].A)
			}
			fmt.Fprintf(ctx.fp, "\t\t]),\n")
		}
		fmt.Fprintf(ctx.fp, "\t},\n")

		fmt.Fprintf(ctx.fp, "\tLineCoords: new Float32Array([\n")
		for _,pts := range obj.LineCoords {
			fmt.Fprintf(ctx.fp, "\t\t %6.3f, %6.3f, %6.3f,    %6.3f, %6.3f, %6.3f,\n",
				pts[0].X, pts[0].Y, pts[0].Z,
				pts[1].X, pts[1].Y, pts[1].Z)

		}
		fmt.Fprintf(ctx.fp, "\t]),\n")

		fmt.Fprintf(ctx.fp, "\tLineColors: new Float32Array([\n")
		for _,col := range obj.LineColors {
			fmt.Fprintf(ctx.fp, "\t\t %1.3f, %1.3f, %1.3f, %1.3f,    %1.3f, %1.3f, %1.3f, %1.3f,\n",
				col[0].R, col[0].G, col[0].B, col[0].A,
				col[1].R, col[1].G, col[1].B, col[1].A)
		}
		fmt.Fprintf(ctx.fp, "\t]),\n")

		/*
		LineCoords [][2]Point
		LineColors [][2]Color

		XRayTriangleCoords [][3]Point
		XRayTriangleNorms [][3]Point
		XRayTriangleColors [][3]Color
		XRayLineCoords [][2]Point
		XRayLineColors [][2]Color
		*/
		fmt.Fprintf(ctx.fp, "},\n")
	}
	fmt.Fprintf(ctx.fp, "];\n")

	ctx.fp.Close()
	ctx.fp = nil
}

type object3d struct {
	Name string
	ColorNames []string

	TriangleCoords [][3]Point
	TriangleNorms [][3]Vector
	TriangleColors [][][3]Color

	LineCoords [][2]Point
	LineColors [][2]Color

	XRayTriangleCoords [][3]Point
	XRayTriangleNorms [][3]Vector
	XRayTriangleColors [][3]Color
	XRayLineCoords [][2]Point
	XRayLineColors [][2]Color
}


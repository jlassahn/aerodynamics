
package main

import (
	"fmt"
	"io"
	"os"
)

type SVGFile struct {
	io.WriteCloser
}

func CreateSVGFile(name string, width int, height int) (SVGFile, error) {

	fp, err := os.Create(name)
	if err != nil {
		return SVGFile{nil}, err
	}

	fmt.Fprintf(fp, "<?xml version=\"1.0\"?>\n")
	fmt.Fprintf(fp, "<svg width=\"%v\" height=\"%v\" version=\"1.1\" xmlns=\"http://www.w3.org/2000/svg\" stroke-linecap=\"round\">\n", width, height)

	return SVGFile{fp}, nil
}

func (svg SVGFile) Finalize() {

	fmt.Fprintf(svg, "</svg>\n")
	svg.Close()
}

func (svg SVGFile) Line(x0 float32, y0 float32, x1 float32, y1 float32, color int, width float32) {
	fmt.Fprintf(svg, "<line x1=\"%v\" y1=\"%v\" x2=\"%v\" y2=\"%v\" stroke=\"#%.6X\" stroke-width=\"%v\" />\n",
		x0, y0, x1, y1, color, width)
}

func (svg SVGFile) Curve(x0 float32, y0 float32, cx0 float32, cy0 float32, cx1 float32, cy1 float32, x1 float32, y1 float32, color int, width float32) {
	fmt.Fprintf(svg, "<path d=\"M %v %v C %v %v %v %v %v %v\" fill=\"none\" stroke=\"#%.6X\" stroke-width=\"%v\" />\n",
		x0, y0, cx0, cy0, cx1, cy1, x1, y1, color, width)
}

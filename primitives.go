
package main

func thick(x float32) float32 {

	return 1 - x*x*4
}

func center(x float32) float32 {
	if x < -0.25 {
		return  3*(-0.25 - x)
	}
	return 0
	/*
	return thick(x)*0.3
	return 0
	*/
}

func AddTestFlat(model *Model, dx float32, dy float32, dz float32) {

	var tip float32 = 0
	var n float32 = 20

	z0 := -dz/2
	for i := 0*n; i<n; i++ {

		z1 := z0 + dz/n

		for j := 0*n; j<n; j++ {

			xFrac0 := -0.5 + j/n
			yCenter0 := center(xFrac0)
			yThick0 := thick(xFrac0)*0.5

			xFrac1 := -0.5 + (j+1)/n
			yCenter1 := center(xFrac1)
			yThick1 := thick(xFrac1)*0.5

			x0 := dx*xFrac0
			y0 := dy*(yCenter0 + yThick0)
			x1 := dx*xFrac1
			y1 := dy*(yCenter1 + yThick1)

			panel := Panel {
				Points: [4]Point{
					{x0, y0, z0},
					{x0, y0, z1},
					{x1, y1, z1},
					{x1, y1, z0}},
				Count: 4,
				Strength: 1,
			}
			model.Panels = append(model.Panels, &panel)
		}
		for j := 0*n; j<n; j++ {

			xFrac0 := 0.5 - j/n
			yCenter0 := center(xFrac0)
			yThick0 := thick(xFrac0)*0.5

			xFrac1 := 0.5 - (j+1)/n
			yCenter1 := center(xFrac1)
			yThick1 := thick(xFrac1)*0.5

			x0 := dx*xFrac0
			y0 := dy*(yCenter0 - yThick0)
			x1 := dx*xFrac1
			y1 := dy*(yCenter1 - yThick1)

			panel := Panel {
				Points: [4]Point{
					{x0, y0, z0},
					{x0, y0, z1},
					{x1, y1, z1},
					{x1, y1, z0}},
				Count: 4,
				Strength: 1,
			}
			model.Panels = append(model.Panels, &panel)
		}

		de := dy*(center(-0.5 + 0.1/dx) - center(-0.5))
		norm := Vector{de, 0.1, 0}
		norm = norm.Scale(1/Sqrt(norm.Dot(norm)))

		edge := Edge {
			Center: Point{-dx/2 - 3*EPSILON, dy*center(-0.5), (z0+z1)/2},
			Normal: norm,
		}
		model.Edges = append(model.Edges, &edge)

		z0 = z1;
	}

	for j := 0*n; j<n; j++ {

		xFrac0 := -0.5 + j/n
		yCenter0 := center(xFrac0)
		yThick0 := thick(xFrac0)*0.5

		xFrac1 := -0.5 + (j+1)/n
		yCenter1 := center(xFrac1)
		yThick1 := thick(xFrac1)*0.5

		x0 := dx*xFrac0
		y0 := dy*yThick0
		yc0 :=  dy*yCenter0
		x1 := dx*xFrac1
		y1 := dy*yThick1
		yc1 :=  dy*yCenter1

		z0 = -dz/2
		panel := &Panel {
			Points: [4]Point{
				{x0, yc0, z0 - tip*y0},
				{x0, yc0 + y0, z0},
				{x1, yc1 + y1, z0},
				{x1, yc1, z0 - tip*y1}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)

		panel = &Panel {
			Points: [4]Point{
				{x1, yc1, z0 - tip*y1},
				{x1, yc1 - y1, z0},
				{x0, yc0 - y0, z0},
				{x0, yc0, z0 - tip*y0}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)

		z0 = dz/2
		panel = &Panel {
			Points: [4]Point{
				{x0, yc0 + y0, z0},
				{x0, yc0, z0 + tip*y0},
				{x1, yc1, z0 + tip*y1},
				{x1, yc1 + y1, z0}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)
		panel = &Panel {
			Points: [4]Point{
				{x0, yc0, z0 + tip*y0},
				{x0, yc0 - y0, z0},
				{x1, yc1 - y1, z0},
				{x1, yc1, z0 + tip*y1}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)
	}

	// FIXME wakes not in right position
	z0 = -dz/2
	for i := 0*n; i<=n; i++ {

		wake := &Wake {
			Points: []Point {
				{ -100, 0, z0},
				{ 0, 0, z0}},
				TreePath: []Point {{ 0, 0, 0}},
				BlurInternal: dy/4,
				BlurWake: dz/(4*n),
				Strength: 0,
		}
		model.Wakes = append(model.Wakes, wake)

		z0 = z0 + dz/n
	}

}


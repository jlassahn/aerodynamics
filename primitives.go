
package main

func AddTestFlat(model *Model, dx float32, dy float32, dz float32) {

	var tip float32 = 0
	var n float32 = 20

	z0 := -dz/2
	for i := 0*n; i<n; i++ {

		z1 := z0 + dz/n

		for j := 0*n; j<n; j++ {

			x0 := -dx/2 + j*dx/n
			y0 := dy/2 * (1 - (-1 + 2*j/n)*(-1 + 2*j/n))
			x1 := -dx/2 + (j+1)*dx/n
			y1 := dy/2 * (1 - (-1 + 2*(j+1)/n)*(-1 + 2*(j+1)/n))

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

			x0 := dx/2 - j*dx/n
			y0 := -dy/2 * (1 - (1 - 2*j/n)*(1 - 2*j/n))
			x1 := dx/2 - (j+1)*dx/n
			y1 := -dy/2 * (1 - (1 - 2*(j+1)/n)*(1 - 2*(j+1)/n))

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

		edge := Edge {
			Center: Point{-dx/2 - 3*EPSILON, 0, (z0+z1)/2},
			Normal: Vector{0, 1, 0},
		}
		model.Edges = append(model.Edges, &edge)

		z0 = z1;
	}

	for j := 0*n; j<n; j++ {

		x0 := -dx/2 + j*dx/n
		y0 := dy/2 * (1 - (-1 + 2*j/n)*(-1 + 2*j/n))
		x1 := -dx/2 + (j+1)*dx/n
		y1 := dy/2 * (1 - (-1 + 2*(j+1)/n)*(-1 + 2*(j+1)/n))

		z0 = -dz/2
		panel := &Panel {
			Points: [4]Point{
				{x0, 0, z0 - tip*y0},
				{x0, y0, z0},
				{x1, y1, z0},
				{x1, 0, z0 - tip*y1}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)

		panel = &Panel {
			Points: [4]Point{
				{x1, 0, z0 - tip*y1},
				{x1, -y1, z0},
				{x0, -y0, z0},
				{x0, 0, z0 - tip*y0}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)

		z0 = dz/2
		panel = &Panel {
			Points: [4]Point{
				{x0, y0, z0},
				{x0, 0, z0 + tip*y0},
				{x1, 0, z0 + tip*y1},
				{x1, y1, z0}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)
		panel = &Panel {
			Points: [4]Point{
				{x0, 0, z0 + tip*y0},
				{x0, -y0, z0},
				{x1, -y1, z0},
				{x1, 0, z0 + tip*y1}},
			Count: 4,
			Strength: 1,
		}
		model.Panels = append(model.Panels, panel)
	}

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


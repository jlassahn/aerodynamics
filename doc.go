
package main

/* Panel Method Solver

Model is an arbitrary mesh of triangles and quads enclosing a volume.
Some points on the model are wake vortex sources, these include a path
to a tree that connects them internally.  There must be at least two wake
vortices.
Some edges of the model have vorticity shedding parameters, a perpendicular
velocity vector (which the solver tries to minimize) and a strength roughly
proportional to the curvature at the edge.  These can be calculated from
the geometry, but we might want to specify an edge sharpness adjustment.

Each polygon has a normal and an area (compute these).

Each polygon has a control point at the centroid.

The wake vortices start out as straight lines going downstream parallel to
the free stream velocity.  We might want to solve iteratively to make the
wakes follow the actual fluid flow.  The exact positions of the internal
tree of wake paths shouldn't matter as long as they connect.

Each polygon acts as a source of constant strength spread over the surface.
Each wake (and the internal paths connecting them) act as a vortex of constant
strength over the length of the path.

Transform the wake lines into N-1 combinations, each consisting of one wake
at strength +1, and all the others at strength -1/(N-1).  To get N-1 of these
you have to leave one wake out, it doesn't matter which one.  This transform
enforces the constraint that vortex lines can't end.

Solving the model consists of optimizing for these constraints:
For each polygon the normal velocity is zero.
Minimize the weighted sum of the squares of the vorticity shedding edges over
variation in the wake lines.

This produces a set of N(polygon) + N(wake vortex)-1 linear equations, which
can be solved directly.
*/

/*
Eqns to solve:

	Velocity(i, point) = sum(Panel[i]) + sum(Wake[i])

	Sum_i(Velocity(i, Center[j]) . Normal[j] ) = -V0.Normal[j]
	dCost/dW = sum_i_j ( (EdgeNormal[j] . Velocity(i, Center[j])) * (EdgeNormal[j] . d.Velocity(i, Center[j])/dW) ) = 0
*/


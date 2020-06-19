
const VertexSource = `
	attribute vec4 vertex_point;
	uniform mat4 transform;
	void main()
	{
		gl_Position = transform*vertex_point;
	}
`;

const FragmentSource = `

	uniform lowp vec4 color;

	void main()
	{
		gl_FragColor = color;
	}
`;

class Graph3D
{
	constructor(name)
	{
		const canvas = document.getElementById(name);
		const gl = canvas.getContext("webgl");

		const vertex_shader = gl.createShader(gl.VERTEX_SHADER);
		gl.shaderSource(vertex_shader, VertexSource);
		gl.compileShader(vertex_shader);
		console.log(gl.getShaderInfoLog(vertex_shader));

		const fragment_shader = gl.createShader(gl.FRAGMENT_SHADER);
		gl.shaderSource(fragment_shader, FragmentSource);
		gl.compileShader(fragment_shader);
		console.log(gl.getShaderInfoLog(fragment_shader));

		const program = gl.createProgram();
		gl.attachShader(program, vertex_shader);
		gl.attachShader(program, fragment_shader);
		gl.linkProgram(program);
		console.log(gl.getProgramInfoLog(program));

		const point_buffer = gl.createBuffer();

		const transform = new Float32Array([
			0.5, 0,   0,   0,
			0,   0.5, 0,   0,
			0,   0,   0.4, 0.05,
			0,   0,   0.5, 1
		]);

		this.gl = gl;
		this.program = program;
		this.point_buffer = point_buffer;
		this.transform = transform;
	}

	StartFrame()
	{
		const gl = this.gl;

		gl.useProgram(this.program);

		gl.uniformMatrix4fv(
			gl.getUniformLocation(this.program, "transform"),
			false,
			this.transform);

		gl.clearColor(0,0,0,1);
		gl.clear(gl.COLOR_BUFFER_BIT);
	}

	EndFrame()
	{
	}

	DrawQuads(points, norms, colors)
	{
		const gl = this.gl;

		gl.bindBuffer(gl.ARRAY_BUFFER, this.point_buffer);
		gl.bufferData(gl.ARRAY_BUFFER, points, gl.STATIC_DRAW);
		gl.vertexAttribPointer(
			gl.getAttribLocation(this.program, "vertex_point"),
			3, //number of components
			gl.FLOAT,
			false, // normalize
			3*4, // stride
			0); // offset
		gl.enableVertexAttribArray(
			gl.getAttribLocation(this.program, "vertex_point"));

		for (var i=0; i<points.length; i+=12)
		{
			gl.uniform4fv(
				gl.getUniformLocation(this.program, "color"),
				colors.slice(i/3, i/3 +4));

			gl.drawArrays(gl.TRIANGLE_FAN, i/3, 4);
		}
	}
}


function main()
{
	console.log("Hello");

	var ctx = new Graph3D("view");

	ctx.StartFrame();
	ctx.DrawQuads(DATA_quads, DATA_quadnorms, DATA_quadcolors);
	ctx.EndFrame();

}


window.onload = main;


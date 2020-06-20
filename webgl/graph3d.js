
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
		gl.clearDepth(1.0);
		gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
		gl.enable(gl.DEPTH_TEST);
		gl.depthFunc(gl.LEQUAL);
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

function Multiply4x4(a, b)
{
	var ret = new Float32Array(16);

	for (var i=0; i<4; i++)
	for (var j=0; j<4; j++)
	{
		var val = 0;
		for (var k=0; k<4; k++)
			val += b[i + 4*k] * a[k + 4*j];
		ret[i + 4*j] = val;
	}
	return ret;
}


function main()
{
	console.log("Hello");

	ctx = new Graph3D("view");

	var controls = document.getElementById("view_controls");
	var inputs = controls.getElementsByTagName("input");

	for (var i of inputs)
	{
		i.oninput = UpdateGraph;
	}
	UpdateGraph();
}

function UpdateGraph()
{
	console.log("Update");

	var tx = new Float32Array([
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1
	]);

	var a = document.getElementById("view_roll").value*3.1415926/180;
	var sx = Math.sin(a);
	var cx = Math.cos(a);
	var m = new Float32Array([
		 cx,  0, sx,  0,
		  0,  1,  0,  0,
		-sx,  0, cx,  0,
		  0,  0,  0,  1
	]);
	tx = Multiply4x4(tx, m);

	var a = document.getElementById("view_pitch").value*3.1415926/180;
	var sx = Math.sin(a);
	var cx = Math.cos(a);
	var m = new Float32Array([
		 1,  0,  0,  0,
		 0, cx, sx,  0,
		 0,-sx, cx,  0,
		 0,  0,  0,  1
	]);
	tx = Multiply4x4(tx, m);

	var a = document.getElementById("view_yaw").value*3.1415926/180;
	var sx = Math.sin(a);
	var cx = Math.cos(a);
	var m = new Float32Array([
		 cx, sx,  0,  0,
		-sx, cx,  0,  0,
		  0,  0,  1,  0,
		  0,  0,  0,  1
	]);
	tx = Multiply4x4(tx, m);

	var a = document.getElementById("view_x").value*0.01;
	tx[12] = a;
	var a = document.getElementById("view_y").value*0.01;
	tx[13] = a;

	var a = document.getElementById("view_zoom").value*0.01;
	var m = new Float32Array([
		 a, 0, 0, 0,
		 0, a, 0, 0,
		 0, 0, a, 0,
		 0, 0, 0, 1
	]);
	tx = Multiply4x4(tx, m);

	tx = Multiply4x4(tx,
		new Float32Array([
			1,  0,  0,   0,
			0,  1,  0,   0,
			0,  0,  0.4, 0.05,
			0,  0,  0.5, 1
		]));

	ctx.transform = tx;

	ctx.StartFrame();
	ctx.DrawQuads(DATA_quads, DATA_quadnorms, DATA_quadcolors);
	ctx.EndFrame();
}


window.onload = main;


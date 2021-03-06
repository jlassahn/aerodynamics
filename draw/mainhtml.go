
package draw

var MainHTML = `
<html>
<head>
	<link rel="stylesheet" href="style.css" type="text/css">
	<script type="application/javascript" src="graph3d.js"></script>
	<script type="application/javascript" src="data.js"></script>
</head>
<body>
	<canvas id="view" width="512" height="512">
	</canvas><div id="view_controls">
		<h1> View Controls </h1>
		<label><input type="range" id="view_roll" min="-180" max="180" value="0"> Roll </label>
		<label><input type="range" id="view_pitch" min="-180" max="180" value="0"> Pitch </label>
		<label><input type="range" id="view_yaw" min="-180" max="180" value="0"> Yaw </label>
		<label><input type="range" id="view_x" min="-180" max="180" value="0"> X </label>
		<label><input type="range" id="view_y" min="-180" max="180" value="0"> Y </label>
		<label><input type="range" id="view_zoom" min="30" max="500" value="100"> Zoom </label>
	</div>
</body>
</html>
`


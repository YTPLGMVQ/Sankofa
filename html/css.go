package html

// build the inline CSS part of the HTML output
//
// #39634F -- dark green
// #636363 -- desaturated dark green
// #23A378 -- vivid green
// #A3A3A3 -- desaturated vivid green
// #B8CF5D -- light green
// #CFCFCF -- desaturated light green
// #E94A5E -- red
// #E8E8E8 -- desaturated red

// inline CSS for the HTML GUI
func CSS() string {
	var r string
	r += `<style>
body, p {
	background-color: #E8E8E8;
	color: #636363;
	width: 1000px;
	text-align: justify;
	font-family: Georgia,Courier New;
}
a:link, a:visited {
	background-color: #E8E8E8;
	color: #39634F;
	text-decoration: none;
}
a:hover {
	color: #39634F;
	font-weight: bold;
}
a.cur {
	color: #E94A5E;
	font-weight: bold;
}
a.cont {
	color: #B8CF5D;
}
h1, h2, h3 {
	background-color: #E8E8E8;
	color: #39634F;
	text-align: center;
}
table, th, td {
	background-color: #E8E8E8;
	color: #636363;
	border-color: #CFCFCF;
	table-layout: fixed;
	overflow: hidden;
	text-align: center;
	border-width: 0px;
	border-collapse: collapse;
	border-style: inset;
	margin: 0px auto;
}
th {
	color: #39634F;
}
#house {
	fill: #B8CF5D;
	stroke: #39634F;
	stroke-width: 2;
}
#check {
	fill: #B8CF5D;
	stroke: #E94A5E;
	stroke-width: 2;
}
#stone {
	fill: #23A378;
	stroke: #39634F;
	stroke-width: 2;
}
#moved {
	fill: #A3A3A3;
	stroke: #636363;
	stroke-width: 2;
}
#left {
	text-align: left;
	min-width: 55px;
}
/</style>
`
	return r
}

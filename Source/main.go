package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 500
	height = 500

	rows    = 20
	columns = 20
)

var (
	square = []float32{
		// X,    Y,    Z,    U,    V
		-0.5, 0.5, 0.0, 0.0, 1.0, // 1
		-0.5, -0.5, 0.0, 0.0, 0.0, // 2
		0.5, -0.5, 0.0, 1.0, 0.0, // 3

		-0.5, 0.5, 0.0, 0.0, 1.0, // 4
		0.5, 0.5, 0.0, 1.0, 1.0, // 5
		0.5, -0.5, 0.0, 1.0, 0.0, // 6
	}

	program uint32
)

type cell struct {
	x        int
	y        int
	drawable uint32
}

func init() {
	runtime.LockOSThread()
}

func main() {
	window := initGlfw()
	defer glfw.Terminate()

	program = initOpenGL()

	texture, err := newTexture("Data/green.png")
	if err != nil {
		log.Fatal(err)
	}

	cells := makeCells()

	for !window.ShouldClose() {
		drawCells(cells, window, program, texture)
	}
}

func drawCells(cells [][]*cell, window *glfw.Window, program, texture uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	//for x := range cells {
	//	for _, c := range cells[x] {
	//		c.draw()
	//	}
	//}

	drawTree(cells)

	glfw.PollEvents()
	window.SwapBuffers()
}

func drawTree(cells [][]*cell) {
	cells[0][0].draw()
	cells[18][18].draw()
}

func makeCells() [][]*cell {
	cells := make([][]*cell, rows, rows)

	for x := 0; x < rows; x++ {
		for y := 0; y < columns; y++ {
			c := newCell(x, 19-y)
			cells[x] = append(cells[x], c)
		}
	}

	return cells
}

func newCell(x, y int) *cell {
	points := make([]float32, len(square), len(square))
	copy(points, square)

	stride := 0
	for i := 0; i < len(points); i++ {
		var position float32
		var size float32

		stride++
		if stride == 6 {
			stride = 1
		}

		switch stride {
		case 1:
			size = 1.0 / float32(columns)
			position = float32(x) * size
		case 2:
			size = 1.0 / float32(rows)
			position = float32(y) * size
		default:
			continue
		}

		if points[i] < 0 {
			points[i] = (position * 2) - 1
		} else {
			points[i] = ((position + size) * 2) - 1
		}
	}

	return &cell{
		x:        x,
		y:        y,
		drawable: makeVao(points),
	}
}

func (c *cell) draw() {
	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/5))
}

func onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if key == glfw.KeyEscape {
			w.SetShouldClose(true)
		}
		if key == glfw.KeySpace {
			var wireframe int32
			gl.GetIntegerv(gl.POLYGON_MODE, &wireframe)
			switch wireframe {
			case gl.FILL:
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			case gl.LINE:
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.POINT)
				gl.PointSize(20.0)
			case gl.POINT:
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			}
		}
	}
}

func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	window, err := glfw.CreateWindow(width, height, "Pacman", nil, nil)
	if err != nil {
		panic(err)
	}

	window.SetKeyCallback(onKey)

	window.MakeContextCurrent()

	return window
}

func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	log.Println("OpenGL version", gl.GoStr(gl.GetString(gl.VERSION)))

	vertexShader, err := compileShader(readShaderCode("Data/Shaders/square.vertexShader.c"), gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(readShaderCode("Data/Shaders/square.fragmentShader.c"), gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	gl.UseProgram(program)

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))

	return prog
}

func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 4*5, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*5, gl.PtrOffset(4*3))
	gl.EnableVertexAttribArray(1)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		err := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(err))

		return 0, fmt.Errorf("failed to compile %v: %v", source, err)
	}

	return shader, nil
}

func readShaderCode(filePath string) string {
	code := ""
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		code += "\n" + scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	code += "\x00"
	return code
}

func newTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

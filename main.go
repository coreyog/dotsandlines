package main

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"sort"

	"github.com/coreyog/diesel"
	"github.com/coreyog/diesel/util"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)

const (
	WindowWidth  = 1024
	WindowHeight = 768
	BufferScale  = 4
	SizeW        = WindowWidth * BufferScale
	SizeH        = WindowHeight * BufferScale
	DotSize      = 20
	LineWidth    = 2
)

var DotColor = colornames.White
var LineLength = 200.0 * BufferScale

type Dot struct {
	pos       pixel.Vec
	direction pixel.Vec
	speed     float64
}

func (d *Dot) update() {
	d.pos = d.pos.Add(d.direction.Scaled(d.speed / 15.0))
	if (d.pos.X < 0 && d.direction.X < 0) || (d.pos.X > SizeW && d.direction.X > 0) {
		d.direction.X *= -1
		d.speed = rand.Float64()*3 + 2
	}

	if (d.pos.Y < 0 && d.direction.Y < 0) || (d.pos.Y > SizeH && d.direction.Y > 0) {
		d.direction.Y *= -1
		d.speed = rand.Float64()*3 + 2
	}
}

func (d *Dot) draw(target *imdraw.IMDraw) {
	target.Color = DotColor
	target.Push(d.pos)
	target.Circle(DotSize, 0)
}

func main() {
	opengl.Run(run)
}

func run() {
	cfg := opengl.WindowConfig{
		Title:       "Dots and Lines",
		Bounds:      pixel.R(0, 0, WindowWidth, WindowHeight),
		VSync:       true,
		SamplesMSAA: 0,
	}

	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(true)

	buffer := opengl.NewCanvas(pixel.R(0, 0, SizeW, SizeH))
	buffer.SetSmooth(true)

	mapper := util.BuildMapper(buffer.Bounds(), win.Bounds())

	dt := diesel.NewDeltaTimer(0, 60)

	dots := make([]*Dot, 0, 75)
	for range cap(dots) {
		d := &Dot{
			pos:       pixel.V(float64(rand.N(SizeW)), float64(rand.N(SizeH))),
			direction: pixel.V(float64(rand.N(10)-5), float64(rand.N(10)-5)).Unit(),
			speed:     rand.Float64()*3 + 2,
		}
		dots = append(dots, d)
	}

	for !win.Closed() {
		dt.Tick()

		win.Clear(colornames.Black)
		win.SetTitle(fmt.Sprintf("Dots and Lines | FPS: %.1f", dt.FPS()))

		// escape
		if win.JustPressed(pixel.KeyEscape) {
			win.SetClosed(true)
		}

		// " "
		if win.Pressed(pixel.KeySpace) {
			d := &Dot{
				pos:       pixel.V(float64(rand.N(SizeW)), float64(rand.N(SizeH))),
				direction: pixel.V(float64(rand.N(10)-5), float64(rand.N(10)-5)).Unit(),
				speed:     rand.Float64()*3 + 2,
			}
			dots = append(dots, d)
			fmt.Printf("Dots: %d\n", len(dots))
		}

		// "\b"
		if win.Pressed(pixel.KeyBackspace) && len(dots) > 1 {
			dots = dots[:len(dots)-1]
			fmt.Printf("Dots: %d\n", len(dots))
		}

		// "+", but it's actually =
		if win.Pressed(pixel.KeyEqual) {
			LineLength++
			fmt.Printf("Length: %.1f\n", LineLength)
		}

		// -
		if win.Pressed(pixel.KeyMinus) && LineLength > 0 {
			LineLength--
			fmt.Printf("Length: %.1f\n", LineLength)
		}

		// stats
		if win.JustPressed(pixel.KeyS) {
			fmt.Printf("Dots: %d\n", len(dots))
			fmt.Printf("Length: %.1f\n", LineLength)
		}

		for _, d := range dots {
			d.update()
		}

		frame := imdraw.New(nil)
		drawLines(frame, dots)

		for _, d := range dots {
			d.draw(frame)
		}

		frame.Draw(buffer)

		buffer.Draw(win, mapper)

		win.Update()
		buffer.Clear(color.Black)
	}
}

func drawLines(frame *imdraw.IMDraw, dots []*Dot) {
	type line struct {
		d1, d2 *Dot
		alpha  float64
	}

	lines := make([]line, 0, len(dots)*(len(dots)-1)/2)

	for i, d1 := range dots {
		for _, d2 := range dots[i+1:] {
			dist := d1.pos.To(d2.pos).Len()
			if dist >= LineLength {
				continue
			}

			lines = append(lines, line{d1, d2, dist / LineLength})
		}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].alpha > lines[j].alpha
	})

	for _, l := range lines {
		frame.Color = DotColor
		frame.Intensity = l.alpha
		frame.Push(l.d1.pos, l.d2.pos)
		frame.Line(LineWidth)
	}
}

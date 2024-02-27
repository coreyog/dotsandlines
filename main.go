package main

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"github.com/coreyog/diesel"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	SizeW   = 1024
	SizeH   = 768
	DotSize = 5
)

var DotColor = colornames.Blue
var LineLength = 200.0

type Dot struct {
	pos      pixel.Vec
	velocity pixel.Vec
	speed    float64
}

func (d *Dot) update() {
	d.pos = d.pos.Add(d.velocity.Scaled(d.speed))
	if (d.pos.X < 0 && d.velocity.X < 0) || (d.pos.X > SizeW && d.velocity.X > 0) {
		d.velocity.X *= -1
		d.speed = rand.Float64()*3 + 2
	}

	if (d.pos.Y < 0 && d.velocity.Y < 0) || (d.pos.Y > SizeH && d.velocity.Y > 0) {
		d.velocity.Y *= -1
		d.speed = rand.Float64()*3 + 2
	}
}

func (d *Dot) draw(target *imdraw.IMDraw) {
	target.Color = DotColor
	target.Push(d.pos)
	target.Circle(DotSize, 0)
}

func main() {
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Dots and Lines",
		Bounds: pixel.R(0, 0, SizeW, SizeH),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(true)

	dt := diesel.NewDeltaTimer(0, 60)

	dots := make([]*Dot, 0, 75)
	for range cap(dots) {
		d := &Dot{
			pos:      pixel.V(float64(rand.N(1024)), float64(rand.N(768))),
			velocity: pixel.V(float64(rand.N(10)-5), float64(rand.N(10)-5)).Unit(),
			speed:    rand.Float64()*3 + 2,
		}
		dots = append(dots, d)
	}

	for !win.Closed() {
		dt.Tick()

		win.Clear(colornames.Black)
		win.SetTitle(fmt.Sprintf("Dots and Lines | FPS: %.1f", dt.FPS()))

		// escape
		if win.JustPressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
		}

		// " "
		if win.Pressed(pixelgl.KeySpace) {
			d := &Dot{
				pos:      pixel.V(float64(rand.N(1024)), float64(rand.N(768))),
				velocity: pixel.V(float64(rand.N(10)-5), float64(rand.N(10)-5)).Unit(),
				speed:    rand.Float64()*3 + 2,
			}
			dots = append(dots, d)
			fmt.Printf("Dots: %d\n", len(dots))
		}

		// "\b"
		if win.Pressed(pixelgl.KeyBackspace) && len(dots) > 1 {
			dots = dots[:len(dots)-1]
			fmt.Printf("Dots: %d\n", len(dots))
		}

		// "+", but it's actually =
		if win.Pressed(pixelgl.KeyEqual) {
			LineLength++
			fmt.Printf("Length: %.1f\n", LineLength)
		}

		// -
		if win.Pressed(pixelgl.KeyMinus) && LineLength > 0 {
			LineLength--
			fmt.Printf("Length: %.1f\n", LineLength)
		}

		// stats
		if win.JustPressed(pixelgl.KeyS) {
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

		frame.Draw(win)

		win.Update()
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
		frame.Line(1)
	}
}

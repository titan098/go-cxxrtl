package waveform

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	svg "github.com/ajstarks/svgo"
)

const (
	signalHeight = 20
	signalGap    = 10
	stepWidth    = 20
	leftMargin   = 150
)

const (
	backgroundStyle = "fill:rgba(20,20,20,1)"
	wireStyle       = "stroke:green;stroke-width:1;"
	shadowStyle     = "stroke:rgba(0,0,0,0.5);stroke-width:1;"
	busStyle        = "stroke:cyan;stroke-width:1"
	busFillStyle    = "fill:cyan;fill-opacity:0.1"
	busValueStyle   = "font-size:10px; font-family:monospace; text-anchor:start; fill:white; text-shadow:1px 1px 1px black;"
	textStyle       = "font-family:monospace; font-size:12px; fill:white; text-shadow:1px 1px 1px black;"
	tickTextStyle   = "font-size:10px; font-family:monospace; text-anchor:middle; fill:white; text-shadow:1px 1px 1px black;"
	tickStyle       = "stroke:grey;stroke-width:1"
	gridStyle       = "stroke:#303030;stroke-width:1;stroke-dasharray:1,1"
	axisStyle       = "stroke:#606060;stroke-width:2"
)

// drawLineWithShadow draws a line from (x0,y0) to (x1,y1) with a shadow effect.
// It first draws a shadow line with a slight offset and then draws the main line
// using the specified style.
func drawLineWithShadow(canvas *svg.SVG, x0 int, y0 int, x1 int, y1 int, style string) {
	if y0 == y1 {
		canvas.Line(x0, y0+1, x1, y1+1, shadowStyle)
	} else {
		canvas.Line(x0+1, y0, x1+1, y1, shadowStyle)
	}
	canvas.Line(x0, y0, x1, y1, style)
}

// DrawSVG generates an SVG waveform visualization from simulation data.
// It takes a map of simulation data where the outer map is indexed by time and the inner map
// is indexed by signal name, and a list of signal names to be displayed.
// Returns the SVG as a byte slice.
func DrawSVG(vcdData *VcdData) []byte {
	var out bytes.Buffer
	sim := vcdData.Sim
	signals := vcdData.Signals
	outputBuffer := bufio.NewWriter(&out)

	width := len(sim)*stepWidth + leftMargin + 10
	height := len(signals)*(signalHeight+signalGap) + 100

	canvas := svg.New(outputBuffer)
	canvas.Start(width, height)
	canvas.Rect(0, 0, width, height, backgroundStyle)

	// Sort time steps
	times := make([]uint64, 0, len(sim))
	for t := range sim {
		times = append(times, t)
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	// Determine the maximum time
	maxTime := times[len(times)-1]

	// Add vertical dotted grid lines and time markers
	gridTop := 40
	gridBottom := height - 30
	for t := 0; t <= int(maxTime); t++ {
		x := t*stepWidth + leftMargin
		strokeStyle := gridStyle
		if t == 0 {
			strokeStyle = axisStyle
		}
		canvas.Line(x, gridTop, x, gridBottom, strokeStyle)

		// Draw tick and label at the top
		canvas.Line(x, 35, x, 45, tickStyle)
		canvas.Text(x, 30, fmt.Sprintf("%d", t), tickTextStyle)
	}

	y := 50
	for _, sig := range signals {
		canvas.Text(10, y+signalHeight/2, sig, textStyle)

		var lastVal string
		var lastX int
		lastLabel := ""
		for i, t := range times {
			x := int(t)*stepWidth + leftMargin
			val := sim[t][sig]

			if i == 0 {
				lastVal = val
				lastX = x
				continue
			}

			isBus := len(val) > 1 || (val != "0" && val != "1")

			if isBus {
				yTop := y
				yBottom := y + (3 * signalHeight / 4)

				// Fill area between bus lines
				canvas.Polygon([]int{lastX, x, x, lastX}, []int{yTop, yTop, yBottom, yBottom}, busFillStyle)

				if val != lastVal {
					// "X" crossing to denote change
					drawLineWithShadow(canvas, lastX, yTop, x, yBottom, busStyle)
					drawLineWithShadow(canvas, lastX, yBottom, x, yTop, busStyle)

				} else {
					// Draw double line for the bus
					drawLineWithShadow(canvas, lastX, yTop, x, yTop, busStyle)
					drawLineWithShadow(canvas, lastX, yBottom, x, yBottom, busStyle)

					// Display value in between lines
					label := val
					if len(label) > 8 {
						bits := strings.TrimPrefix(label, "b")
						if i, err := strconv.ParseUint(bits, 2, 64); err == nil {
							label = fmt.Sprintf("0x%X", i)
						}
					}

					if lastLabel != label {
						canvas.Text(lastX+1, y+(signalHeight/2), label, busValueStyle)
						lastLabel = label
					}
				}
			} else {
				y0 := y + signalHeight
				if lastVal == "1" {
					y0 = y
				}
				y1 := y + signalHeight
				if val == "1" {
					y1 = y
				}

				drawLineWithShadow(canvas, lastX, y0, x, y0, wireStyle)
				if lastVal != val {
					drawLineWithShadow(canvas, x, y0, x, y1, wireStyle)
				}
			}

			lastX = x
			lastVal = val
		}
		y += signalHeight + signalGap
	}

	canvas.End()
	outputBuffer.Flush()
	return out.Bytes()
}
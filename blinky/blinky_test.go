package blinky_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/titan098/go-cxxrtl"
	"github.com/titan098/go-vcd2svg/waveform"
)

func setup() (*cxxrtl.Cxxrtl, *cxxrtl.Vcd) {
	cxx := cxxrtl.NewCxxrtl()
	vcd := cxxrtl.NewCxxrtlVcd()
	_ = vcd.Timescale(1, cxxrtl.MICROSECOND)
	vcd.AddFromWithoutMemories(cxx)
	return cxx, vcd
}

// TestBlinky is an example testbench for blinky.v
// This will create an instance of the cxxrtl and vcd object and step through a simulation for 10 cycles.
// The code demonstrates stepping through the simulation, asserting on signal values and emitting
// a vcd and svg output waveform.
func TestBlinky(t *testing.T) {
	cxx, vcd := setup()
	defer cxx.Delete()
	defer vcd.Destroy()

	clk := cxx.GetPart("clk").AsValue()
	blink := cxx.GetPart("blink").AsWire()

	cxx.Reset()
	cxx.Step()
	vcd.Sample(0)
	for i := range 10 {
		clk.Next(1)
		cxx.Step()
		vcd.Sample(i*2 + 0)

		clk.Next(0)
		cxx.Step()
		vcd.Sample(i*2 + 1)
	}

	// blink should be 0 at the end on the simulation
	assert.Equal(t, uint64(0), blink.Current())

	// Save the VCD file
	b := vcd.Read()
	f, _ := os.CreateTemp("", "test.*.vcd")
	defer f.Close()
	f.Write(b)

	// Save the output SVG file
	svgBytes, _ := waveform.SvgFromFile(f.Name())
	s, _ := os.CreateTemp("", "test.*.svg")
	defer s.Close()
	s.Write(svgBytes)
}
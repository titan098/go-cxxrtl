package cxxrtl_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/titan098/go-cxxrtl"
)

func TestMain(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	defer cxx.Delete()

	cxx.Reset()
	assert.True(t, cxx != nil)
}

func TestVcdGeneration(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	defer cxx.Delete()

	vcd := cxxrtl.NewCxxrtlVcd()
	defer vcd.Destroy()

	_ = vcd.Timescale(1, cxxrtl.MICROSECOND)
	vcd.AddFromWithoutMemories(cxx)

	clk := cxx.GetPart("clk")

	assert.True(t, clk != nil)
	clk.Next(1)
	cxx.Step()
	vcd.Sample(0)

	c := vcd.Read()
	assert.NotNil(t, c)
	assert.Contains(t, string(c), "$timescale 1 us")
}

func TestGetPart(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	defer cxx.Delete()

	clk := cxx.GetPart("clk")

	assert.True(t, clk != nil)
	clk.Next(0)
	cxx.Step()
	assert.Equal(t, uint64(0), clk.Current())

	clk.Next(1)
	cxx.Step()
	assert.Equal(t, uint64(1), clk.Current())
}

func TestGetMemory(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	defer cxx.Delete()
	cxx.Reset()

	// We expect that we will get a [][]bytes defined by the 
	// depth and width of the memory. In this case there is 
	// an array of 32 bit works.
	mem := cxx.GetPart("MEM").AsMemory().Current()

	assert.True(t, slices.Compare(mem[0], []byte{0xde, 0, 0, 0}) == 0)
	assert.True(t, slices.Compare(mem[1], []byte{0xad, 0, 0, 0}) == 0)
	assert.True(t, slices.Compare(mem[2], []byte{0xbe, 0, 0, 0}) == 0)
	assert.True(t, slices.Compare(mem[3], []byte{0xef, 0, 0, 0}) == 0)

	// Retrive the memory as a continuous slice of bytes.
	memBytes := cxx.GetPart("MEM").AsMemory().Bytes()
	assert.Equal(t, uint8(0xde), memBytes[0])
	assert.Equal(t, uint8(0xad), memBytes[4])
	assert.Equal(t, uint8(0xbe), memBytes[8])
	assert.Equal(t, uint8(0xef), memBytes[12])
}
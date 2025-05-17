package cxxrtl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/titan098/go-cxxrtl"
)

func TestMain(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	assert.True(t, cxx != nil)
}

func TestGetPart(t *testing.T) {
	cxx := cxxrtl.NewCxxrtl()
	clk := cxx.GetPart("clk")

	assert.True(t, clk != nil)
	clk.Next(0)
	cxx.Step()
	assert.Equal(t, uint64(0), clk.Current())

	clk.Next(1)
	cxx.Step()
	assert.Equal(t, uint64(1), clk.Current())
}
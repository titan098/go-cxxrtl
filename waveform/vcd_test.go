package waveform

import (
	"os"
	"strings"
	"testing"

	"github.com/filmil/go-vcd-parser/vcd"
	"github.com/stretchr/testify/assert"
)

const simpleVcd = `$date
  Date text
$end
$version
  test
$end
$timescale 1ns $end
$scope module test $end
$var wire 1 ! clk $end
$var wire 1 " rst $end
$upscope $end
$enddefinitions $end
#0
0!
1"
#1
1!
0"
#2
0!
1"
`

func TestProcessVcd(t *testing.T) {
	parser := vcd.NewParser[vcd.File]()
	ast, err := parser.Parse("blah", strings.NewReader(simpleVcd))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vcdData := processVcd(ast)

	assert.Len(t, vcdData.Signals, 2)
	assert.Len(t, vcdData.Sim, 3)
	assert.Contains(t, vcdData.Signals, "test clk")
	assert.Contains(t, vcdData.Signals, "test rst")
}

func TestSvgFromBytes_Valid(t *testing.T) {
	svg, err := SvgFromBytes([]byte(simpleVcd))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Contains(t, string(svg), "<svg")
	assert.Contains(t, string(svg), "clk")
	assert.Contains(t, string(svg), "rst")
}

func TestSvgFromBytes_Invalid(t *testing.T) {
	_, err := SvgFromBytes([]byte("$This is not a VCD$"))
	if err == nil {
		t.Error("expected parse error for invalid VCD input, got none")
	}
}

func TestSvgFromFile_FileNotExist(t *testing.T) {
	_, err := SvgFromFile("/this/should/not/exist.vcd")
	if err == nil {
		t.Error("expected error when reading missing file, got none")
	}
}

func TestSvgFromFile_Valid(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.vcd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString(simpleVcd)
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	svg, err := SvgFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Contains(t, string(svg), "<svg")
}

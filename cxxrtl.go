package cxxrtl

/*
#cgo CFLAGS: -g -O3 -I${SRCDIR}/include -I/opt/oss-cad-suite/share/yosys/include/backends/cxxrtl/runtime/
#cgo LDFLAGS: -lstdc++ -L${SRCDIR}/lib -lcxxrtl
#include <soc.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Cxxrtl represents a CXXRTL simulation instance that manages the top-level design
// and handles C-string memory management.
type Cxxrtl struct {
	top    C.cxxrtl_toplevel // Top-level design representation
	handle C.cxxrtl_handle   // Handle to the CXXRTL simulation instance

	cstrHeap map[string]*C.char // Cache of C strings to prevent memory leaks
}

type Vcd struct {
	vcd C.cxxrtl_vcd

	sampleCalled bool
	cstrHeap     map[string]*C.char
}

const (
	SECOND      = "s"
	MILLISECOND = "ms"
	MICROSECOND = "us"
	NANOSECOND  = "ns"
	PICOSECOND  = "ps"
	FEMTOSECOND = "fs"
)

// Part represents a basic component in the CXXRTL design with access to its
// underlying C structure object.
type Part struct {
	obj *C.struct_cxxrtl_object // Pointer to the underlying CXXRTL object
}

// Value represents a value type component in the CXXRTL design.
// It embeds the Part type for basic component functionality.
type Value struct {
	Part
}

// Wire represents a wire type component in the CXXRTL design.
// It embeds the Part type for basic component functionality.
type Wire struct {
	Part
}

// Memory represents a memory type component in the CXXRTL design.
// It embeds the Part type for basic component functionality.
type Memory struct {
	Part
}

// Alias represents an alias type component in the CXXRTL design.
// It embeds the Part type for basic component functionality.
type Alias struct {
	Part
}

// Outline represents an outline type component in the CXXRTL design.
// It embeds the Part type for basic component functionality.
type Outline struct {
	Part
}

// PartType is an interface that constrains a type to be one of the CXXRTL component types:
// Value, Wire, Memory, Alias, or Outline. This interface is used with type parameters
// to provide type safety when working with different component types.
type PartType interface {
	Value | Wire | Memory | Alias | Outline
}

// Type constants representing the different kinds of components in a CXXRTL design.
// These constants are used internally to identify the type of component when
// converting between generic Part objects and specific typed objects.
const (
	TypeValue   = iota // Represents a simple value component
	TypeWire           // Represents a wire component that can carry signals
	TypeMemory         // Represents a memory component that stores data
	TypeAlias          // Represents an alias component that refers to another component
	TypeOutline        // Represents an outline component that encapsulates a submodule
)

// AsMemory converts the generic Part to a Memory if it represents a memory component in the design.
// This method checks if the underlying object is of memory type and returns the appropriate
// typed representation. This allows for type-safe access to memory-specific functionality.
func (p *Part) AsMemory() *Memory {
	if p.obj._type == TypeMemory {
		return &Memory{Part{p.obj}}
	}
	return nil
}

// AsValue converts the generic Part to a Value if it represents a value component in the design.
// This method checks if the underlying object is of value type and returns the appropriate
// typed representation. This allows for type-safe access to value-specific functionality.
func (p *Part) AsValue() *Value {
	if p.obj._type == TypeValue {
		return &Value{Part{p.obj}}
	}
	return nil
}

// AsWire returns the part as a Wire if it is of Wire type, otherwise returns nil.
// This method allows safe type conversion from a generic Part to a typed Wire.
func (p *Part) AsWire() *Wire {
	if p.obj._type == TypeWire {
		return &Wire{Part{p.obj}}
	}
	return nil
}

func (p *Part) AsAlias() *Alias {
	if p.obj._type == TypeAlias {
		return &Alias{Part{p.obj}}
	}
	return nil
}

func (p *Part) AsOutline() *Outline {
	if p.obj._type == TypeOutline {
		return &Outline{Part{p.obj}}
	}
	return nil
}

// Current returns the current value of the part as a uint64.
// This method reads the value directly from the underlying C object.
func (p *Part) Current() uint64 {
	return uint64(*(p.obj.curr))
}

// Next sets the next value for this part to be committed in a future cycle.
// It properly handles bit widths and chunks the data into 32-bit segments
// as required by the underlying C implementation.
//
// Parameters:
//   - value: The uint64 value to set
//
// Returns an error if the part does not support being updated.
func (p *Part) Next(value uint64) error {
	if p.obj.next == nil {
		return fmt.Errorf("part not updatable")
	}

	width := uint(p.obj.width)
	depth := uint(p.obj.depth)
	chunks := ((width + 31) / 32) * depth

	if width < 64 {
		value &= (1 << width) - 1
	}

	nextPtr := (*[1 << 30]uint32)(unsafe.Pointer(p.obj.next))

	for i := uint(0); i < chunks; i++ {
		nextPtr[i] = uint32(value & 0xFFFFFFFF)
		value >>= 32
	}

	return nil
}

// Current returns the current content of the memory as a 2D byte slice.
// The outer slice represents memory rows (depth), and inner slices represent
// individual memory words (width).
//
// Returns a slice of byte slices, where each inner slice represents one word of memory.
func (m *Memory) Current() [][]byte {
	widthBits := int(m.obj.width)
	depth := int(m.obj.depth)

	widthBytes := (widthBits + 7) / 8

	totalChunks := ((widthBits + 31) / 32) * depth
	totalBytes := totalChunks * 4
	currPtr := unsafe.Pointer(m.obj.curr)
	data := unsafe.Slice((*byte)(currPtr), totalBytes)

	out := make([][]byte, depth)
	for i := range depth {
		start := i * widthBytes
		end := start + widthBytes
		out[i] = data[start:end]
	}

	return out
}

// Bytes returns the memory as a contiguous byte slice regardless of depth.
func (m *Memory) Bytes() []byte {
	widthBits := int(m.obj.width)
	depth := int(m.obj.depth)

	totalChunks := ((widthBits + 31) / 32) * depth
	totalBytes := totalChunks * 4
	currPtr := unsafe.Pointer(m.obj.curr)
	return unsafe.Slice((*byte)(currPtr), totalBytes)
}

// Next is not supported for Memory objects and always returns an error.
// This method is implemented to satisfy interface requirements, but memory
// updates must be performed through other mechanisms.
func (m *Memory) Next(value uint64) error {
	return fmt.Errorf("memory does not support setting a next")
}

// getCStr returns a C string pointer for the given Go string.
// It caches C strings to prevent memory leaks and ensure consistent
// pointer values for the same string.
//
// Parameters:
//   - s: The Go string to convert to a C string
//
// Returns a pointer to a C string that should not be manually freed.
func getCStr(s string, cstrHeap map[string]*C.char) *C.char {
	cStr, ok := cstrHeap[s]
	if !ok {
		cStr = C.CString(s)
		cstrHeap[s] = cStr
	}
	return cStr
}

// GetPart retrieves a design component by its fully qualified path.
//
// Parameters:
//   - part: A string representing the hierarchical path to the component
//
// Returns a pointer to a Part that represents the requested component.
func (cxx *Cxxrtl) GetPart(part string) *Part {
	obj := C.cxxrtl_get(cxx.handle, getCStr(part, cxx.cstrHeap))
	return &Part{obj}
}

// Eval evaluates the design, computing outputs for the current inputs.
// This updates the current state of all components based on their inputs.
//
// Returns an int32 result code from the underlying C evaluation function.
func (cxx *Cxxrtl) Eval() int32 {
	return int32(C.cxxrtl_eval(cxx.handle))
}

// Commit propagates scheduled next values to current values.
// This moves all values set with Next() methods into the Current() state.
//
// Returns an int32 result code from the underlying C commit function.
func (cxx *Cxxrtl) Commit() int32 {
	return int32(C.cxxrtl_commit(cxx.handle))
}

// Step advances the simulation by one clock cycle.
// This is a high-level operation that combines Eval and Commit as needed.
//
// Returns a uint64 representing the simulation step or time.
func (cxx *Cxxrtl) Step() uint64 {
	return uint64(C.cxxrtl_step(cxx.handle))
}

func (cxx *Cxxrtl) OutlineEval(outline *Outline) *Outline {
	C.cxxrtl_outline_eval(outline.obj.outline)
	return outline
}

// Reset resets the simulation to its initial state.
// This restores all components to their default values.
func (cxx *Cxxrtl) Reset() {
	C.cxxrtl_reset(cxx.handle)
}

// Delete frees all resources associated with this simulation instance.
// This includes the C strings cache and the underlying CXXRTL simulation handle.
// The instance should not be used after calling Delete.
func (cxx *Cxxrtl) Delete() {
	C.cxxrtl_destroy(cxx.handle)
	for _, v := range cxx.cstrHeap {
		C.free(unsafe.Pointer(v))
	}
}

// Add registers a design component with the VCD instance for later output during simulation.
func (v *Vcd) Add(name string, p Part) {
	C.cxxrtl_vcd_add(v.vcd, getCStr(name, v.cstrHeap), p.obj)
}

// AddFrom adds all signals from the given CXXRTL design to the VCD trace.
// This includes both regular signals and memory signals.
func (v *Vcd) AddFrom(cxx *Cxxrtl) {
	C.cxxrtl_vcd_add_from(v.vcd, cxx.handle)
}

// AddFromWithoutMemories adds all signals from the given CXXRTL design to the VCD trace.
// Note: Despite the name, this function actually excludes memory signals due to the
// underlying C function called (cxxrtl_vcd_add_from_without_memories).
// Consider renaming this method to AddFromWithoutMemories for clarity.
func (v *Vcd) AddFromWithoutMemories(cxx *Cxxrtl) {
	C.cxxrtl_vcd_add_from_without_memories(v.vcd, cxx.handle)
}

func (v *Vcd) Read() []byte {
	var size C.size_t
	var data *C.char

	size = 1
	out := make([]byte, 0)

	// Call the C function correctly - it expects pointers to the data and size variables
	for size > 0 {
		C.cxxrtl_vcd_read(v.vcd, &data, &size)

		// If no data is available, return an empty slice
		if size == 0 || data == nil {
			return out
		}

		// Convert the C data to a Go byte slice
		// Note: this data is valid only until the next call to cxxrtl_vcd_sample or cxxrtl_vcd_read
		out = append(out, C.GoBytes(unsafe.Pointer(data), C.int(size))...)
	}
	return out
}

// Sample captures the current state of all registered signals at a specific simulation time.
// This function records the value of all signals added to the VCD instance for later output.
// After calling Sample, you can use Read() to retrieve the VCD data.
//
// Parameters:
//   - time: The simulation timestamp (in whatever time unit the simulation is using)
//     when this sample was taken. This value will appear in the VCD output.
//
// Note: Once Sample has been called, Timescale cannot be modified anymore.
func (v *Vcd) Sample(time int) {
	v.sampleCalled = true
	C.cxxrtl_vcd_sample(v.vcd, C.uint64_t(time))
}

func (v *Vcd) Timescale(number int, timescale string) error {
	// Validate that number is one of the allowed values
	if number != 1 && number != 10 && number != 100 {
		return fmt.Errorf("timescale number must be 1, 10, or 100, got %d", number)
	}

	if !v.sampleCalled {
		C.cxxrtl_vcd_timescale(v.vcd, C.int(number), getCStr(timescale, v.cstrHeap))
		return nil
	}
	return fmt.Errorf("timescale can only be set before the first call to sample")
}

func (v *Vcd) Destroy() {
	C.cxxrtl_vcd_destroy(v.vcd)
	for _, v := range v.cstrHeap {
		C.free(unsafe.Pointer(v))
	}
}

func NewCxxrtlVcd() *Vcd {
	return &Vcd{
		vcd:          C.cxxrtl_vcd_create(),
		sampleCalled: false,
		cstrHeap:     make(map[string]*C.char),
	}
}

// NewCxxrtl creates and initializes a new CXXRTL simulation instance.
// It sets up the necessary C data structures and returns a pointer to
// a fully initialized Cxxrtl object ready for simulation.
//
// Returns a pointer to the new Cxxrtl instance.
func NewCxxrtl() *Cxxrtl {
	cxx := &Cxxrtl{
		top:      C.cxxrtl_design_create(),
		cstrHeap: make(map[string]*C.char),
	}
	cxx.handle = C.cxxrtl_create(cxx.top)
	return cxx
}

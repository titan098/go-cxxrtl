YOSYS_BINDIR := $(shell yosys-config --bindir)
YOSYS_DATDIR := $(shell yosys-config --datdir)
YOSYS := $(YOSYS_BINDIR)/yosys

CXX := $(shell yosys-config --cxx)
CXXFLAGS := -g -O3 -fPIC --std=c++14 -I$(YOSYS_DATDIR)/include/backends/cxxrtl/runtime

BUILD_DIR := build
CXXRTL_SRC := $(YOSYS_DATDIR)/include/backends/cxxrtl/runtime/cxxrtl/capi
SRC_FILES := cxxrtl_capi.cc cxxrtl_capi_vcd.cc blinky.cc
SRCS := $(addprefix $(BUILD_DIR)/, $(SRC_FILES))
OBJS := $(SRCS:.cc=.o)

all: $(BUILD_DIR)/libcxxrtl.so

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/blinky.cc: blinky.v | $(BUILD_DIR)
	$(YOSYS) -p "read_verilog $<; write_cxxrtl $@"

$(BUILD_DIR)/%.cc: $(CXXRTL_SRC)/%.cc | $(BUILD_DIR)
	cp $< $@

$(BUILD_DIR)/%.o: $(BUILD_DIR)/%.cc | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) -c $< -o $@

$(BUILD_DIR)/libcxxrtl.so: $(OBJS)
	$(CXX) -shared -fPIC -rdynamic -o $@ $^

clean:
	rm -rf $(BUILD_DIR)

.PHONY: all clean

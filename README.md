# go-cxxrtl
A Go wrapper around cxxrtl-capi

## Usage

Using this library requires a working installation of [Yosys](https://github.com/YosysHQ/yosys) and CXXRTL. The easiest way is to install [oss-cad-suite](https://github.com/YosysHQ/oss-cad-suite-build), which should include everything you need. There are some predefined references to `/opt/oss-cad-suite/` however, these can be overridden at compilation time.

The library also depends on `cgo` to reference a shared library built from the C++ output from CXXRTL as well as their C bindings. A working C++ compiler is also required.

## Testing

There is a test suite that includes tests against a `blinky` implementation. The `Makefile` encapsulates the general build process for this project. Running the tests can be done by executing:

```bash
make test
```

## Generalised Workflow

There is a contrived "blinky" Verilog implementation in [blinky.v](blinky/blinky.v) as well as a corresponding Makefile that can be used in combination with this library. The general flow is as follows:

1. Generate the CXXRTL C++ code from your top-level Verilog file:
    ```bash
    yosys -p "read_verilog blinky.v; write_cxxrtl blinky.cc;"
    ```
2. Compile and generate a library from the generated C++ and the CXXRTL bindings (it is possible to use a static library if preferred)
    ```bash
    YOSYS_DATADIR=$(yosys-config --datdir)/include/backends/cxxrtl/runtime
    clang++ -g -O3 -fPIC --std=c++14 -I${YOSYS_DATADIR} -c ${YOSYS_DATADIR}/cxxrtl/capi/cxxrtl_capi.cc -o cxxrtl-capi.o
    clang++ -g -O3 -fPIC --std=c++14 -I${YOSYS_DATADIR} -c ${YOSYS_DATADIR}/cxxrtl/capi/cxxrtl_capi_vcd.cc -o cxxrtl-capi-vcd.o
    clang++ -g -O3 -fPIC --std=c++14 -I${YOSYS_DATADIR} -c blinky.cc -o blinky.o
    clang++ -shared -fPIC -rdynamic -o lib/libcxxrtl.so cxxrtl-capi.o cxxrtl-capi-vcd.o blinky.o
    ```
3. Build your Go application referencing the location of the built library:
    ```bash
    CGO_LDFLAGS=-L./lib go build
    ```
4. Execute your Go application referencing the location of the built library:
    ```bash
    LD_LIBRARY_PATH=./lib ./your-go-application
    ```

## References

- [Yosys](https://github.com/YosysHQ/yosys)
- [Yosys CXXRTL](https://github.com/YosysHQ/yosys/tree/main/backends/cxxrtl)


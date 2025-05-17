export CGO_LDFLAGS=-L./blinky/build/
export LD_LIBRARY_PATH=./blinky/build/

blinky:
	$(MAKE) -C blinky

test: blinky
	go test ./...

clean:
	rm -rf blinky/build

.PHONY: blinky test clean
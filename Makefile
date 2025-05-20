export CGO_LDFLAGS=-L$(PWD)/blinky/build/
export LD_LIBRARY_PATH=$(PWD)/blinky/build/

blinky:
	$(MAKE) -C blinky

test: blinky
	go test ./...

clean:
	rm -rf blinky/build

.PHONY: blinky test clean
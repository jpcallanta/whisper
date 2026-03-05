BINARY := whisper

.PHONY: all test build run install clean

all: build

test:
	go test ./...

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY) $(ARGS)

# Install to /usr/local/bin; run: sudo make install
install: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY) *.test *.out

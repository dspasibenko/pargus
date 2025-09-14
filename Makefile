.PHONY: build clean test install

# Build the C++ generator executable
build:
	go build -o build/pargus ./cmd/pargus

# Install the C++ generator to GOPATH/bin
install:
	go install ./cmd/pargus

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf build/

# Build and install
all: build install

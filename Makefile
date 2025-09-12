.PHONY: build clean test install

# Build the C++ generator executable
build:
	go build -o bin/pargus-cpp-generator ./cmd/pargus-cpp-generator

# Install the C++ generator to GOPATH/bin
install:
	go install ./cmd/pargus-cpp-generator

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build and install
all: build install

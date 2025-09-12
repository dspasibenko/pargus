# Pargus

A language and toolchain for describing embedded device registers and generating Arduino C++ code.

## Features

- **Pargus Language**: A domain-specific language for describing device registers, fields, and their properties
- **Arduino C++ Generator**: Generates Arduino-compatible C++ header files with register access functions
- **Trailing Comments Support**: Properly handles comments on the same line as field declarations
- **Optimized Code Generation**: Generates efficient code with simplified functions for unused field types

## Quick Start

### Building the Generator

```bash
make build
```

### Using the Generator

```bash
./bin/pargus-cpp-generator -n MyNamespace device.pa
```

### Example Pargus File

```pargus
device sensor

register Control(1) {
    // Enable sensor
    enable uint32{bit0: 0, mode: 1-3, high: 22-31};
    // Temperature reading
    temperature int16;
    
    // Data buffer
    data [4]uint8;
};

register Status(2):r {
    ready uint8; // Ready status
};
```

## Command Line Tool

The `pargus-cpp-generator` tool accepts the following parameters:

- `input.pa` - Input Pargus file (required)
- `-n, --namespace <namespace>` - C++ namespace name (required)
- `-o, --output <filename.h>` - Output header file (optional, defaults to input.h)

### Examples

```bash
# Generate with default output name
./bin/pargus-cpp-generator -n MyNamespace device.pa

# Generate with custom output name
./bin/pargus-cpp-generator -n MyNamespace -o my_device.h device.pa
```

## Development

### Running Tests

```bash
make test
```

### Building and Installing

```bash
make all
```

## License

[Add your license here]

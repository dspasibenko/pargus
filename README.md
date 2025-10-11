# Pargus

**Pargus** (Argus Protocol) is a simple Domain-Specific Language (DSL) and code generator for describing network communication protocols in the Argus system. It enables developers to define device APIs with registers and fields, then automatically generate type-safe code for multiple target languages.

## Overview

The Argus system consists of devices that communicate with each other over a network. Each device exposes registers that can be read from or written to via network requests. Pargus provides a clean, simple syntax to describe these registers and their fields, eliminating the need to manually write serialization/deserialization code.

## Wait, is it serialization?
Yes, it’s similar to Protobuf, but designed to be a much lighter and more compact way to send messages over the wire. It’s suitable for microcontrollers with very limited memory, where most of Protobuf’s features aren’t necessary. Pargus focuses on extremely compact serialization, where every byte matters, and simplicity is more important than flexibility. In essence, it’s a lightweight serialization protocol for the Argus microcontroller network.

## Key Features

- **Simple Syntax**: Define device APIs using an intuitive `.pa` file format
- **Type Safety**: Support for various primitive types (int8-64, uint8-64, float32/64) and complex types (arrays, bit fields, nested registers)
- **Read/Write Control**: Specify read-only, write-only, or read-write access for registers and fields
- **Code Generation**: Automatically generate code for multiple target languages:
  - **Go** - idiomatic Go structs with encoding/decoding methods
  - **Arduino C++** - embedded-friendly C++ code with minimal overhead
- **Bit Field Support**: Define and manipulate individual bits or bit ranges within integer fields
- **Variable-Length Arrays**: Support for dynamic arrays with sizes determined by other fields or bit masks

## Usage

```bash
# Build the generator
make build

# Generate code from a .pa file
./build/pargus -input device.pa -output-dir ./generated -lang go
./build/pargus -input device.pa -output-dir ./generated -lang arduino-cpp
```

## Specification

For the complete language specification and detailed examples, see [spec/pargus.md](spec/pargus.md).

## Example

```pargus
device argus-p

// Configuration register (read-write)
register Config(0) {
    mode uint8;
    enabled uint8;
}

// Status register (read-only)
register Status(1): r {
    counter int32;
    flags uint8{ready: 0, error: 1-3};
}
```



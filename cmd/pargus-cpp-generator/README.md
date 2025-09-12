# Pargus C++ Generator

A command-line tool for generating Arduino C++ header files from Pargus device descriptions.

## Usage

```bash
pargus-cpp-generator [options] input.pa
```

## Options

- `-n, -namespace <namespace>` - C++ namespace name (required)
- `-o, -output <filename.h>` - Output header file (default: input.h)
- `-help` - Show help

## Examples

Generate a header file with default name:
```bash
pargus-cpp-generator -n MyNamespace device.pa
```

Generate a header file with custom name:
```bash
pargus-cpp-generator -n MyNamespace -o my_device.h device.pa
```

## Building

```bash
make build
```

This will create `bin/pargus-cpp-generator` executable.

## Installation

```bash
make install
```

This will install the executable to `$GOPATH/bin/pargus-cpp-generator`.

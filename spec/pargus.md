# Pargus (Argus Protocol) Specification

This document describes the Pargus language - an extremely simple DSL for describing objects passed via network in the Argus system. The language is developed for describing APIs on various Argus devices with the purpose of generating simple messages for the target language.

## Introduction

- The Argus system includes one or more devices that may communicate with each other.
- Each device has its identifier - a device ID, or simply a unique name that describes the device.
- A device normally exposes one or multiple registers that may be read from or written to via network requests.
- Each register consists of zero or more fields with specified types.
- Read and write operations for the same register may contain different fields.

## The Spec

Pargus normally describes an API supported by a device that exposes the API.
A device API in Pargus is always described in a single file with the `.pa` extension. Multiple files are not supported.
The `.pa` file contains directives and comments. Comments start with the `//` sequence.

### device directive

Each `.pa` file should contain a first line with the keyword `device` followed by the device name.

For example, in `argus-p.pa`:

```
device argus-p
```

The file describes the API for "argus-p". Only one `device` directive is allowed per file.

### register directive

A register directive describes a register that can be read from or written to for the device.

The directive has the following form:

```
register RegisterName(0) {
    // fields that can be read
};
```

The register name is followed by a positive number in parentheses. No two registers may have the same register number for the device. The register number is mandatory and must be specified for each register.
After the register name, it may be followed by the specifier `r` (read only) or `w` (write only). If nothing is specified, the register may be read and written.

For example:

```
// read-write register
register RW(0) {
    field int16;
};

register ReadOnly(1): r {
    field int8;
};

register WriteOnly(2): w {
    // no fields are ok.
};
```

### Register constants
The register definition may contain constant definitions. The constant are always integer values declared with `const` word, for example:

```
register R(1) {
  const someValue = uint8(123);
}
```

### Register fields

Each field is described in the following form:

`<field_name>[:r|w] <field_type> [<options>];`

For example:

`counter int32;`

Like the register, any field may also have a specifier `r` or `w` which makes the field `read only` or `write only`.

If no specifier is provided, the field may be read and written:

```
register R(1) {
    read_only: r int8;
    write_only: w int8;
    read_write int8;
};
```

If the register has the specifier itself, all its fields will inherit the register specifier:

```
register ReadOnly(2): r {
    write_only: w int8; // error: the field cannot be write only because the register is read only
    read_only: r int8; // ok, the register is read only
    read_only2 int8; // ok, it is read only, the whole register is read only
};

register WriteOnly(3): w {
    write_only: w int8; // ok, the register is write only
    read_only: r int8; // error: the register is write only
    write_only2 int8; // ok, it is write only, the whole register is write only
};
```

#### Field types

The following simple types are supported:

- `int8`/`uint8`: signed/unsigned 1 byte field
- `int16`/`uint16`: signed/unsigned 2 bytes field
- `int32`/`uint32`: signed/unsigned 4 bytes field
- `int64`/`uint64`: signed/unsigned 8 bytes field
- `float32`: 4 bytes real number
- `float64`: 8 bytes real number

Complex types:

- `[x]<type>` - fixed-size array of x elements, where x is a constant like `5`. Example: `[5]int8`
- `[field_or_bitmask_ref]<type>` - variable-length array, where the size is determined by the value of the referenced field. Two important notes:
  1. The field must be declared before the variable array
  2. The field can be a bit mask (just 1 or few bits long). In this case, the reference name will be `<fieldname_bitmaskname>`
- `uint<N>{bit_name: bit_pos, ...}` - a bit field. After the bit-field name (colon), follows either the bit number or the bit range for the field
- `<RegisterName>` - a reference to another register defined in the same file. This creates a field of the register's struct type. The referenced register must exist in the device definition. **Important:** Circular dependencies are not allowed (e.g., if register A contains a field of type B, then register B cannot contain a field of type A, directly or indirectly).

Example:

```
register Config(1) {
    mode uint8;
    enabled uint8;
}

register R1(2) {
    some_int int32;
    fixed_size_array [3]int16;
    string [some_int]uint8; // the size of the field will be in some_int
    
    bit_field uint8{bit0: 0, bit57: 5-7};
    another_buf [bit_field_bit57]float32; // variable array with the size encoded into the bit field
    
    config Config; // field with type of Config register
}
```

**Note:** Bit fields can only be unsigned integer types. The number of bits cannot exceed the size of the bit-field type.

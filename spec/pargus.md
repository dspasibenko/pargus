# Pargus (Argus Protocol) specification
This document describes the Pargus language - an extremely simple DSL for describing objects passed via network in the Argus system. The language is developed for describing APIs on various Argus devices with the purpose of generating simple messages for the target language. 

## Introduction
The Argus system includes one or many devices that may communicate with each other. 

Each device has its identifier - a device ID, or simply a unique name that describes the device.

A device normally exposes one or multiple registers that may be read from or written to via network requests. 

Each register consists of zero or more fields with specified types. 

Read and write operations for the same register may contain different fields.

## The spec
Pargus normally describes an API supported by a device that exposes the API. 

A device API in Pargus is always described in a single file with the `.pa` extension. Multiple files are not supported. 

The `.pa` file contains directives and comments. Comments start with the `//` sequence. 

### device directive
Each `.pa` file should contain a first line with the keyword `device` followed by the device name. 

For example, in argus-p.pa:
```
device argus-p
```
The file describes the API for "argus-p". Only one directive is allowed.

### register directive
A register directive describes a register that can be read from or written to for the device. 

The directive has the following form:
```
register RegisterName(0) {
    // fields that can be read
};
```
The register name is followed by a positive number in parentheses. No two registers may have the same register number for the device. The register number is mandatory and must be specified for each register. 

After the register name it may follow the specifier `r` which means `read only` and `w` which means `write only`, if nothing is specified, this means the register may be read and written. For example:
```
// read-write register
register RW(0) {
    field int16;
};

register ReadOnly(1): r {
    field int8;
};

register WriteOnly(2):w {
    // no fields are ok.
};

```

### Register fields
Each field is described in the following form:
`<field_name>[:r|w] <field_type> [<options>];`

For example:
`counter int32;`

As the register, any field may also have a specifier `r` or `w` which makes the field `read only` or `write only` 

if no specifier the field may be rw:
```
register R(1) {
    read_only:r int8;
    write_only:w int8;
    read_write int8;
};
 ```

 If the register has the specifier itself all its fields will have the register specifier:

 ```
 register ReadOnly(2): r {
    write_only: w int8; // error the field cannot be write only casue the register is read only
    read_only: r int8; // ok, the regeister is read only
    read_only2 int8; // ok. it is read only, the whole register is read only
 };

 register WriteOnly(3): w {
    write_only: w int8; // ok, the register is write only
    read_only: r int8; // error, the regeister is write only
    write_only2 int8; // ok. it is write only, the whole register is write only
 };

 ```

#### Field types
The following simple types are supported:
- int8/uint8: signed/unsigned 1 byte field
- int16/uint16: signed/unsigned 2 bytes field
- int32/uint32: signed/unsigned 4 bytes field
- int64/uint64: signed/unsigned 8 bytes field
- float32: 4 bytes real number
- float64: 8 bytes real number

Complex types:
- [x]int8 - fixed-size array of x bytes, where x is a constant like `5`
- [<base type - unsigned integer>]int8 - variable-length array, where the size will be 1 byte in length and passed with the field as `array_field_name_size`.
- uint8{bit_X : 0, three_bits_val: 1-3 ...} - a bit field. In the bit field, after the bit-field name (colon), follows either the bit number or the bit range for the field. 

Example:
```
register R1(2) {
    some_int int32;
    fixed_size_array [3]int16;
    string [uint8]uint8; // the size of the field will be in string_size
    // we cannot declare field string_size here, because the size for the string array is already generated
    bit_field uint8{bit0: 0, bit57: 5-7};
}
```

Bit fields can only be unsigned integer types. The number of bits cannot exceed the size of the bit-field type. 
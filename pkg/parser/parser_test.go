package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_DeviceDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Device
		hasError bool
	}{
		{
			name:  "simple device declaration",
			input: "device argus_p",
			expected: &Device{
				Name:      "argus_p",
				Registers: nil,
			},
			hasError: false,
		},
		{
			name:  "device with underscore",
			input: "device argus_device_1",
			expected: &Device{
				Name:      "argus_device_1",
				Registers: nil,
			},
			hasError: false,
		},
		{
			name:  "device with numbers",
			input: "device device123",
			expected: &Device{
				Name:      "device123",
				Registers: nil,
			},
			hasError: false,
		},
		{
			name:     "missing device keyword",
			input:    "argus-p",
			expected: nil,
			hasError: true,
		},
		{
			name:     "missing device name",
			input:    "device",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid device name with spaces",
			input:    "device argus p",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				if tt.expected.Registers == nil {
					assert.Nil(t, result.Registers)
				} else {
					assert.Equal(t, tt.expected.Registers, result.Registers)
				}
			}
		})
	}
}

func TestParse_RegisterDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Device
		hasError bool
	}{
		{
			name:  "simple register without fields",
			input: "device test\nregister R1(1) {\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields:    []*Field{},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "register with simple field",
			input: "device test\nregister R1(1) {\n    counter int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "counter",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "multiple registers",
			input: "device test\nregister R1(1) {\n    field1 int32;\n};\nregister R2(2) {\n    field2 uint8;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field1",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
					{
						Name:      "R2",
						Number:    2,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field2",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "uint8"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:     "register without closing brace",
			input:    "device test\nregister R1(1) {\n    field1 int32;",
			expected: nil,
			hasError: true,
		},
		{
			name:     "register without semicolon",
			input:    "device test\nregister R1(1) {\n    field1 int32;\n}",
			expected: nil,
			hasError: true,
		},
		{
			name:     "register without number",
			input:    "device test\nregister R1 {\n    field1 int32;\n};",
			expected: nil,
			hasError: true,
		},
		{
			name:     "duplicate register numbers",
			input:    "device test\nregister R1(1) {\n    field1 int32;\n};\nregister R2(1) {\n    field2 uint8;\n};",
			expected: nil,
			hasError: true,
		},
		{
			name:  "register with read-only specifier",
			input: "device test\nregister R1(1): r {\n    field1 int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: stringPtr("r"),
						Fields: []*Field{
							{
								Name:      "field1",
								Specifier: stringPtr("r"), // inherits from register
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "register with write-only specifier",
			input: "device test\nregister R1(1): w {\n    field1 int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: stringPtr("w"),
						Fields: []*Field{
							{
								Name:      "field1",
								Specifier: stringPtr("w"), // inherits from register
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "register without specifier (read-write)",
			input: "device test\nregister R1(1) {\n    field1 int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field1",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, len(tt.expected.Registers), len(result.Registers))

				for i, expectedReg := range tt.expected.Registers {
					actualReg := result.Registers[i]
					assert.Equal(t, expectedReg.Name, actualReg.Name)
					assert.Equal(t, expectedReg.Number, actualReg.Number)
					if expectedReg.Specifier != nil {
						require.NotNil(t, actualReg.Specifier)
						assert.Equal(t, *expectedReg.Specifier, *actualReg.Specifier)
					} else {
						assert.Nil(t, actualReg.Specifier)
					}
					assert.Equal(t, len(expectedReg.Fields), len(actualReg.Fields))

					for j, expectedField := range expectedReg.Fields {
						actualField := actualReg.Fields[j]
						assert.Equal(t, expectedField.Name, actualField.Name)
						if expectedField.Specifier != nil {
							require.NotNil(t, actualField.Specifier)
							assert.Equal(t, *expectedField.Specifier, *actualField.Specifier)
						} else {
							assert.Nil(t, actualField.Specifier)
						}
						assert.Equal(t, expectedField.Type, actualField.Type)
					}
				}
			}
		})
	}
}

func TestParse_SimpleTypes(t *testing.T) {
	simpleTypes := []string{
		"int8", "uint8", "int16", "uint16",
		"int32", "uint32", "int64", "uint64",
		"float32", "float64",
	}

	for _, typeName := range simpleTypes {
		t.Run("simple_type_"+typeName, func(t *testing.T) {
			input := "device test\nregister R1(1) {\n    field " + typeName + ";\n};"

			result, err := Parse(input)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Registers, 1)
			require.Len(t, result.Registers[0].Fields, 1)

			field := result.Registers[0].Fields[0]
			assert.Equal(t, "field", field.Name)

			require.NotNil(t, field.Type)
			require.NotNil(t, field.Type.Simple)
			assert.Equal(t, typeName, field.Type.Simple.Name)
		})
	}
}

func TestParse_ArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ArrayType
		hasError bool
	}{
		{
			name:  "fixed size array with number",
			input: "device test\nregister R1(1) {\n    arr [5]int32;\n};",
			expected: &ArrayType{
				Size: "5",
				Element: &TypeUnion{
					Simple: &SimpleType{Name: "int32"},
				},
			},
			hasError: false,
		},
		{
			name:  "variable length array with uint8 size",
			input: "device test\nregister R1(1) {\n    str [uint8]uint8;\n};",
			expected: &ArrayType{
				Size: "uint8",
				Element: &TypeUnion{
					Simple: &SimpleType{Name: "uint8"},
				},
			},
			hasError: false,
		},
		{
			name:  "nested array",
			input: "device test\nregister R1(1) {\n    matrix [3][2]int16;\n};",
			expected: &ArrayType{
				Size: "3",
				Element: &TypeUnion{
					Array: &ArrayType{
						Size: "2",
						Element: &TypeUnion{
							Simple: &SimpleType{Name: "int16"},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:     "array without closing bracket",
			input:    "device test\nregister R1(1) {\n    arr [5int32;\n};",
			expected: nil,
			hasError: true,
		},
		{
			name:     "array without size",
			input:    "device test\nregister R1(1) {\n    arr []int32;\n};",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Registers, 1)
				require.Len(t, result.Registers[0].Fields, 1)

				field := result.Registers[0].Fields[0]
				require.NotNil(t, field.Type)

				if tt.expected.Element.Array != nil {
					// Nested array case
					require.NotNil(t, field.Type.Array)
					assert.Equal(t, tt.expected.Size, field.Type.Array.Size)
					assert.Equal(t, tt.expected.Element.Array.Size, field.Type.Array.Element.Array.Size)
					assert.Equal(t, tt.expected.Element.Array.Element.Simple.Name, field.Type.Array.Element.Array.Element.Simple.Name)
				} else {
					// Simple array case
					require.NotNil(t, field.Type.Array)
					assert.Equal(t, tt.expected.Size, field.Type.Array.Size)
					assert.Equal(t, tt.expected.Element.Simple.Name, field.Type.Array.Element.Simple.Name)
				}
			}
		})
	}
}

func TestParse_BitFieldTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *BitField
		hasError bool
	}{
		{
			name:  "simple bit field with single bit",
			input: "device test\nregister R1(1) {\n    flags uint8{bit0: 0};\n};",
			expected: &BitField{
				Base: "uint8",
				Bits: []*BitMember{
					{Name: "bit0", Start: 0, End: nil},
				},
			},
			hasError: false,
		},
		{
			name:  "bit field with bit range",
			input: "device test\nregister R1(1) {\n    flags uint8{three_bits: 1-3};\n};",
			expected: &BitField{
				Base: "uint8",
				Bits: []*BitMember{
					{Name: "three_bits", Start: 1, End: intPtr(3)},
				},
			},
			hasError: false,
		},
		{
			name:  "bit field with multiple members",
			input: "device test\nregister R1(1) {\n    flags uint8{bit0: 0, three_bits: 1-3, bit7: 7};\n};",
			expected: &BitField{
				Base: "uint8",
				Bits: []*BitMember{
					{Name: "bit0", Start: 0, End: nil},
					{Name: "three_bits", Start: 1, End: intPtr(3)},
					{Name: "bit7", Start: 7, End: nil},
				},
			},
			hasError: false,
		},
		{
			name:  "bit field with uint16 base",
			input: "device test\nregister R1(1) {\n    flags uint16{high_bits: 8-15};\n};",
			expected: &BitField{
				Base: "uint16",
				Bits: []*BitMember{
					{Name: "high_bits", Start: 8, End: intPtr(15)},
				},
			},
			hasError: false,
		},
		{
			name:     "bit field without closing brace",
			input:    "device test\nregister R1(1) {\n    flags uint8{bit0: 0;\n};",
			expected: nil,
			hasError: true,
		},
		{
			name:     "bit field with invalid base type",
			input:    "device test\nregister R1(1) {\n    flags int8{bit0: 0};\n};",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Registers, 1)
				require.Len(t, result.Registers[0].Fields, 1)

				field := result.Registers[0].Fields[0]
				require.NotNil(t, field.Type)
				require.NotNil(t, field.Type.Bitfield)

				bitField := field.Type.Bitfield
				assert.Equal(t, tt.expected.Base, bitField.Base)
				assert.Equal(t, len(tt.expected.Bits), len(bitField.Bits))

				for i, expectedBit := range tt.expected.Bits {
					actualBit := bitField.Bits[i]
					assert.Equal(t, expectedBit.Name, actualBit.Name)
					assert.Equal(t, expectedBit.Start, actualBit.Start)
					if expectedBit.End != nil {
						require.NotNil(t, actualBit.End)
						assert.Equal(t, *expectedBit.End, *actualBit.End)
					} else {
						assert.Nil(t, actualBit.End)
					}
				}
			}
		})
	}
}

func TestParse_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Device
		hasError bool
	}{
		{
			name: "device with comments",
			input: `device test
// This is a comment
register R1(1) {
    // Field comment
    counter int32;
    // Another comment
};`,
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "counter",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "multiple line comments",
			input: `device test
// Comment before register
register R1(1) 

{
    // Comment before field
    field1 int32;
    // Comment between fields


    field2 uint8;
    // Comment after last field
};`,
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field1",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
							{
								Name:      "field2",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "uint8"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, len(tt.expected.Registers), len(result.Registers))

				for i, expectedReg := range tt.expected.Registers {
					actualReg := result.Registers[i]
					assert.Equal(t, expectedReg.Name, actualReg.Name)
					assert.Equal(t, expectedReg.Number, actualReg.Number)
					if expectedReg.Specifier != nil {
						require.NotNil(t, actualReg.Specifier)
						assert.Equal(t, *expectedReg.Specifier, *actualReg.Specifier)
					} else {
						assert.Nil(t, actualReg.Specifier)
					}
					assert.Equal(t, len(expectedReg.Fields), len(actualReg.Fields))

					for j, expectedField := range expectedReg.Fields {
						actualField := actualReg.Fields[j]
						assert.Equal(t, expectedField.Name, actualField.Name)
						if expectedField.Specifier != nil {
							require.NotNil(t, actualField.Specifier)
							assert.Equal(t, *expectedField.Specifier, *actualField.Specifier)
						} else {
							assert.Nil(t, actualField.Specifier)
						}
						assert.Equal(t, expectedField.Type, actualField.Type)
					}
				}
			}
		})
	}
}

func TestParse_ComplexExample(t *testing.T) {
	// Test the example from the specification
	input := `device argus_p
register R1(2) {
    some_int int32;
    fixed_size_array [3]int16;
    string [uint8]uint8; // the size of the field will be in string_size
    // we cannot declare field string_size here, because the size for the string array is already generated
    bit_field uint8{bit0: 0, bit57: 5-7};
};`

	result, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "argus_p", result.Name)
	require.Len(t, result.Registers, 1)

	register := result.Registers[0]
	assert.Equal(t, "R1", register.Name)
	assert.Equal(t, 2, register.Number)
	require.Len(t, register.Fields, 4)

	// Check some_int field
	field1 := register.Fields[0]
	assert.Equal(t, "some_int", field1.Name)
	require.NotNil(t, field1.Type)
	require.NotNil(t, field1.Type.Simple)
	assert.Equal(t, "int32", field1.Type.Simple.Name)

	// Check fixed_size_array field
	field2 := register.Fields[1]
	assert.Equal(t, "fixed_size_array", field2.Name)
	require.NotNil(t, field2.Type)
	require.NotNil(t, field2.Type.Array)
	assert.Equal(t, "3", field2.Type.Array.Size)
	assert.Equal(t, "int16", field2.Type.Array.Element.Simple.Name)

	// Check string field
	field3 := register.Fields[2]
	assert.Equal(t, "string", field3.Name)
	require.NotNil(t, field3.Type)
	require.NotNil(t, field3.Type.Array)
	assert.Equal(t, "uint8", field3.Type.Array.Size)
	assert.Equal(t, "uint8", field3.Type.Array.Element.Simple.Name)

	// Check bit_field
	field4 := register.Fields[3]
	assert.Equal(t, "bit_field", field4.Name)
	require.NotNil(t, field4.Type)
	require.NotNil(t, field4.Type.Bitfield)
	assert.Equal(t, "uint8", field4.Type.Bitfield.Base)
	require.Len(t, field4.Type.Bitfield.Bits, 2)

	bit0 := field4.Type.Bitfield.Bits[0]
	assert.Equal(t, "bit0", bit0.Name)
	assert.Equal(t, 0, bit0.Start)
	assert.Nil(t, bit0.End)

	bit57 := field4.Type.Bitfield.Bits[1]
	assert.Equal(t, "bit57", bit57.Name)
	assert.Equal(t, 5, bit57.Start)
	require.NotNil(t, bit57.End)
	assert.Equal(t, 7, *bit57.End)
}

func TestParse_SpecificationExample(t *testing.T) {
	// Test the example from the updated specification with specifiers
	input := `device argus_p
// read-write register
register RW(0) {
    field int16;
};

register ReadOnly(1): r {
    field int8;
};

register WriteOnly(2): w {
    // no fields are ok.
};`

	result, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "argus_p", result.Name)
	require.Len(t, result.Registers, 3)

	// Check RW register (read-write, no specifier)
	rwRegister := result.Registers[0]
	assert.Equal(t, "RW", rwRegister.Name)
	assert.Equal(t, 0, rwRegister.Number)
	assert.Nil(t, rwRegister.Specifier)
	require.Len(t, rwRegister.Fields, 1)
	assert.Equal(t, "field", rwRegister.Fields[0].Name)
	assert.Equal(t, "int16", rwRegister.Fields[0].Type.Simple.Name)

	// Check ReadOnly register (read-only)
	readOnlyRegister := result.Registers[1]
	assert.Equal(t, "ReadOnly", readOnlyRegister.Name)
	assert.Equal(t, 1, readOnlyRegister.Number)
	require.NotNil(t, readOnlyRegister.Specifier)
	assert.Equal(t, "r", *readOnlyRegister.Specifier)
	require.Len(t, readOnlyRegister.Fields, 1)
	assert.Equal(t, "field", readOnlyRegister.Fields[0].Name)
	assert.Equal(t, "int8", readOnlyRegister.Fields[0].Type.Simple.Name)

	// Check WriteOnly register (write-only)
	writeOnlyRegister := result.Registers[2]
	assert.Equal(t, "WriteOnly", writeOnlyRegister.Name)
	assert.Equal(t, 2, writeOnlyRegister.Number)
	require.NotNil(t, writeOnlyRegister.Specifier)
	assert.Equal(t, "w", *writeOnlyRegister.Specifier)
	require.Len(t, writeOnlyRegister.Fields, 0)
}

func TestParse_SpecificationFieldSpecifiersExample(t *testing.T) {
	// Test the example from the updated specification with field specifiers
	input := `device test
register R(1) {
    read_only:r int8;
    write_only:w int8;
    read_write int8;
};

register ReadOnly(2): r {
    read_only: r int8;
    read_only2 int8;
};

register WriteOnly(3): w {
    write_only: w int8;
    write_only2 int8;
};`

	result, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test", result.Name)
	require.Len(t, result.Registers, 3)

	// Check R register (read-write, no register specifier)
	rRegister := result.Registers[0]
	assert.Equal(t, "R", rRegister.Name)
	assert.Equal(t, 1, rRegister.Number)
	assert.Nil(t, rRegister.Specifier)
	require.Len(t, rRegister.Fields, 3)

	// Check read_only field
	readOnlyField := rRegister.Fields[0]
	assert.Equal(t, "read_only", readOnlyField.Name)
	require.NotNil(t, readOnlyField.Specifier)
	assert.Equal(t, "r", *readOnlyField.Specifier)

	// Check write_only field
	writeOnlyField := rRegister.Fields[1]
	assert.Equal(t, "write_only", writeOnlyField.Name)
	require.NotNil(t, writeOnlyField.Specifier)
	assert.Equal(t, "w", *writeOnlyField.Specifier)

	// Check read_write field
	readWriteField := rRegister.Fields[2]
	assert.Equal(t, "read_write", readWriteField.Name)
	assert.Nil(t, readWriteField.Specifier)

	// Check ReadOnly register
	readOnlyRegister := result.Registers[1]
	assert.Equal(t, "ReadOnly", readOnlyRegister.Name)
	assert.Equal(t, 2, readOnlyRegister.Number)
	require.NotNil(t, readOnlyRegister.Specifier)
	assert.Equal(t, "r", *readOnlyRegister.Specifier)
	require.Len(t, readOnlyRegister.Fields, 2)

	// Check read_only field in read-only register
	readOnlyField2 := readOnlyRegister.Fields[0]
	assert.Equal(t, "read_only", readOnlyField2.Name)
	require.NotNil(t, readOnlyField2.Specifier)
	assert.Equal(t, "r", *readOnlyField2.Specifier)

	// Check read_only2 field (inherits register specifier)
	readOnlyField3 := readOnlyRegister.Fields[1]
	assert.Equal(t, "read_only2", readOnlyField3.Name)
	require.NotNil(t, readOnlyField3.Specifier)
	assert.Equal(t, "r", *readOnlyField3.Specifier)

	// Check WriteOnly register
	writeOnlyRegister := result.Registers[2]
	assert.Equal(t, "WriteOnly", writeOnlyRegister.Name)
	assert.Equal(t, 3, writeOnlyRegister.Number)
	require.NotNil(t, writeOnlyRegister.Specifier)
	assert.Equal(t, "w", *writeOnlyRegister.Specifier)
	require.Len(t, writeOnlyRegister.Fields, 2)

	// Check write_only field in write-only register
	writeOnlyField2 := writeOnlyRegister.Fields[0]
	assert.Equal(t, "write_only", writeOnlyField2.Name)
	require.NotNil(t, writeOnlyField2.Specifier)
	assert.Equal(t, "w", *writeOnlyField2.Specifier)

	// Check write_only2 field (inherits register specifier)
	writeOnlyField3 := writeOnlyRegister.Fields[1]
	assert.Equal(t, "write_only2", writeOnlyField3.Name)
	require.NotNil(t, writeOnlyField3.Specifier)
	assert.Equal(t, "w", *writeOnlyField3.Specifier)
}

func TestParse_FieldSpecifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Device
		hasError bool
	}{
		{
			name:  "field with read-only specifier",
			input: "device test\nregister R1(1) {\n    field:r int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field",
								Specifier: stringPtr("r"),
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "field with write-only specifier",
			input: "device test\nregister R1(1) {\n    field:w int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field",
								Specifier: stringPtr("w"),
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "field without specifier",
			input: "device test\nregister R1(1) {\n    field int32;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "field",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name:  "multiple fields with different specifiers",
			input: "device test\nregister R1(1) {\n    read_field:r int32;\n    write_field:w uint8;\n    rw_field int16;\n};",
			expected: &Device{
				Name: "test",
				Registers: []*Register{
					{
						Name:      "R1",
						Number:    1,
						Specifier: nil,
						Fields: []*Field{
							{
								Name:      "read_field",
								Specifier: stringPtr("r"),
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int32"},
								},
							},
							{
								Name:      "write_field",
								Specifier: stringPtr("w"),
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "uint8"},
								},
							},
							{
								Name:      "rw_field",
								Specifier: nil,
								Type: &TypeUnion{
									Simple: &SimpleType{Name: "int16"},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, len(tt.expected.Registers), len(result.Registers))

				for i, expectedReg := range tt.expected.Registers {
					actualReg := result.Registers[i]
					assert.Equal(t, expectedReg.Name, actualReg.Name)
					assert.Equal(t, expectedReg.Number, actualReg.Number)
					if expectedReg.Specifier != nil {
						require.NotNil(t, actualReg.Specifier)
						assert.Equal(t, *expectedReg.Specifier, *actualReg.Specifier)
					} else {
						assert.Nil(t, actualReg.Specifier)
					}
					assert.Equal(t, len(expectedReg.Fields), len(actualReg.Fields))

					for j, expectedField := range expectedReg.Fields {
						actualField := actualReg.Fields[j]
						assert.Equal(t, expectedField.Name, actualField.Name)
						if expectedField.Specifier != nil {
							require.NotNil(t, actualField.Specifier)
							assert.Equal(t, *expectedField.Specifier, *actualField.Specifier)
						} else {
							assert.Nil(t, actualField.Specifier)
						}
						assert.Equal(t, expectedField.Type, actualField.Type)
					}
				}
			}
		})
	}
}

func TestParse_FieldSpecifierValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "read-only register with write-only field",
			input: "device test\nregister ReadOnly(1): r {\n    field:w int32;\n};",
		},
		{
			name:  "write-only register with read-only field",
			input: "device test\nregister WriteOnly(1): w {\n    field:r int32;\n};",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestParse_ValidFieldSpecifierCombinations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "read-only register with read-only field",
			input: "device test\nregister ReadOnly(1): r {\n    field:r int32;\n};",
		},
		{
			name:  "read-only register with no field specifier",
			input: "device test\nregister ReadOnly(1): r {\n    field int32;\n};",
		},
		{
			name:  "write-only register with write-only field",
			input: "device test\nregister WriteOnly(1): w {\n    field:w int32;\n};",
		},
		{
			name:  "write-only register with no field specifier",
			input: "device test\nregister WriteOnly(1): w {\n    field int32;\n};",
		},
		{
			name:  "read-write register with any field specifiers",
			input: "device test\nregister RW(1) {\n    read_field:r int32;\n    write_field:w uint8;\n    rw_field int16;\n};",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestParse_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "invalid syntax - missing semicolon",
			input: "device test\nregister R1(1) {\n    field int32\n};",
		},
		{
			name:  "invalid syntax - wrong keyword",
			input: "devices test",
		},
		{
			name:  "invalid syntax - malformed register",
			input: "device test\nregister {\n    field int32;\n};",
		},
		{
			name:  "invalid syntax - malformed field",
			input: "device test\nregister R1(1) {\n    int32;\n};",
		},
		{
			name:  "invalid syntax - unknown type",
			input: "device test\nregister R1(1) {\n    field unknown_type;\n};",
		},
		{
			name:  "invalid syntax - malformed array",
			input: "device test\nregister R1(1) {\n    field [int32;\n};",
		},
		{
			name:  "invalid syntax - malformed bitfield",
			input: "device test\nregister R1(1) {\n    field uint8{bit0 0};\n};",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

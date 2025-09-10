package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBitFields_ValidBitFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "valid uint8 bit field",
			input: `device test
register R(0) {
    flags uint8 {bit0: 0, bit1: 1, bit7: 7};
};`,
		},
		{
			name: "valid uint16 bit field with ranges",
			input: `device test
register R(0) {
    flags uint16 {bit0: 0, bit57: 5-7, bit15: 15};
};`,
		},
		{
			name: "valid uint32 bit field",
			input: `device test
register R(0) {
    flags uint32 {bit0: 0, bit31: 31, bit1631: 16-31};
};`,
		},
		{
			name: "valid uint64 bit field",
			input: `device test
register R(0) {
    flags uint64 {bit0: 0, bit63: 63, bit3263: 32-63};
};`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			device, err := Parse(test.input)
			require.NoError(t, err)
			assert.NotNil(t, device)
		})
	}
}

func TestValidateBitFields_InvalidBaseTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "signed int8 bit field",
			input: `device test
register R(0) {
    flags int8 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
		{
			name: "signed int16 bit field",
			input: `device test
register R(0) {
    flags int16 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
		{
			name: "signed int32 bit field",
			input: `device test
register R(0) {
    flags int32 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
		{
			name: "signed int64 bit field",
			input: `device test
register R(0) {
    flags int64 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
		{
			name: "float32 bit field",
			input: `device test
register R(0) {
    flags float32 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
		{
			name: "float64 bit field",
			input: `device test
register R(0) {
    flags float64 {bit0: 0};
};`,
			expectedErr: "unexpected token \"{\" (expected \";\")",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedErr)
		})
	}
}

func TestValidateBitFields_InvalidBitRanges(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "uint8 bit exceeds size",
			input: `device test
register R(0) {
    flags uint8 {bit8: 8};
};`,
			expectedErr: "bit field 'flags' in register 'R': bit range 8-8 exceeds size of base type 'uint8' (8 bits)",
		},
		{
			name: "uint16 bit exceeds size",
			input: `device test
register R(0) {
    flags uint16 {bit16: 16};
};`,
			expectedErr: "bit field 'flags' in register 'R': bit range 16-16 exceeds size of base type 'uint16' (16 bits)",
		},
		{
			name: "uint32 bit exceeds size",
			input: `device test
register R(0) {
    flags uint32 {bit32: 32};
};`,
			expectedErr: "bit field 'flags' in register 'R': bit range 32-32 exceeds size of base type 'uint32' (32 bits)",
		},
		{
			name: "uint64 bit exceeds size",
			input: `device test
register R(0) {
    flags uint64 {bit64: 64};
};`,
			expectedErr: "bit field 'flags' in register 'R': bit range 64-64 exceeds size of base type 'uint64' (64 bits)",
		},
		// Note: Negative numbers cannot be parsed by the grammar, so we skip this test
		// The parser will fail before reaching validation
		{
			name: "start bit greater than end bit",
			input: `device test
register R(0) {
    flags uint8 {bit_range: 5-3};
};`,
			expectedErr: "bit field 'flags' in register 'R': start bit 5 cannot be greater than end bit 3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedErr)
		})
	}
}

func TestIsUnsignedType(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"uint8", true},
		{"uint16", true},
		{"uint32", true},
		{"uint64", true},
		{"int8", false},
		{"int16", false},
		{"int32", false},
		{"int64", false},
		{"float32", false},
		{"float64", false},
		{"unknown", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isUnsignedType(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetTypeSizeInBits(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"uint8", 8},
		{"uint16", 16},
		{"uint32", 32},
		{"uint64", 64},
		{"int8", 0},
		{"int16", 0},
		{"int32", 0},
		{"int64", 0},
		{"float32", 0},
		{"float64", 0},
		{"unknown", 0},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := getTypeSizeInBits(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

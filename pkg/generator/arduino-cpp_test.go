package generator

import (
	"testing"

	"github.com/dspasibenko/pargus/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_SimpleDevice(t *testing.T) {
	// Parse a simple device specification
	input := `device test_device
register Status(0) {
    temperature int16;
    pressure uint32;
};
register Config(1): w {
    threshold int8;
};`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	// Generate C++ code
	code, err := Generate(device, "TestNamespace")
	require.NoError(t, err)

	// Verify the generated code contains expected elements
	assert.Contains(t, code, "#include <Arduino.h>")
	assert.Contains(t, code, "#include \"bigendian.h\"")
	assert.Contains(t, code, "namespace TestNamespace {")
	assert.Contains(t, code, "struct Status {")
	assert.Contains(t, code, "int16_t temperature;")
	assert.Contains(t, code, "uint32_t pressure;")
	assert.Contains(t, code, "struct Config {")
	assert.Contains(t, code, "int8_t threshold;")
	assert.Contains(t, code, "static constexpr int StatusID = 0;")
	assert.Contains(t, code, "static constexpr int ConfigID = 1;")
	assert.Contains(t, code, "} // namespace TestNamespace")
}

func TestGenerate_ArrayTypes(t *testing.T) {
	input := `device test_device
register Data(0) {
    fixed_array [3]int16;
    var_array [uint8]uint8;
};`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	code, err := Generate(device, "TestNamespace")
	require.NoError(t, err)

	assert.Contains(t, code, "int16_t fixed_array[3];")
	assert.Contains(t, code, "uint8_t* var_array;")
}

func TestGenerate_BitField(t *testing.T) {
	input := `device test_device
register Flags(0) {
    control uint8{bit0: 0, bit57: 5-7};
};`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	code, err := Generate(device, "TestNamespace")
	require.NoError(t, err)

	assert.Contains(t, code, "uint8_t control;")
	assert.Contains(t, code, "static constexpr uint8_t bit0_bm = 0x1; // bit 0")
	assert.Contains(t, code, "static constexpr uint8_t bit57_bm = 0xE0; // bits 5-7")
}

func TestCppType_SimpleTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int8", "int8_t"},
		{"uint8", "uint8_t"},
		{"int16", "int16_t"},
		{"uint16", "uint16_t"},
		{"int32", "int32_t"},
		{"uint32", "uint32_t"},
		{"int64", "int64_t"},
		{"uint64", "uint64_t"},
		{"float32", "float"},
		{"float64", "double"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			simpleType := &parser.SimpleType{Name: test.input}
			result := cppType(simpleType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCppType_ArrayTypes(t *testing.T) {
	// Fixed-size array
	arrayType := &parser.ArrayType{
		Size:    "5",
		Element: &parser.TypeUnion{Simple: &parser.SimpleType{Name: "int16"}},
	}
	result := cppType(arrayType)
	assert.Equal(t, "int16_t", result) // cppType now returns just the element type

	// Variable-size array
	varArrayType := &parser.ArrayType{
		Size:    "size",
		Element: &parser.TypeUnion{Simple: &parser.SimpleType{Name: "uint8"}},
	}
	result = cppType(varArrayType)
	assert.Equal(t, "uint8_t*", result)
}

func TestGenerate_ComprehensiveExample(t *testing.T) {
	// Comprehensive Pargus specification with all possible features
	pargusSpec := `// Comprehensive test device with all features
device comprehensive_test
// Read-write register with mixed field types
register RW(0) {
    // Read-write field
    rw_field1 uint8;
    // Read-write field
    rw_field2 int16;
    // Read-only field
    read_field1 uint32;
    // Read-only field
    read_field2 float32;
    // Write-only field
    write_field1 uint8;
    // Write-only field
    write_field2 int64;
};

// Status register with detailed documentation
//
// The Status register contains device status information:
// - Bit 0: Device ready flag
// - Bits 1-3: Error code (0-7)
// - Bit 4: Power on indicator
// - Bits 5-7: Reserved for future use
//
// Error codes:
// 0 = No error
// 1 = Communication error
// 2 = Sensor error
// 3 = Power error
// 4-7 = Reserved
register Status(1): r {
    // Device status byte
    status uint8{ready: 0, power_on: 4};
    // Operation counter
    counter int32;
    // Configuration flags
    flags uint8{bit0: 0, bit15: 1};
};

// Command register for device control
register Command(2): w {
    // Command type
    cmd_type uint16;
    // Command value
    cmd_value int8;
    // Command configuration
    config uint8{enable: 0, priority: 1-2, mode: 3-4};
};

// Data register with arrays
register Data(3) {
    // Fixed size array
    samples int16[10];
    // Variable length array
    buffer uint8[size];
    // Simple field
    value float32;
};`

	// Expected C++ output
	expectedCpp := `#include <Arduino.h>
#include "bigendian.h"

namespace comprehensive_test {


// Read-write register with mixed field types
static constexpr int Reg_RW_ID = 0;
struct RW {
    // Read-write fields
    // Read-write field
    uint8_t rw_field1;
    // Read-write field
    int16_t rw_field2;
    
    // Read-only fields
    // Read-only field
    uint32_t read_field1;
    // Read-only field
    float32_t read_field2;
    
    // Write-only fields
    // Write-only field
    uint8_t write_field1;
    // Write-only field
    int64_t write_field2;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->rw_field1);
        written += bigendian::encode(buf + written, this->rw_field2);
        written += bigendian::encode(buf + written, this->read_field1);
        written += bigendian::encode(buf + written, this->read_field2);
        written += bigendian::encode(buf + written, this->write_field1);
        written += bigendian::encode(buf + written, this->write_field2);
        return written;
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->rw_field1);
        written += bigendian::encode(buf + written, this->rw_field2);
        written += bigendian::encode(buf + written, this->read_field1);
        written += bigendian::encode(buf + written, this->read_field2);
        written += bigendian::encode(buf + written, this->write_field1);
        written += bigendian::encode(buf + written, this->write_field2);
        return written;
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->rw_field1, buf + read);
        read += bigendian::decode(this->rw_field2, buf + read);
        read += bigendian::decode(this->read_field1, buf + read);
        read += bigendian::decode(this->read_field2, buf + read);
        read += bigendian::decode(this->write_field1, buf + read);
        read += bigendian::decode(this->write_field2, buf + read);
        return read;
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->rw_field1, buf + read);
        read += bigendian::decode(this->rw_field2, buf + read);
        read += bigendian::decode(this->read_field1, buf + read);
        read += bigendian::decode(this->read_field2, buf + read);
        read += bigendian::decode(this->write_field1, buf + read);
        read += bigendian::decode(this->write_field2, buf + read);
        return read;
    }
};

// Status register with detailed documentation
//
// The Status register contains device status information:
// - Bit 0: Device ready flag
// - Bits 1-3: Error code (0-7)
// - Bit 4: Power on indicator
// - Bits 5-7: Reserved for future use
//
// Error codes:
// 0 = No error
// 1 = Communication error
// 2 = Sensor error
// 3 = Power error
// 4-7 = Reserved
// Read-only register
static constexpr int Reg_Status_ID = 1;
struct Status {
    // Read-only fields
    // Device status byte
    // Bit field: status
    static constexpr uint8_t ready_bm = 0x1; // bit 0
    static constexpr uint8_t power_on_bm = 0x10; // bit 4
    uint8_t status;
    // Operation counter
    int32_t counter;
    // Configuration flags
    // Bit field: flags
    static constexpr uint8_t bit0_bm = 0x1; // bit 0
    static constexpr uint8_t bit15_bm = 0x2; // bit 1
    uint8_t flags;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->status);
        written += bigendian::encode(buf + written, this->counter);
        written += bigendian::encode(buf + written, this->flags);
        return written;
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        return -1; // read-only register has no write data
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->status, buf + read);
        read += bigendian::decode(this->counter, buf + read);
        read += bigendian::decode(this->flags, buf + read);
        return read;
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        return -1; // read-only register cannot receive write data
    }
};

// Command register for device control
// Write-only register
static constexpr int Reg_Command_ID = 2;
struct Command {
    // Write-only fields
    // Command type
    uint16_t cmd_type;
    // Command value
    int8_t cmd_value;
    // Command configuration
    // Bit field: config
    static constexpr uint8_t enable_bm = 0x1; // bit 0
    static constexpr uint8_t priority_bm = 0x6; // bits 1-2
    static constexpr uint8_t mode_bm = 0x18; // bits 3-4
    uint8_t config;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        return -1; // write-only register has no read data
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->cmd_type);
        written += bigendian::encode(buf + written, this->cmd_value);
        written += bigendian::encode(buf + written, this->config);
        return written;
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        return -1; // write-only register cannot receive read data
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->cmd_type, buf + read);
        read += bigendian::decode(this->cmd_value, buf + read);
        read += bigendian::decode(this->config, buf + read);
        return read;
    }
};

// Data register with arrays
static constexpr int Reg_Data_ID = 3;
struct Data {
    // Read-write fields
    // Fixed size array
    int16_t samples[10];
    // Variable length array
    uint8_t* buffer;
    uint8_t buffer_size;
    // Simple field
    float32_t value;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        int written = 0;
        for (int i = 0; i < 10; i++) {
            written += bigendian::encode(buf + written, this->samples[i]);
        }
        written += bigendian::encode_varray(buf + written, this->buffer, this->buffer_size);
        written += bigendian::encode(buf + written, this->value);
        return written;
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        int written = 0;
        for (int i = 0; i < 10; i++) {
            written += bigendian::encode(buf + written, this->samples[i]);
        }
        written += bigendian::encode_varray(buf + written, this->buffer, this->buffer_size);
        written += bigendian::encode(buf + written, this->value);
        return written;
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        int read = 0;
        for (int i = 0; i < 10; i++) {
            read += bigendian::decode(this->samples[i], buf + read);
        }
        read += bigendian::decode_varray(this->buffer, 1024, buf + read, this->buffer_size);
        read += bigendian::decode(this->value, buf + read);
        return read;
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        int read = 0;
        for (int i = 0; i < 10; i++) {
            read += bigendian::decode(this->samples[i], buf + read);
        }
        read += bigendian::decode_varray(this->buffer, 1024, buf + read, this->buffer_size);
        read += bigendian::decode(this->value, buf + read);
        return read;
    }
};


} // namespace comprehensive_test
`

	// Parse the Pargus specification
	device, err := parser.Parse(pargusSpec)
	require.NoError(t, err)
	require.NotNil(t, device)

	// Generate C++ code
	actualCpp, err := Generate(device, "comprehensive_test")
	require.NoError(t, err)
	require.NotEmpty(t, actualCpp)

	// Compare with expected output
	assert.Equal(t, expectedCpp, actualCpp, "Generated C++ code should match expected output")
}

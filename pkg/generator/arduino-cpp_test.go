package generator

import (
	"fmt"
	"testing"

	"github.com/dspasibenko/pargus/pkg/parser"
	"github.com/stretchr/testify/require"
)

func TestGenerateCpp(t *testing.T) {
	input := `


    // This is the test device
    device sensor
    
    
    
    register Control(1) {
        // Enable sensor
        enable uint32{bit0: 0, mode: 1-3, high: 22-31};
        // Temperature reading
        temperature int16;
    
        // this is the write only fild
        // write only field
        write_only_temp:w uint8;
        
        // Data buffer
        data [4]uint8;
    
        // Size field for variable array
        data_size uint16;
        
        // Data buffer using field reference
        data_buffer [data_size]uint32;
    
        // one more field to check flow
        flow:r float32;
    
        // Size field for another variable array
        buffer_size uint8;
        
        data_buffer2 [buffer_size]uint8; // do not remove this comment
        another uint8{bit0: 0, mode: 1-3, high: 5}; // another bit field
    
        var_array [another_high]uint8; // keeps size in bit field
    };

    // Check is the anoter read-only register
    register Check(2):r {
        field uint8; // this is going to be read only, I will check it
    };

    register WriteOnly(3):w {
        // this is the write only field I will check it
        write_only_field:w uint8; // and the comment too
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	//res, err := GenerateCpp(device, "test", "test_h")
	res, err := GenerateGo(device, "test")
	require.NoError(t, err)
	fmt.Println(res)
}

func TestGenerateCppWithBitfieldComments(t *testing.T) {
	input := `
    device sensor
    
    register Control(1) {
        // Enable sensor
        enable uint32{
            // bit0 comment
            bit0: 0,
            // mode comment
            // line2
            mode: 1-3, high: 22-31};
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	hpp, cpp, err := GenerateHppCpp(device, "test", "test_h", "test.h")
	require.NoError(t, err)
	fmt.Println("=== HPP ===")
	fmt.Println(hpp)
	fmt.Println("=== CPP ===")
	fmt.Println(cpp)
}

func TestGenerateGoWithBitfieldComments(t *testing.T) {
	input := `
    device sensor
    
    register Control(1) {
        // Enable sensor
        enable uint32{
            // bit0 comment
            bit0: 0,
            // mode comment
            // line2
            mode: 1-3, high: 22-31};
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	res, err := GenerateGo(device, "test")
	require.NoError(t, err)
	fmt.Println(res)
}

func TestGenerateCppWithRegisterRef(t *testing.T) {
	input := `
    device test
    
    register Config(1) {
        mode uint8;
        enabled uint8;
    };
    
    register Main(2) {
        id uint16;
        // Configuration settings
        config Config;
        data [4]uint8;
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	hpp, cpp, err := GenerateHppCpp(device, "test", "test_h", "test.h")
	require.NoError(t, err)
	fmt.Println("=== HPP ===")
	fmt.Println(hpp)
	fmt.Println("=== CPP ===")
	fmt.Println(cpp)

	// Verify that the generated code contains the register reference
	require.Contains(t, hpp, "Config config;")
	require.Contains(t, cpp, "{auto res = config.serialize_read(buf + offset, size - offset); if (res < 0) return res; offset += res;}")
	require.Contains(t, cpp, "{auto res = config.deserialize_read(buf + offset, size - offset); if (res < 0) return res; offset += res;}")
}

func TestGenerateGoWithRegisterRef(t *testing.T) {
	input := `
    device test
    
    register Config(1) {
        mode uint8;
        enabled uint8;
    };
    
    register Main(2) {
        id uint16;
        // Configuration settings
        config Config;
        data [4]uint8;
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	res, err := GenerateGo(device, "test")
	require.NoError(t, err)
	fmt.Println(res)

	// Verify that the generated code contains the register reference
	require.Contains(t, res, "config Config")
	// In SerializeRead we serialize nested struct by calling DeserializeRead (reads from struct to buffer)
	require.Contains(t, res, "offset += r.config.DeserializeRead(buf[offset:])")
	// In DeserializeRead we deserialize nested struct by calling SerializeRead (reads from buffer to struct)
	require.Contains(t, res, "offset += r.config.SerializeRead(buf[offset:])")
}

func TestGenerateCppWithRegisterRefReadWrite(t *testing.T) {
	input := `
    device test
    
    register Config(1) {
        mode uint8;
    };
    
    register Main(2) {
        read_config: r Config;
        write_config: w Config;
    };`

	device, err := parser.Parse(input)
	require.NoError(t, err)

	hpp, cpp, err := GenerateHppCpp(device, "test", "test_h", "test.h")
	require.NoError(t, err)
	fmt.Println("=== HPP ===")
	fmt.Println(hpp)
	fmt.Println("=== CPP ===")
	fmt.Println(cpp)

	// Verify that read_config uses serialize_read
	require.Contains(t, hpp, "Config read_config;")
	require.Contains(t, cpp, "read_config.serialize_read")

	// Verify that write_config uses serialize_write
	require.Contains(t, hpp, "Config write_config;")
	require.Contains(t, cpp, "write_config.serialize_write")
}

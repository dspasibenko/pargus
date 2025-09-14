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

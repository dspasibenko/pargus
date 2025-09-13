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

	expected := `#include <Arduino.h>
#include "bigendian.h"


// This is the test device
namespace test {

// Register IDs
static constexpr uint8_t Reg_Control_ID = 1;
static constexpr uint8_t Reg_Check_ID = 2;
static constexpr uint8_t Reg_WriteOnly_ID = 3;


struct Control {
    // Enable sensor
    static constexpr uint32_t enable_bit0_bm = 0x1;
    static constexpr uint32_t enable_mode_bm = 0xE;
    static constexpr uint32_t enable_high_bm = 0xFFC00000;
    uint32 enable;
    // Temperature reading
    int16 temperature;


    // this is the write only fild
    // write only field
    uint8 write_only_temp;


    // Data buffer
    uint8_t data[4];


    // Size field for variable array
    uint16 data_size;


    // Data buffer using field reference
    uint32_t* data_buffer;


    // one more field to check flow
    float32 flow;


    // Size field for another variable array
    uint8 buffer_size;


    uint8_t* data_buffer2; // do not remove this comment
    static constexpr uint8_t another_bit0_bm = 0x1;
    static constexpr uint8_t another_mode_bm = 0xE;
    static constexpr uint8_t another_high_bm = 0x20;
    uint8 another; // another bit field


    uint8_t* var_array; // keeps size in bit field
    // Send read-only fields to wire (register read fields -> wire)
    int send_read_data(uint8_t* buf, size_t size) {
        int offset = 0;

        if (offset + sizeof(int16_t) <= size) {
            offset += bigendian::encode(buf + offset, temperature);
        }

        if (offset + sizeof(uint8_t)*4 <= size) {
            offset += bigendian::encode(buf + offset, data);
        }

        if (offset + sizeof(uint16_t) <= size) {
            offset += bigendian::encode(buf + offset, data_size);
        }

        if (offset + sizeof(uint32_t)*data_size <= size) {
            offset += bigendian::encode_varray(buf + offset, data_buffer, data_size);
        }

        if (offset + sizeof(float) <= size) {
            offset += bigendian::encode(buf + offset, flow);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::encode(buf + offset, buffer_size);
        }

        if (offset + sizeof(uint8_t)*buffer_size <= size) {
            offset += bigendian::encode_varray(buf + offset, data_buffer2, buffer_size);
        }


        {
            uint8_t elems = (another&another_high_bm)>>5;
            if (offset + sizeof(uint8_t)*elems <= size) {
                 offset += bigendian::encode_varray(buf + offset, var_array, elems);
            }
        }

        return offset;
    }

    // Send write-only fields to wire (register write fields -> wire)
    int send_write_data(uint8_t* buf, size_t size) {
        int offset = 0;

        if (offset + sizeof(int16_t) <= size) {
            offset += bigendian::encode(buf + offset, temperature);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::encode(buf + offset, write_only_temp);
        }

        if (offset + sizeof(uint8_t)*4 <= size) {
            offset += bigendian::encode(buf + offset, data);
        }

        if (offset + sizeof(uint16_t) <= size) {
            offset += bigendian::encode(buf + offset, data_size);
        }

        if (offset + sizeof(uint32_t)*data_size <= size) {
            offset += bigendian::encode_varray(buf + offset, data_buffer, data_size);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::encode(buf + offset, buffer_size);
        }

        if (offset + sizeof(uint8_t)*buffer_size <= size) {
            offset += bigendian::encode_varray(buf + offset, data_buffer2, buffer_size);
        }


        {
            uint8_t elems = (another&another_high_bm)>>5;
            if (offset + sizeof(uint8_t)*elems <= size) {
                 offset += bigendian::encode_varray(buf + offset, var_array, elems);
            }
        }

        return offset;
    }

    // Get read-only fields from wire (wire -> the register read fields)
    int receive_read_data(uint8_t* buf, size_t size) {
        int offset = 0;

        if (offset + sizeof(int16_t) <= size) {
            offset += bigendian::decode(temperature, buf + offset);
        }

        if (offset + sizeof(uint8_t)*4 <= size) {
            offset += bigendian::decode(data, buf + offset);
        }

        if (offset + sizeof(uint16_t) <= size) {
            offset += bigendian::decode(data_size, buf + offset);
        }

        if (offset + sizeof(uint32_t)*data_size <= size) {
            offset += bigendian::decode_varray(data_buffer, buf + offset, data_size);
        }

        if (offset + sizeof(float) <= size) {
            offset += bigendian::decode(flow, buf + offset);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::decode(buffer_size, buf + offset);
        }

        if (offset + sizeof(uint8_t)*buffer_size <= size) {
            offset += bigendian::decode_varray(data_buffer2, buf + offset, buffer_size);
        }


        {
            uint8_t elems = (another&another_high_bm)>>5;
            if (offset + sizeof(uint8_t)*elems <= size) {
                 offset += bigendian::decode_varray(var_array, buf + offset, elems);
            }
        }

        return offset;
    }

    // Get write-only fields from wire (wire -> the register writable fields)
    int receive_write_data(uint8_t* buf, size_t size) {
        int offset = 0;

        if (offset + sizeof(int16_t) <= size) {
            offset += bigendian::decode(temperature, buf + offset);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::decode(write_only_temp, buf + offset);
        }

        if (offset + sizeof(uint8_t)*4 <= size) {
            offset += bigendian::decode(data, buf + offset);
        }

        if (offset + sizeof(uint16_t) <= size) {
            offset += bigendian::decode(data_size, buf + offset);
        }

        if (offset + sizeof(uint32_t)*data_size <= size) {
            offset += bigendian::decode_varray(data_buffer, buf + offset, data_size);
        }

        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::decode(buffer_size, buf + offset);
        }

        if (offset + sizeof(uint8_t)*buffer_size <= size) {
            offset += bigendian::decode_varray(data_buffer2, buf + offset, buffer_size);
        }


        {
            uint8_t elems = (another&another_high_bm)>>5;
            if (offset + sizeof(uint8_t)*elems <= size) {
                 offset += bigendian::decode_varray(var_array, buf + offset, elems);
            }
        }

        return offset;
    }
};


// Check is the anoter read-only register
struct Check {
    uint8 field; // this is going to be read only, I will check it
    // Send read-only fields to wire (register read fields -> wire)
    int send_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::encode(buf + offset, field);
        }

        return offset;
    }

    // Send write-only fields to wire (register write fields -> wire)
    int send_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
        return offset;
    }

    // Get read-only fields from wire (wire -> the register read fields)
    int receive_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::decode(field, buf + offset);
        }

        return offset;
    }

    // Get write-only fields from wire (wire -> the register writable fields)
    int receive_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
        return offset;
    }
};


struct WriteOnly {
    // this is the write only field I will check it
    uint8 write_only_field; // and the comment too
    // Send read-only fields to wire (register read fields -> wire)
    int send_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
        return offset;
    }

    // Send write-only fields to wire (register write fields -> wire)
    int send_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::encode(buf + offset, write_only_field);
        }

        return offset;
    }

    // Get read-only fields from wire (wire -> the register read fields)
    int receive_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
        return offset;
    }

    // Get write-only fields from wire (wire -> the register writable fields)
    int receive_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
        if (offset + sizeof(uint8_t) <= size) {
            offset += bigendian::decode(write_only_field, buf + offset);
        }

        return offset;
    }
};

} // namespace test
`
	_ = expected
	device, err := parser.Parse(input)
	require.NoError(t, err)

	res, err := GenerateCpp(device, "test")
	require.NoError(t, err)
	fmt.Println(res)
	//assert.Equal(t, expected, res)
}

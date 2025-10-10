package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorExampleComments(t *testing.T) {
	// Test the specific example from the generator test to ensure comments are parsed correctly
	input := `// This is the test device
device sensor


register Control(1) {
    // Enable sensor
    enable uint32{
		// bit0 comment
		bit0: 0,
		// mode comment
		// line2
		mode: 1-3, high: 22-31};
    // Temperature reading
    temperature int16;

	// this is the write only fild
	// write only field
	write_only_temp:w uint8;
    
	// Data buffer
    data [4]uint8;

	// Data buffer we expect
	// to have extra filed for the size of the buffer in th struct like data_buffer_size
    data_buffer [write_only_temp]uint32;

	// one more field to check flow
	flow:r float32;
};

// Check is the anoter read-only register
register Check(2):r {
	field uint8; // this is going to be read only, I will check it
};

register WriteOnly(3):w {
	// this is the write only field I will check it
	write_only_field:w uint8; // and the comment too
};


`

	device, err := Parse(input)
	require.NoError(t, err)

	// Test device comments
	require.NotNil(t, device.Doc, "Device should have comments")
	assert.Len(t, device.Doc.Elements, 1, "Device should have 1 comment elements")
	assert.Equal(t, "// This is the test device", *device.Doc.Elements[0].Comment, "Device comment should match")

	// Test Control register
	require.Len(t, device.Registers, 3, "Should have 3 registers")
	controlRegister := device.Registers[0]
	assert.Equal(t, "Control", controlRegister.Name, "Control register name should match")
	assert.Equal(t, int64(1), controlRegister.Number(), "Control register number should match")
	assert.Len(t, controlRegister.Fields, 6, "Control register should have 6 fields")

	// Test enable field
	enableField := controlRegister.Fields[0]
	assert.Equal(t, "enable", enableField.Name, "Enable field name should match")
	require.NotNil(t, enableField.Doc, "Enable field should have leading comments")
	assert.Len(t, enableField.Doc.Elements, 1, "Enable field should have 1 leading comment")
	require.NotNil(t, enableField.Doc.Elements[0].Comment, "Enable field leading comment should exist")
	assert.Equal(t, "// Enable sensor", *enableField.Doc.Elements[0].Comment, "Enable field leading comment should match")
	require.NotNil(t, enableField.TrailingComment, "Enable field should have trailing comment")
	assert.Equal(t, "", *enableField.TrailingComment, "Enable field trailing comment should be empty")

	// Test enable bitfield members
	require.NotNil(t, enableField.Type, "Enable field should have type")
	require.NotNil(t, enableField.Type.Bitfield, "Enable field should be a bitfield")
	assert.Len(t, enableField.Type.Bitfield.Bits, 3, "Enable bitfield should have 3 bit members")

	// Test bit0 member
	bit0Member := enableField.Type.Bitfield.Bits[0]
	assert.Equal(t, "bit0", bit0Member.Name, "First bit member name should be bit0")
	require.NotNil(t, bit0Member.Doc, "bit0 should have comments")
	assert.Len(t, bit0Member.Doc.Elements, 1, "bit0 should have 1 comment")
	require.NotNil(t, bit0Member.Doc.Elements[0].Comment, "bit0 comment should exist")
	assert.Equal(t, "// bit0 comment", *bit0Member.Doc.Elements[0].Comment, "bit0 comment should match")

	// Test mode member
	modeMember := enableField.Type.Bitfield.Bits[1]
	assert.Equal(t, "mode", modeMember.Name, "Second bit member name should be mode")
	require.NotNil(t, modeMember.Doc, "mode should have comments")
	assert.Len(t, modeMember.Doc.Elements, 2, "mode should have 2 comments")
	require.NotNil(t, modeMember.Doc.Elements[0].Comment, "First mode comment should exist")
	assert.Equal(t, "// mode comment", *modeMember.Doc.Elements[0].Comment, "First mode comment should match")
	require.NotNil(t, modeMember.Doc.Elements[1].Comment, "Second mode comment should exist")
	assert.Equal(t, "// line2", *modeMember.Doc.Elements[1].Comment, "Second mode comment should match")

	// Test high member
	highMember := enableField.Type.Bitfield.Bits[2]
	assert.Equal(t, "high", highMember.Name, "Third bit member name should be high")
	assert.True(t, highMember.Doc == nil || len(highMember.Doc.Elements) == 0, "high should not have comments")

	// Test temperature field
	temperatureField := controlRegister.Fields[1]
	assert.Equal(t, "temperature", temperatureField.Name, "Temperature field name should match")
	require.NotNil(t, temperatureField.Doc, "Temperature field should have leading comments")
	assert.Len(t, temperatureField.Doc.Elements, 1, "Temperature field should have 1 leading comment")
	require.NotNil(t, temperatureField.Doc.Elements[0].Comment, "Temperature field leading comment should exist")
	assert.Equal(t, "// Temperature reading", *temperatureField.Doc.Elements[0].Comment, "Temperature field leading comment should match")
	require.NotNil(t, temperatureField.TrailingComment, "Temperature field should have trailing comment")
	assert.Equal(t, "", *temperatureField.TrailingComment, "Temperature field trailing comment should be empty")

	// Test write_only_temp field
	writeOnlyTempField := controlRegister.Fields[2]
	assert.Equal(t, "write_only_temp", writeOnlyTempField.Name, "WriteOnlyTemp field name should match")
	require.NotNil(t, writeOnlyTempField.Doc, "WriteOnlyTemp field should have leading comments")
	assert.Len(t, writeOnlyTempField.Doc.Elements, 3, "WriteOnlyTemp field should have 3 leading comments")
	require.NotNil(t, writeOnlyTempField.Doc.Elements[0].EmptyLine, "First leading element should be empty line")
	require.NotNil(t, writeOnlyTempField.Doc.Elements[1].Comment, "Second leading comment should exist")
	assert.Equal(t, "// this is the write only fild", *writeOnlyTempField.Doc.Elements[1].Comment, "Second leading comment should match")
	require.NotNil(t, writeOnlyTempField.Doc.Elements[2].Comment, "Third leading comment should exist")
	assert.Equal(t, "// write only field", *writeOnlyTempField.Doc.Elements[2].Comment, "Third leading comment should match")
	require.NotNil(t, writeOnlyTempField.TrailingComment, "WriteOnlyTemp field should have trailing comment")
	assert.Equal(t, "", *writeOnlyTempField.TrailingComment, "WriteOnlyTemp field trailing comment should be empty")

	// Test data field
	dataField := controlRegister.Fields[3]
	assert.Equal(t, "data", dataField.Name, "Data field name should match")
	require.NotNil(t, dataField.Doc, "Data field should have leading comments")
	assert.Len(t, dataField.Doc.Elements, 2, "Data field should have 2 leading comments")
	require.NotNil(t, dataField.Doc.Elements[0].EmptyLine, "First leading element should be empty line")
	require.NotNil(t, dataField.Doc.Elements[1].Comment, "Second leading comment should exist")
	assert.Equal(t, "// Data buffer", *dataField.Doc.Elements[1].Comment, "Second leading comment should match")
	require.NotNil(t, dataField.TrailingComment, "Data field should have trailing comment")
	assert.Equal(t, "", *dataField.TrailingComment, "Data field trailing comment should be empty")

	// Test data_buffer field
	dataBufferField := controlRegister.Fields[4]
	assert.Equal(t, "data_buffer", dataBufferField.Name, "DataBuffer field name should match")
	require.NotNil(t, dataBufferField.Doc, "DataBuffer field should have leading comments")
	assert.Len(t, dataBufferField.Doc.Elements, 3, "DataBuffer field should have 3 leading comments")
	require.NotNil(t, dataBufferField.Doc.Elements[0].EmptyLine, "First leading element should be empty line")
	require.NotNil(t, dataBufferField.Doc.Elements[1].Comment, "Second leading comment should exist")
	assert.Equal(t, "// Data buffer we expect", *dataBufferField.Doc.Elements[1].Comment, "Second leading comment should match")
	require.NotNil(t, dataBufferField.Doc.Elements[2].Comment, "Third leading comment should exist")
	assert.Equal(t, "// to have extra filed for the size of the buffer in th struct like data_buffer_size", *dataBufferField.Doc.Elements[2].Comment, "Third leading comment should match")
	require.NotNil(t, dataBufferField.TrailingComment, "DataBuffer field should have trailing comment")
	assert.Equal(t, "", *dataBufferField.TrailingComment, "DataBuffer field trailing comment should be empty")

	// Test flow field
	flowField := controlRegister.Fields[5]
	assert.Equal(t, "flow", flowField.Name, "Flow field name should match")
	require.NotNil(t, flowField.Doc, "Flow field should have leading comments")
	assert.Len(t, flowField.Doc.Elements, 2, "Flow field should have 2 leading comments")
	require.NotNil(t, flowField.Doc.Elements[0].EmptyLine, "First leading element should be empty line")
	require.NotNil(t, flowField.Doc.Elements[1].Comment, "Second leading comment should exist")
	assert.Equal(t, "// one more field to check flow", *flowField.Doc.Elements[1].Comment, "Second leading comment should match")
	require.NotNil(t, flowField.TrailingComment, "Flow field should have trailing comment")
	assert.Equal(t, "", *flowField.TrailingComment, "Flow field trailing comment should be empty")

	// Test Check register
	checkRegister := device.Registers[1]
	assert.Equal(t, "Check", checkRegister.Name, "Check register name should match")
	assert.Equal(t, int64(2), checkRegister.Number(), "Check register number should match")
	require.NotNil(t, checkRegister.Specifier, "Check register should have specifier")
	assert.Equal(t, "r", checkRegister.Specifier, "Check register should be read-only")
	assert.Len(t, checkRegister.Fields, 1, "Check register should have 1 field")

	// Test Check field
	checkField := checkRegister.Fields[0]
	assert.Equal(t, "field", checkField.Name, "Check field name should match")
	assert.True(t, checkField.Doc == nil || len(checkField.Doc.Elements) == 0, "Check field should not have leading comments")
	require.NotNil(t, checkField.TrailingComment, "Check field should have trailing comment")
	assert.Equal(t, "// this is going to be read only, I will check it", *checkField.TrailingComment, "Check field trailing comment should match")

	// Test WriteOnly register
	writeOnlyRegister := device.Registers[2]
	assert.Equal(t, "WriteOnly", writeOnlyRegister.Name, "WriteOnly register name should match")
	assert.Equal(t, int64(3), writeOnlyRegister.Number(), "WriteOnly register number should match")
	require.NotNil(t, writeOnlyRegister.Specifier, "WriteOnly register should have specifier")
	assert.Equal(t, "w", writeOnlyRegister.Specifier, "WriteOnly register should be write-only")
	assert.Len(t, writeOnlyRegister.Fields, 1, "WriteOnly register should have 1 field")

	// Test WriteOnly field
	writeOnlyField := writeOnlyRegister.Fields[0]
	assert.Equal(t, "write_only_field", writeOnlyField.Name, "WriteOnly field name should match")
	require.NotNil(t, writeOnlyField.Doc, "WriteOnly field should have leading comments")
	assert.Len(t, writeOnlyField.Doc.Elements, 1, "WriteOnly field should have 1 leading comment")
	require.NotNil(t, writeOnlyField.Doc.Elements[0].Comment, "WriteOnly field leading comment should exist")
	assert.Equal(t, "// this is the write only field I will check it", *writeOnlyField.Doc.Elements[0].Comment, "WriteOnly field leading comment should match")
	require.NotNil(t, writeOnlyField.TrailingComment, "WriteOnly field should have trailing comment")
	assert.Equal(t, "// and the comment too", *writeOnlyField.TrailingComment, "WriteOnly field trailing comment should match")
}

package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentDistribution(t *testing.T) {
	// Test comment distribution across different sections
	input := `// this is first line of device
// this is second line 
device device

// this comment doesn't relate to the R1

//this is first comment of R1
// this is second comment of R1
register R1(1) {
// this is about field
field int8; // this is trailing comment for field

// this comment is not for any field, but we have to keep it in the order

// this is the comment for field2
field2 uint16; 
};`

	device, err := Parse(input)
	require.NoError(t, err)

	// Test device comments
	require.NotNil(t, device.Doc, "Device should have comments")
	assert.Len(t, device.Doc.Elements, 2, "Device should have 2 comment elements")

	// Check first comment
	require.NotNil(t, device.Doc.Elements[0].Comment, "First element should be a comment")
	assert.Equal(t, "// this is first line of device", *device.Doc.Elements[0].Comment, "First device comment should match")

	// Check second comment
	require.NotNil(t, device.Doc.Elements[1].Comment, "Second element should be a comment")
	assert.Equal(t, "// this is second line ", *device.Doc.Elements[1].Comment, "Second device comment should match")

	// Test register comments
	require.Len(t, device.Registers, 1, "Should have 1 register")

	register := device.Registers[0]
	require.NotNil(t, register.Doc, "Register should have comments")
	assert.Len(t, register.Doc.Elements, 5, "Register should have 5 comment elements (empty lines are not elided)")

	// Check first element (empty line)
	require.NotNil(t, register.Doc.Elements[0].EmptyLine, "First element should be empty line")

	// Check second comment
	require.NotNil(t, register.Doc.Elements[1].Comment, "Second element should be a comment")
	assert.Equal(t, "// this comment doesn't relate to the R1", *register.Doc.Elements[1].Comment, "Second register comment should match")

	// Check third element (empty line)
	require.NotNil(t, register.Doc.Elements[2].EmptyLine, "Third element should be empty line")

	// Check fourth comment
	require.NotNil(t, register.Doc.Elements[3].Comment, "Fourth element should be a comment")
	assert.Equal(t, "//this is first comment of R1", *register.Doc.Elements[3].Comment, "Fourth register comment should match")

	// Check fifth comment
	require.NotNil(t, register.Doc.Elements[4].Comment, "Fifth element should be a comment")
	assert.Equal(t, "// this is second comment of R1", *register.Doc.Elements[4].Comment, "Fifth register comment should match")

	// Test field comments
	require.Len(t, register.Fields, 2, "Should have 2 fields")

	// Test first field
	field1 := register.Fields[0]
	assert.Equal(t, "field", field1.Name, "First field name should be 'field'")

	// Test first field leading comments
	require.NotNil(t, field1.Doc, "First field should have leading comments")
	assert.Len(t, field1.Doc.Elements, 1, "First field should have 1 leading comment element")
	require.NotNil(t, field1.Doc.Elements[0].Comment, "First field leading element should be a comment")
	expectedLeadingComment := "// this is about field"
	assert.Equal(t, expectedLeadingComment, *field1.Doc.Elements[0].Comment, "First field leading comment should match")

	// Test first field trailing comment (should stay with the field)
	require.NotNil(t, field1.TrailingComment, "First field should have trailing comment")
	assert.Equal(t, "// this is trailing comment for field", *field1.TrailingComment, "First field trailing comment should match")

	// Test second field
	field2 := register.Fields[1]
	assert.Equal(t, "field2", field2.Name, "Second field name should be 'field2'")

	// Test second field leading comments
	require.NotNil(t, field2.Doc, "Second field should have leading comments")
	assert.Len(t, field2.Doc.Elements, 4, "Second field should have 4 leading comment elements")

	// Check first element (empty line)
	require.NotNil(t, field2.Doc.Elements[0].EmptyLine, "First element should be empty line")

	// Check second comment
	require.NotNil(t, field2.Doc.Elements[1].Comment, "Second element should be a comment")
	assert.Equal(t, "// this comment is not for any field, but we have to keep it in the order", *field2.Doc.Elements[1].Comment, "Second field2 leading comment should match")

	// Check third element (empty line)
	require.NotNil(t, field2.Doc.Elements[2].EmptyLine, "Third element should be empty line")

	// Check fourth comment
	require.NotNil(t, field2.Doc.Elements[3].Comment, "Fourth element should be a comment")
	assert.Equal(t, "// this is the comment for field2", *field2.Doc.Elements[3].Comment, "Fourth field2 leading comment should match")

	// Test second field trailing comment
	require.NotNil(t, field2.TrailingComment, "Second field should have trailing comment")
	assert.Equal(t, "", *field2.TrailingComment, "Second field trailing comment should be empty")
}

func TestBasicParsing(t *testing.T) {
	// Test basic parsing functionality
	input := `device test-device
register R1(1) {
    counter int32;
    status uint8{bit0: 0, flags: 1-3};
    data [5]int16;
    string [uint8]uint8;
};`

	device, err := Parse(input)
	require.NoError(t, err)

	assert.Equal(t, "test-device", device.Name, "Device name should match")
	assert.Len(t, device.Registers, 1, "Should have 1 register")

	register := device.Registers[0]
	assert.Equal(t, "R1", register.Name, "Register name should match")
	assert.Equal(t, 1, register.Number, "Register number should match")
	assert.Len(t, register.Fields, 4, "Should have 4 fields")
}

func TestSpecifiers(t *testing.T) {
	// Test register and field specifiers
	input := `device test
register ReadOnly(1): r {
    field1 int8;
    field2: r int16;
};

register WriteOnly(2): w {
    field3: w int32;
    field4 uint8;
};`

	device, err := Parse(input)
	require.NoError(t, err)

	assert.Len(t, device.Registers, 2, "Should have 2 registers")

	// Check first register (read-only)
	roRegister := device.Registers[0]
	assert.Equal(t, "ReadOnly", roRegister.Name, "First register name should match")
	require.NotNil(t, roRegister.Specifier, "First register should have specifier")
	assert.Equal(t, "r", *roRegister.Specifier, "First register should be read-only")

	// Check second register (write-only)
	woRegister := device.Registers[1]
	assert.Equal(t, "WriteOnly", woRegister.Name, "Second register name should match")
	require.NotNil(t, woRegister.Specifier, "Second register should have specifier")
	assert.Equal(t, "w", *woRegister.Specifier, "Second register should be write-only")
}

func TestValidation(t *testing.T) {
	// Test validation errors
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "duplicate register numbers",
			input: `device test
register R1(1) { field1 int8; };
register R2(1) { field2 int8; };`,
			expectError: true,
			errorMsg:    "duplicate register number",
		},
		{
			name: "incompatible field specifier",
			input: `device test
register ReadOnly(1): r {
    field1: w int8;
};`,
			expectError: true,
			errorMsg:    "cannot be write-only because register is read-only",
		},
		{
			name: "invalid bitfield type",
			input: `device test
register R1(1) {
    field1 int8{bit0: 0};
};`,
			expectError: true,
			errorMsg:    "must use unsigned integer type",
		},
		{
			name: "invalid array size type",
			input: `device test
register R1(1) {
    field1 [int8]uint8;
};`,
			expectError: true,
			errorMsg:    "references undefined field",
		},
		{
			name: "variable array with undefined field reference",
			input: `device test
register R1(1) {
    field1 [undefined_field]uint8;
};`,
			expectError: true,
			errorMsg:    "references undefined field",
		},
		{
			name: "variable array with field reference after array",
			input: `device test
register R1(1) {
    field1 [size_field]uint8;
    size_field uint8;
};`,
			expectError: true,
			errorMsg:    "references undefined field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}
		})
	}
}

func TestGeneratorExampleComments(t *testing.T) {
	// Test the specific example from the generator test to ensure comments are parsed correctly
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

	// Data buffer we expect
	// to have extra filed for the size of the buffer in th struct like data_buffer_size
    data_buffer [uint16]uint32;

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
	assert.Len(t, device.Doc.Elements, 2, "Device should have 2 comment elements")
	require.NotNil(t, device.Doc.Elements[0].EmptyLine, "First device element should be empty line")
	require.NotNil(t, device.Doc.Elements[1].Comment, "Second device element should be comment")
	assert.Equal(t, "// This is the test device", *device.Doc.Elements[1].Comment, "Device comment should match")

	// Test Control register
	require.Len(t, device.Registers, 3, "Should have 3 registers")
	controlRegister := device.Registers[0]
	assert.Equal(t, "Control", controlRegister.Name, "Control register name should match")
	assert.Equal(t, 1, controlRegister.Number, "Control register number should match")
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
	assert.Equal(t, 2, checkRegister.Number, "Check register number should match")
	require.NotNil(t, checkRegister.Specifier, "Check register should have specifier")
	assert.Equal(t, "r", *checkRegister.Specifier, "Check register should be read-only")
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
	assert.Equal(t, 3, writeOnlyRegister.Number, "WriteOnly register number should match")
	require.NotNil(t, writeOnlyRegister.Specifier, "WriteOnly register should have specifier")
	assert.Equal(t, "w", *writeOnlyRegister.Specifier, "WriteOnly register should be write-only")
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

func TestTrailingCommentDebug(t *testing.T) {
	// Simple test to debug trailing comment parsing
	input := `device test
register R1(1) {
    field int8; // this is trailing comment
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 1, "Should have 1 field")

	field := register.Fields[0]
	assert.Equal(t, "field", field.Name, "Field name should match")

	// Debug output
	t.Logf("Field trailing comment: %v", field.TrailingComment)
	if field.TrailingComment != nil {
		t.Logf("Field trailing comment value: '%s'", *field.TrailingComment)
	}
}

func TestEnableFieldDebug(t *testing.T) {
	// Test the specific enable field from the failing test
	input := `device sensor
register Control(1) {
    // Enable sensor
    enable uint32{bit0: 0, mode: 1-3, high: 22-31};
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 1, "Should have 1 field")

	field := register.Fields[0]
	assert.Equal(t, "enable", field.Name, "Field name should match")

	// Debug output
	t.Logf("Enable field trailing comment: %v", field.TrailingComment)
	if field.TrailingComment != nil {
		t.Logf("Enable field trailing comment value: '%s'", *field.TrailingComment)
	}

	t.Logf("Enable field leading comments: %v", field.Doc)
	if field.Doc != nil {
		t.Logf("Enable field leading comments elements: %d", len(field.Doc.Elements))
		for i, element := range field.Doc.Elements {
			if element.Comment != nil {
				t.Logf("  Element %d comment: '%s'", i, *element.Comment)
			} else if element.EmptyLine != nil {
				t.Logf("  Element %d empty line: '%s'", i, *element.EmptyLine)
			}
		}
	}
}

func TestFullContextDebug(t *testing.T) {
	// Test with the full context from the failing test
	input := `device sensor
register Control(1) {
    // Enable sensor
    enable uint32{bit0: 0, mode: 1-3, high: 22-31};
    // Temperature reading
    temperature int16;
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 2, "Should have 2 fields")

	// Debug enable field
	enableField := register.Fields[0]
	assert.Equal(t, "enable", enableField.Name, "Enable field name should match")
	t.Logf("Enable field trailing comment: %v", enableField.TrailingComment)
	if enableField.TrailingComment != nil {
		t.Logf("Enable field trailing comment value: '%s'", *enableField.TrailingComment)
	}

	// Debug temperature field
	tempField := register.Fields[1]
	assert.Equal(t, "temperature", tempField.Name, "Temperature field name should match")
	t.Logf("Temperature field trailing comment: %v", tempField.TrailingComment)
	if tempField.TrailingComment != nil {
		t.Logf("Temperature field trailing comment value: '%s'", *tempField.TrailingComment)
	}

	t.Logf("Temperature field leading comments: %v", tempField.Doc)
	if tempField.Doc != nil {
		t.Logf("Temperature field leading comments elements: %d", len(tempField.Doc.Elements))
		for i, element := range tempField.Doc.Elements {
			if element.Comment != nil {
				t.Logf("  Element %d comment: '%s'", i, *element.Comment)
			} else if element.EmptyLine != nil {
				t.Logf("  Element %d empty line: '%s'", i, *element.EmptyLine)
			}
		}
	}
}

func TestLineBreakDebug(t *testing.T) {
	// Test with explicit line breaks
	input := `device sensor
register Control(1) {
    // Enable sensor
    enable uint32{bit0: 0, mode: 1-3, high: 22-31};
    
    // Temperature reading
    temperature int16;
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 2, "Should have 2 fields")

	// Debug enable field
	enableField := register.Fields[0]
	assert.Equal(t, "enable", enableField.Name, "Enable field name should match")
	t.Logf("Enable field trailing comment: %v", enableField.TrailingComment)
	if enableField.TrailingComment != nil {
		t.Logf("Enable field trailing comment value: '%s'", *enableField.TrailingComment)
	}

	// Debug temperature field
	tempField := register.Fields[1]
	assert.Equal(t, "temperature", tempField.Name, "Temperature field name should match")
	t.Logf("Temperature field trailing comment: %v", tempField.TrailingComment)
	if tempField.TrailingComment != nil {
		t.Logf("Temperature field trailing comment value: '%s'", *tempField.TrailingComment)
	}

	t.Logf("Temperature field leading comments: %v", tempField.Doc)
	if tempField.Doc != nil {
		t.Logf("Temperature field leading comments elements: %d", len(tempField.Doc.Elements))
		for i, element := range tempField.Doc.Elements {
			if element.Comment != nil {
				t.Logf("  Element %d comment: '%s'", i, *element.Comment)
			} else if element.EmptyLine != nil {
				t.Logf("  Element %d empty line: '%s'", i, *element.EmptyLine)
			}
		}
	}
}

func TestCommentHeuristicDebug(t *testing.T) {
	// Test the heuristic for comment detection
	comments := []string{
		"// this is trailing comment for field", // should NOT be moved (has space after //)
		"//Temperature reading",                 // should be moved (no space after //)
		"// Temperature reading",                // should NOT be moved (has space after //)
	}

	for i, comment := range comments {
		t.Logf("Comment %d: '%s'", i, comment)
		hasSpace := strings.HasPrefix(comment, "// ")
		noSpace := strings.HasPrefix(comment, "//") && !strings.HasPrefix(comment, "// ")
		t.Logf("  Has space after //: %v", hasSpace)
		t.Logf("  No space after //: %v", noSpace)
		t.Logf("  Should move: %v", noSpace)
	}
}

func TestRealCommentDebug(t *testing.T) {
	// Test with the actual input from TestGeneratorExampleComments
	input := `device sensor
register Control(1) {
    // Enable sensor
    enable uint32{bit0: 0, mode: 1-3, high: 22-31};
    // Temperature reading
    temperature int16;
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 2, "Should have 2 fields")

	// Debug enable field
	enableField := register.Fields[0]
	assert.Equal(t, "enable", enableField.Name, "Enable field name should match")
	t.Logf("Enable field trailing comment: %v", enableField.TrailingComment)
	if enableField.TrailingComment != nil {
		comment := *enableField.TrailingComment
		t.Logf("Enable field trailing comment value: '%s'", comment)
		t.Logf("Comment bytes: %v", []byte(comment))
		t.Logf("Has space after //: %v", strings.HasPrefix(comment, "// "))
		t.Logf("No space after //: %v", strings.HasPrefix(comment, "//") && !strings.HasPrefix(comment, "// "))
	}
}

func TestTrailingCommentOnNewLine(t *testing.T) {
	// Test that trailing comments on new lines are NOT captured
	input := `device test
register R1(1) {
    field1 int8;
    // This comment should NOT be captured as trailing comment for field1
    field2 int16; // This comment SHOULD be captured as trailing comment
};`

	device, err := Parse(input)
	require.NoError(t, err)

	require.Len(t, device.Registers, 1, "Should have 1 register")
	register := device.Registers[0]
	require.Len(t, register.Fields, 2, "Should have 2 fields")

	// Test field1 - should have empty trailing comment
	field1 := register.Fields[0]
	assert.Equal(t, "field1", field1.Name, "Field1 name should match")
	require.NotNil(t, field1.TrailingComment, "Field1 should have trailing comment")
	assert.Equal(t, "", *field1.TrailingComment, "Field1 trailing comment should be empty")

	// Test field2 - should have trailing comment
	field2 := register.Fields[1]
	assert.Equal(t, "field2", field2.Name, "Field2 name should match")
	require.NotNil(t, field2.TrailingComment, "Field2 should have trailing comment")
	assert.Equal(t, "// This comment SHOULD be captured as trailing comment", *field2.TrailingComment, "Field2 trailing comment should match")
}

func TestVariableArrayValidation(t *testing.T) {
	// Test successful cases for variable arrays
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "variable array with field reference",
			input: `device test
register R1(1) {
    size_field uint8;
    data [size_field]uint8;
};`,
		},
		{
			name: "variable array with bitfield reference",
			input: `device test
register R1(1) {
    flags uint16{size: 0-7, other: 8-15};
    data [size]uint8;
};`,
		},
		{
			name: "multiple variable arrays with different references",
			input: `device test
register R1(1) {
    count1 uint8;
    count2 uint16;
    data1 [count1]uint8;
    data2 [count2]uint16;
};`,
		},
		{
			name: "variable array with uint8 type size",
			input: `device test
register R1(1) {
    data [uint8]uint8;
};`,
		},
		{
			name: "variable array with uint16 type size",
			input: `device test
register R1(1) {
    data [uint16]uint8;
};`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			assert.NoError(t, err, "Expected no error for valid variable array: %s", tt.name)
		})
	}
}

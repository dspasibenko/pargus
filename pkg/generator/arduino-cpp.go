package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/dspasibenko/pargus/pkg/parser"
)

// ArduinoCppGenerator generates Arduino C++ header files from AST
type ArduinoCppGenerator struct {
	template *template.Template
}

// NewArduinoCppGenerator creates a new Arduino C++ generator
func NewArduinoCppGenerator() (*ArduinoCppGenerator, error) {
	tmpl, err := template.New("arduino-cpp").Funcs(template.FuncMap{
		"toCppType":                               toCppType,
		"toCppArrayType":                          toCppArrayType,
		"toCppFieldType":                          toCppFieldType,
		"toCppArraySize":                          toCppArraySize,
		"toBitMask":                               toBitMask,
		"formatComments":                          formatComments,
		"formatFieldComments":                     formatFieldComments,
		"formatFieldCommentsWithTrailing":         formatFieldCommentsWithTrailing,
		"formatFieldCommentsWithNextTrailing":     formatFieldCommentsWithNextTrailing,
		"formatFieldCommentsWithPreviousTrailing": formatFieldCommentsWithPreviousTrailing,
		"getFieldSpecifier":                       getFieldSpecifier,
		"getFieldSpecifierFromContext":            getFieldSpecifierFromContext,
		"hasReadFields":                           hasReadFields,
		"hasWriteFields":                          hasWriteFields,
		"getReadFields":                           getReadFields,
		"getWriteFields":                          getWriteFields,
		"getBitFields":                            getBitFields,
		"getNonBitFields":                         getNonBitFields,
		"getRegisterAddress":                      getRegisterAddress,
		"findFieldByName":                         findFieldByName,
		"findBitFieldMemberByName":                findBitFieldMemberByName,
		"getArraySizeField":                       getArraySizeField,
		"getArraySizeFieldForField":               getArraySizeFieldForField,
		"getArraySizeFieldName":                   getArraySizeFieldName,
		"getArraySizeFieldExpression":             getArraySizeFieldExpression,
		"isBitFieldReference":                     isBitFieldReference,
		"sub":                                     func(a, b int) int { return a - b },
	}).Parse(arduinoCppTemplate)

	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &ArduinoCppGenerator{
		template: tmpl,
	}, nil
}

// Generate generates Arduino C++ header content from device AST
func (g *ArduinoCppGenerator) Generate(device *parser.Device, namespace string) (string, error) {
	var buf bytes.Buffer

	data := struct {
		Device    *parser.Device
		Namespace string
	}{
		Device:    device,
		Namespace: namespace,
	}

	if err := g.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Template functions

// toCppType converts Go type to C++ type
func toCppType(typeUnion *parser.TypeUnion) string {
	if typeUnion.Simple != nil {
		switch typeUnion.Simple.Name {
		case "int8":
			return "int8_t"
		case "uint8":
			return "uint8_t"
		case "int16":
			return "int16_t"
		case "uint16":
			return "uint16_t"
		case "int32":
			return "int32_t"
		case "uint32":
			return "uint32_t"
		case "int64":
			return "int64_t"
		case "uint64":
			return "uint64_t"
		case "float32":
			return "float"
		case "float64":
			return "double"
		default:
			return typeUnion.Simple.Name
		}
	}

	if typeUnion.Array != nil {
		return toCppArrayType(typeUnion.Array)
	}

	if typeUnion.Bitfield != nil {
		return toCppType(&parser.TypeUnion{Simple: &parser.SimpleType{Name: typeUnion.Bitfield.Base}})
	}

	return "unknown"
}

// toCppArrayType converts array type to C++ array type
func toCppArrayType(arrayType *parser.ArrayType) string {
	elementType := toCppType(arrayType.Element)

	if arrayType.Size.Constant != nil {
		return fmt.Sprintf("%s[%d]", elementType, *arrayType.Size.Constant)
	}

	if arrayType.Size.Variable != nil {
		return fmt.Sprintf("%s*", elementType)
	}

	if arrayType.Size.Type != nil {
		return fmt.Sprintf("%s*", elementType)
	}

	return elementType
}

// toCppFieldType converts field type to C++ field declaration
func toCppFieldType(field *parser.Field) string {
	if field.Type.Array != nil {
		elementType := toCppType(field.Type.Array.Element)
		if field.Type.Array.Size.Constant != nil {
			return elementType // Return just the element type, array size will be added in template
		}
		if field.Type.Array.Size.Variable != nil {
			return fmt.Sprintf("%s*", elementType)
		}
		if field.Type.Array.Size.Type != nil {
			return fmt.Sprintf("%s*", elementType)
		}
		return elementType
	}
	return toCppType(field.Type)
}

// toCppArraySize returns array size for field declaration
func toCppArraySize(field *parser.Field) string {
	if field.Type.Array != nil && field.Type.Array.Size.Constant != nil {
		return fmt.Sprintf("[%d]", *field.Type.Array.Size.Constant)
	}
	return ""
}

// toBitMask calculates bit mask for bit field
func toBitMask(start int, end *int) string {
	var actualEnd int
	if end != nil {
		actualEnd = *end
	} else {
		actualEnd = start
	}

	if actualEnd == start {
		return fmt.Sprintf("0x%02X", 1<<start)
	}

	mask := 0
	for i := start; i <= actualEnd; i++ {
		mask |= 1 << i
	}

	return fmt.Sprintf("0x%02X", mask)
}

// formatComments formats comment group for output
func formatComments(comments *parser.CommentGroup) string {
	if comments == nil {
		return ""
	}

	var result strings.Builder
	for i, element := range comments.Elements {
		if element.Comment != nil {
			result.WriteString(*element.Comment)
			// Only add newline if this is not the last element
			if i < len(comments.Elements)-1 {
				result.WriteString("\n")
			}
		}
		// Skip empty lines in comments
	}
	return result.String()
}

// formatFieldComments formats field comments for output
func formatFieldComments(field *parser.Field) string {
	var result strings.Builder

	// Leading comments
	if field.Doc != nil {
		comments := formatComments(field.Doc)
		if comments != "" {
			// Add proper indentation to each line
			lines := strings.Split(comments, "\n")
			for i, line := range lines {
				if line != "" {
					if i > 0 {
						result.WriteString("\n")
					}
					result.WriteString("    ")
					result.WriteString(line)
				}
			}
			result.WriteString("\n")
		}
	}

	// Trailing comment should be treated as leading comment for the next field
	// This is handled in the template by placing it after the field declaration

	return result.String()
}

// formatFieldCommentsWithTrailing formats field comments including trailing comments from previous field
func formatFieldCommentsWithTrailing(field *parser.Field, prevField *parser.Field) string {
	var result strings.Builder

	// Leading comments
	if field.Doc != nil {
		result.WriteString(formatComments(field.Doc))
	}

	// If current field has no leading comments but previous field has trailing comment,
	// use the trailing comment as leading comment for current field
	if field.Doc == nil && prevField != nil && prevField.TrailingComment != nil && *prevField.TrailingComment != "" {
		result.WriteString(*prevField.TrailingComment)
		result.WriteString("\n")
	}

	return result.String()
}

// formatFieldCommentsWithNextTrailing formats field comments including trailing comment for next field
func formatFieldCommentsWithNextTrailing(field *parser.Field, nextField *parser.Field) string {
	var result strings.Builder

	// Leading comments
	if field.Doc != nil {
		result.WriteString(formatComments(field.Doc))
	}

	// If current field has trailing comment, it should be treated as leading comment for next field
	// This is handled in the template by placing it after the field declaration

	return result.String()
}

// formatFieldCommentsWithPreviousTrailing formats field comments including trailing comment from previous field
func formatFieldCommentsWithPreviousTrailing(field *parser.Field, prevField *parser.Field) string {
	var result strings.Builder

	// Leading comments
	if field.Doc != nil {
		comments := formatComments(field.Doc)
		if comments != "" {
			// Add proper indentation to each line
			lines := strings.Split(comments, "\n")
			for i, line := range lines {
				if line != "" {
					if i > 0 {
						result.WriteString("\n")
					}
					result.WriteString("    ")
					result.WriteString(line)
				}
			}
			result.WriteString("\n")
		}
	}

	// If current field has no leading comments but previous field has trailing comment,
	// use the trailing comment as leading comment for current field
	if field.Doc == nil && prevField != nil && prevField.TrailingComment != nil && *prevField.TrailingComment != "" {
		result.WriteString("    ")
		result.WriteString(*prevField.TrailingComment)
		result.WriteString("\n")
	}

	return result.String()
}

// getFieldSpecifier returns field specifier or register specifier
func getFieldSpecifier(field *parser.Field, register *parser.Register) string {
	if field.Specifier != nil {
		return *field.Specifier
	}
	if register.Specifier != nil {
		return *register.Specifier
	}
	return "rw" // default to read-write
}

// getFieldSpecifierFromContext returns field specifier using template context
func getFieldSpecifierFromContext(field *parser.Field, rootContext interface{}) string {
	// Extract register from root context
	if data, ok := rootContext.(struct {
		Device    *parser.Device
		Namespace string
	}); ok {
		// Find the register that contains this field
		for _, reg := range data.Device.Registers {
			for _, f := range reg.Fields {
				if f == field {
					return getFieldSpecifier(field, reg)
				}
			}
		}
	}
	return "rw" // default
}

// hasReadFields checks if register has read-only fields
func hasReadFields(register *parser.Register) bool {
	for _, field := range register.Fields {
		spec := getFieldSpecifier(field, register)
		if spec == "r" || spec == "rw" {
			return true
		}
	}
	return false
}

// hasWriteFields checks if register has write-only fields
func hasWriteFields(register *parser.Register) bool {
	for _, field := range register.Fields {
		spec := getFieldSpecifier(field, register)
		if spec == "w" || spec == "rw" {
			return true
		}
	}
	return false
}

// getReadFields returns read-only fields
func getReadFields(register *parser.Register) []*parser.Field {
	var fields []*parser.Field
	for _, field := range register.Fields {
		spec := getFieldSpecifier(field, register)
		if spec == "r" || spec == "rw" {
			fields = append(fields, field)
		}
	}
	return fields
}

// getWriteFields returns write-only fields
func getWriteFields(register *parser.Register) []*parser.Field {
	var fields []*parser.Field
	for _, field := range register.Fields {
		spec := getFieldSpecifier(field, register)
		if spec == "w" || spec == "rw" {
			fields = append(fields, field)
		}
	}
	return fields
}

// getBitFields returns bit field members
func getBitFields(field *parser.Field) []*parser.BitMember {
	if field.Type.Bitfield != nil {
		return field.Type.Bitfield.Bits
	}
	return nil
}

// getNonBitFields returns non-bit fields
func getNonBitFields(register *parser.Register) []*parser.Field {
	var fields []*parser.Field
	for _, field := range register.Fields {
		if field.Type.Bitfield == nil {
			fields = append(fields, field)
		}
	}
	return fields
}

// getRegisterAddress returns register address (using register number as address)
func getRegisterAddress(register *parser.Register) int {
	return register.Number
}

// findFieldByName finds a field by name in the register
func findFieldByName(register *parser.Register, fieldName string) *parser.Field {
	for _, field := range register.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// findBitFieldMemberByName finds a bit field member by name in the register
func findBitFieldMemberByName(register *parser.Register, memberName string) string {
	for _, field := range register.Fields {
		if field.Type.Bitfield != nil {
			for _, bitMember := range field.Type.Bitfield.Bits {
				if bitMember.Name == memberName {
					return fmt.Sprintf("(%s & %s_%s_bm)", field.Name, field.Name, bitMember.Name)
				}
				// Check if it's a bitfield reference (fieldName_bitMemberName)
				bitFieldRefName := field.Name + "_" + bitMember.Name
				if memberName == bitFieldRefName {
					// Calculate bit shift for the bit field
					shift := ""
					if bitMember.Start > 0 {
						shift = fmt.Sprintf(" >> %d", bitMember.Start)
					}
					return fmt.Sprintf("((%s & %s_%s_bm)%s)", field.Name, field.Name, bitMember.Name, shift)
				}
				// Check if it's a bitmask reference (fieldName_bitMemberName_bm)
				bitmaskName := field.Name + "_" + bitMember.Name + "_bm"
				if memberName == bitmaskName {
					// For bitmask references, we need to calculate the actual bit value
					// For now, return the bitmask itself - this might need adjustment
					return bitmaskName
				}
			}
		}
	}
	return ""
}

// getArraySizeField returns the field or bit field member that provides the size for a variable array
func getArraySizeField(register *parser.Register, arrayField *parser.Field) string {
	if arrayField.Type.Array == nil {
		return ""
	}

	if arrayField.Type.Array.Size.Variable != nil {
		fieldName := *arrayField.Type.Array.Size.Variable

		// First check if it's a regular field
		if field := findFieldByName(register, fieldName); field != nil {
			return field.Name
		}

		// Then check if it's a bit field member
		if bitFieldExpr := findBitFieldMemberByName(register, fieldName); bitFieldExpr != "" {
			return bitFieldExpr
		}
	}

	if arrayField.Type.Array.Size.Type != nil {
		// For type-based size, we need to find a field that matches the type
		// This is a bit more complex and might need additional logic
		return ""
	}

	return ""
}

// getArraySizeFieldForField is a template helper that finds the size field for a given array field
func getArraySizeFieldForField(register *parser.Register, field *parser.Field) string {
	return getArraySizeField(register, field)
}

// getArraySizeFieldName returns the name of the field that provides the size for a variable array
func getArraySizeFieldName(field *parser.Field) string {
	if field.Type.Array == nil || field.Type.Array.Size.Variable == nil {
		return ""
	}
	return *field.Type.Array.Size.Variable
}

// getArraySizeFieldExpression returns the proper C++ expression to get the array size
func getArraySizeFieldExpression(register *parser.Register, field *parser.Field) string {
	return getArraySizeField(register, field)
}

// isBitFieldReference checks if the array size refers to a bit field
func isBitFieldReference(register *parser.Register, field *parser.Field) bool {
	if field.Type.Array == nil || field.Type.Array.Size.Variable == nil {
		return false
	}

	fieldName := *field.Type.Array.Size.Variable

	// Check if it's a bitfield reference (fieldName_bitMemberName)
	for _, f := range register.Fields {
		if f.Type.Bitfield != nil {
			for _, bitMember := range f.Type.Bitfield.Bits {
				bitFieldRefName := f.Name + "_" + bitMember.Name
				if fieldName == bitFieldRefName {
					return true
				}
			}
		}
	}

	return false
}

// Arduino C++ template
const arduinoCppTemplate = `#include <Arduino.h>
#include "bigendian.h"

namespace {{.Namespace}} {

// Register IDs
{{range .Device.Registers}}static constexpr uint8_t Reg_{{.Name}}_ID = {{.Number}};
{{end}}
{{range .Device.Registers}}{{formatComments .Doc}}
struct {{.Name}} {
{{$register := .}}{{$fields := .Fields}}{{range $i, $field := $fields}}{{if eq $i 0}}{{formatFieldComments $field}}{{else}}{{formatFieldCommentsWithPreviousTrailing $field (index $fields (sub $i 1))}}{{end}}{{if $field.Type.Bitfield}}{{range $field.Type.Bitfield.Bits}}    // {{.Name}} bit field (bits {{.Start}}{{if .End}}-{{.End}}{{end}})
    static constexpr uint8_t {{$field.Name}}_{{.Name}}_bm = {{toBitMask .Start .End}};
{{end}}{{end}}    {{toCppFieldType $field}} {{$field.Name}}{{toCppArraySize $field}};{{if $field.TrailingComment}} {{$field.TrailingComment}}{{end}}
{{end}}
{{range .Fields}}{{if and .Type.Array (not .Type.Array.Size.Constant) (not .Type.Array.Size.Variable)}}    // Size field for variable length array {{.Name}}
    {{if .Type.Array.Size.Type}}{{.Type.Array.Size.Type}} {{.Name}}_size;
    {{end}}
{{end}}{{end}}

    // Send read-only fields to wire (register read fields -> wire)
    int send_read_data(uint8_t* buf, size_t size) {
{{if hasReadFields .}}        int offset = 0;
{{range .Fields}}{{if or (eq (getFieldSpecifierFromContext . $) "r") (eq (getFieldSpecifierFromContext . $) "rw")}}{{if .Type.Simple}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{else if .Type.Array}}{{if .Type.Array.Size.Constant}}        if (offset + {{.Type.Array.Size.Constant}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{else}}        // Variable length array - encode size and data
        {{if .Type.Array.Size.Type}}        if (offset + sizeof({{.Type.Array.Size.Type}}) + {{.Name}}_size * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode_varray(buf + offset, {{.Name}}, {{.Name}}_size);
        }
        {{else if .Type.Array.Size.Variable}}        if (offset + sizeof({{.Type.Array.Size.Variable}}) + {{getArraySizeFieldExpression $register .}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode_varray(buf + offset, {{.Name}}, {{getArraySizeFieldExpression $register .}});
        }
        {{end}}{{end}}{{else if .Type.Bitfield}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{end}}{{end}}{{end}}        return offset;
{{else}}        return 0;
{{end}}    }

    // Send write-only fields to wire (register write fields -> wire)
    int send_write_data(uint8_t* buf, size_t size) {
{{if hasWriteFields .}}        int offset = 0;
{{range .Fields}}{{if or (eq (getFieldSpecifierFromContext . $) "w") (eq (getFieldSpecifierFromContext . $) "rw")}}{{if .Type.Simple}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{else if .Type.Array}}{{if .Type.Array.Size.Constant}}        if (offset + {{.Type.Array.Size.Constant}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{else}}        // Variable length array - encode size and data
        {{if .Type.Array.Size.Type}}        if (offset + sizeof({{.Type.Array.Size.Type}}) + {{.Name}}_size * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode_varray(buf + offset, {{.Name}}, {{.Name}}_size);
        }
        {{else if .Type.Array.Size.Variable}}        if (offset + sizeof({{.Type.Array.Size.Variable}}) + {{getArraySizeFieldExpression $register .}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::encode_varray(buf + offset, {{.Name}}, {{getArraySizeFieldExpression $register .}});
        }
        {{end}}{{end}}{{else if .Type.Bitfield}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::encode(buf + offset, {{.Name}});
        }
{{end}}{{end}}{{end}}        return offset;
{{else}}        return 0;
{{end}}    }

    // Get read-only fields from wire (wire -> the register read fields)
    int receive_read_data(uint8_t* buf, size_t size) {
{{if hasReadFields .}}        int offset = 0;
{{range .Fields}}{{if or (eq (getFieldSpecifierFromContext . $) "r") (eq (getFieldSpecifierFromContext . $) "rw")}}{{if .Type.Simple}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{else if .Type.Array}}{{if .Type.Array.Size.Constant}}        if (offset + {{.Type.Array.Size.Constant}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{else}}        // Variable length array - decode size and data
        {{if .Type.Array.Size.Type}}        if (offset + sizeof({{.Type.Array.Size.Type}}) + {{.Name}}_size * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode_varray({{.Name}}, buf + offset, {{.Name}}_size);
        }
        {{else if .Type.Array.Size.Variable}}{{if isBitFieldReference $register .}}        {
            uint8_t size_value = {{getArraySizeFieldExpression $register .}};
            if (offset + sizeof(size_value) + size_value * sizeof({{toCppType .Type.Array.Element}}) <= size) {
                offset += bigendian::decode_varray({{.Name}}, buf + offset, size_value);
            }
        }
        {{else}}        if (offset + sizeof({{.Type.Array.Size.Variable}}) + {{getArraySizeFieldExpression $register .}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode_varray({{.Name}}, buf + offset, {{getArraySizeFieldExpression $register .}});
        }
        {{end}}{{end}}{{end}}{{else if .Type.Bitfield}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{end}}{{end}}{{end}}        return offset;
{{else}}        return 0;
{{end}}    }

    // Getting write-only fields from wire (wire -> the register write fields)
    int receive_write_data(uint8_t* buf, size_t size) {
{{if hasWriteFields .}}        int offset = 0;
{{range .Fields}}{{if or (eq (getFieldSpecifierFromContext . $) "w") (eq (getFieldSpecifierFromContext . $) "rw")}}{{if .Type.Simple}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{else if .Type.Array}}{{if .Type.Array.Size.Constant}}        if (offset + {{.Type.Array.Size.Constant}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{else}}        // Variable length array - decode size and data
        {{if .Type.Array.Size.Type}}        if (offset + sizeof({{.Type.Array.Size.Type}}) + {{.Name}}_size * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode_varray({{.Name}}, buf + offset, {{.Name}}_size);
        }
        {{else if .Type.Array.Size.Variable}}{{if isBitFieldReference $register .}}        {
            uint8_t size_value = {{getArraySizeFieldExpression $register .}};
            if (offset + sizeof(size_value) + size_value * sizeof({{toCppType .Type.Array.Element}}) <= size) {
                offset += bigendian::decode_varray({{.Name}}, buf + offset, size_value);
            }
        }
        {{else}}        if (offset + sizeof({{.Type.Array.Size.Variable}}) + {{getArraySizeFieldExpression $register .}} * sizeof({{toCppType .Type.Array.Element}}) <= size) {
            offset += bigendian::decode_varray({{.Name}}, buf + offset, {{getArraySizeFieldExpression $register .}});
        }
        {{end}}{{end}}{{end}}{{else if .Type.Bitfield}}        if (offset + sizeof({{toCppType .Type}}) <= size) {
            offset += bigendian::decode({{.Name}}, buf + offset);
        }
{{end}}{{end}}{{end}}        return offset;
{{else}}        return 0;
{{end}}    }
};

{{end}}} // namespace {{.Namespace}}
`

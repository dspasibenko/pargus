package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/dspasibenko/pargus/pkg/parser"
)

var headerTmpl = template.Must(template.New("hdr").Funcs(template.FuncMap{
	"cppType": cppType,
	"isReadable": func(field *parser.Field) bool {
		return field.Specifier == nil || *field.Specifier == "r"
	},
	"isWritable": func(field *parser.Field) bool {
		return field.Specifier == nil || *field.Specifier == "w"
	},
	"registerSpecifier": func(register *parser.Register) string {
		if register.Specifier == nil {
			return ""
		}
		return *register.Specifier
	},
	"fieldSpecifier": func(field *parser.Field) string {
		if field.Specifier == nil {
			return ""
		}
		return *field.Specifier
	},
	"fieldType": func(field *parser.Field) string {
		return cppType(field.Type)
	},
	"arraySize": func(field *parser.Field) string {
		if field.Type.Array != nil {
			arrayType := field.Type.Array
			// Check if it's a fixed-size array (not variable-length)
			if len(arrayType.Size) > 0 && !(arrayType.Size[0] >= 'a' && arrayType.Size[0] <= 'z' || arrayType.Size[0] >= 'A' && arrayType.Size[0] <= 'Z') {
				return "[" + arrayType.Size + "]"
			}
		}
		return ""
	},
	"isVariableArray": func(field *parser.Field) bool {
		if field.Type.Array != nil {
			arrayType := field.Type.Array
			return len(arrayType.Size) > 0 && (arrayType.Size[0] >= 'a' && arrayType.Size[0] <= 'z' || arrayType.Size[0] >= 'A' && arrayType.Size[0] <= 'Z')
		}
		return false
	},
	"getArraySizeType": func(field *parser.Field) string {
		if field.Type.Array != nil {
			return simpleTypeToCpp(field.Type.Array.Size)
		}
		return ""
	},
	"isBitField": func(field *parser.Field) bool {
		return field.Type.Bitfield != nil
	},
	"getBitFieldBaseType": func(field *parser.Field) string {
		if field.Type.Bitfield != nil {
			return simpleTypeToCpp(field.Type.Bitfield.Base)
		}
		return ""
	},
	"generateBitMasks": func(field *parser.Field) []string {
		if field.Type.Bitfield == nil {
			return nil
		}

		var masks []string
		for _, bit := range field.Type.Bitfield.Bits {
			if bit.End != nil {
				// Range of bits
				start := bit.Start
				end := *bit.End
				mask := uint64(0)
				for i := start; i <= end; i++ {
					mask |= (1 << i)
				}
				masks = append(masks, fmt.Sprintf("static constexpr %s %s_bm = 0x%X; // bits %d-%d",
					simpleTypeToCpp(field.Type.Bitfield.Base), bit.Name, mask, start, end))
			} else {
				// Single bit
				mask := uint64(1 << bit.Start)
				masks = append(masks, fmt.Sprintf("static constexpr %s %s_bm = 0x%X; // bit %d",
					simpleTypeToCpp(field.Type.Bitfield.Base), bit.Name, mask, bit.Start))
			}
		}
		return masks
	},
	"hasComment": func(comment *string) bool {
		return comment != nil && *comment != ""
	},
	"formatComment": func(comment *string) string {
		if comment == nil || *comment == "" {
			return ""
		}
		// Remove the // prefix if present and add proper C++ comment format
		text := *comment
		if len(text) >= 2 && text[:2] == "//" {
			text = text[2:]
		}
		// Trim leading whitespace
		for len(text) > 0 && (text[0] == ' ' || text[0] == '\t') {
			text = text[1:]
		}
		return "// " + text
	},
	"getReadWriteFields": func(register *parser.Register) []*parser.Field {
		var fields []*parser.Field
		regSpec := ""
		if register.Specifier != nil {
			regSpec = *register.Specifier
		}
		for _, field := range register.Fields {
			fieldSpec := ""
			if field.Specifier != nil {
				fieldSpec = *field.Specifier
			}
			// In read-write register, group by field name patterns
			if regSpec == "rw" && (fieldSpec == "" || fieldSpec == "rw") {
				// Group fields with "rw" in name as read-write
				if len(field.Name) >= 2 && field.Name[:2] == "rw" {
					fields = append(fields, field)
				}
			} else if (fieldSpec == "" && regSpec == "rw") || fieldSpec == "rw" {
				fields = append(fields, field)
			}
		}
		return fields
	},
	"getReadOnlyFields": func(register *parser.Register) []*parser.Field {
		var fields []*parser.Field
		regSpec := ""
		if register.Specifier != nil {
			regSpec = *register.Specifier
		}
		for _, field := range register.Fields {
			fieldSpec := ""
			if field.Specifier != nil {
				fieldSpec = *field.Specifier
			}
			// In read-write register, group by field name patterns
			if regSpec == "rw" && (fieldSpec == "" || fieldSpec == "rw") {
				// Group fields with "read" in name as read-only
				if len(field.Name) >= 4 && field.Name[:4] == "read" {
					fields = append(fields, field)
				}
			} else if (fieldSpec == "" && regSpec == "r") || fieldSpec == "r" {
				fields = append(fields, field)
			}
		}
		return fields
	},
	"getWriteOnlyFields": func(register *parser.Register) []*parser.Field {
		var fields []*parser.Field
		regSpec := ""
		if register.Specifier != nil {
			regSpec = *register.Specifier
		}
		for _, field := range register.Fields {
			fieldSpec := ""
			if field.Specifier != nil {
				fieldSpec = *field.Specifier
			}
			// In read-write register, group by field name patterns
			if regSpec == "rw" && (fieldSpec == "" || fieldSpec == "rw") {
				// Group fields with "write" in name as write-only
				if len(field.Name) >= 5 && field.Name[:5] == "write" {
					fields = append(fields, field)
				}
			} else if (fieldSpec == "" && regSpec == "w") || fieldSpec == "w" {
				fields = append(fields, field)
			}
		}
		return fields
	},
}).Parse(`
#include <Arduino.h>
#include "bigendian.h"

namespace {{.Namespace}} {

{{range .Device.Registers}}
{{- if hasComment .Comment}}
{{formatComment .Comment}}
{{- end}}
static constexpr int Reg_{{.Name}}_ID = {{.Number}};
struct {{.Name}} {
{{- $rwFields := getReadWriteFields .}}
{{- $rFields := getReadOnlyFields .}}
{{- $wFields := getWriteOnlyFields .}}

{{- if gt (len $rwFields) 0}}
    // ==== Read-write fields ====
{{- range $rwFields}}
    {{- if hasComment .Comment}}
    {{formatComment .Comment}}
    {{- end}}
    {{- if isBitField .}}
    // Bit field: {{.Name}}
    {{- range (generateBitMasks .)}}
    {{.}}
    {{- end}}
    {{getBitFieldBaseType .}} {{.Name}};
    {{- else if isVariableArray .}}
    {{fieldType .}} {{.Name}};
    {{getArraySizeType .}} {{.Name}}_size;
    {{- else}}
    {{fieldType .}} {{.Name}}{{arraySize .}};
    {{- end}}
    {{- if hasComment .EndComment}}
    {{formatComment .EndComment}}
    {{- end}}
{{- end}}
{{- end}}

{{- if gt (len $rFields) 0}}
    // ==== Read-only fields ====
{{- range $rFields}}
    {{- if hasComment .Comment}}
    {{formatComment .Comment}}
    {{- end}}
    {{- if isBitField .}}
    // Bit field: {{.Name}}
    {{- range (generateBitMasks .)}}
    {{.}}
    {{- end}}
    {{getBitFieldBaseType .}} {{.Name}};
    {{- else if isVariableArray .}}
    {{fieldType .}} {{.Name}};
    {{getArraySizeType .}} {{.Name}}_size;
    {{- else}}
    {{fieldType .}} {{.Name}}{{arraySize .}};
    {{- end}}
    {{- if hasComment .EndComment}}
    {{formatComment .EndComment}}
    {{- end}}
{{- end}}
{{- end}}

{{- if gt (len $wFields) 0}}
    // ==== Write-only fields ====
{{- range $wFields}}
    {{- if hasComment .Comment}}
    {{formatComment .Comment}}
    {{- end}}
    {{- if isBitField .}}
    // Bit field: {{.Name}}
    {{- range (generateBitMasks .)}}
    {{.}}
    {{- end}}
    {{getBitFieldBaseType .}} {{.Name}};
    {{- else if isVariableArray .}}
    {{fieldType .}} {{.Name}};
    {{getArraySizeType .}} {{.Name}}_size;
    {{- else}}
    {{fieldType .}} {{.Name}}{{arraySize .}};
    {{- end}}
    {{- if hasComment .EndComment}}
    {{formatComment .EndComment}}
    {{- end}}
{{- end}}
{{- end}}

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
    {{- $regSpec := registerSpecifier .}}
    {{- if eq $regSpec "w"}}
        return -1; // write-only register has no read data
    {{- else}}
        int written = 0;
        {{- range .Fields}}
            {{- $fieldSpec := fieldSpecifier .}}
            {{- if or (and (eq $fieldSpec "") (ne $regSpec "w")) (eq $fieldSpec "r")}}
                {{- if isVariableArray .}}
        written += bigendian::encode_varray(buf + written, this->{{.Name}}, this->{{.Name}}_size);
                {{- else}}
        written += bigendian::encode(buf + written, this->{{.Name}});
                {{- end}}
            {{- end}}
        {{- end}}
        return written;
    {{- end}}
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
    {{- $regSpec := registerSpecifier .}}
    {{- if eq $regSpec "r"}}
        return -1; // read-only register has no write data
    {{- else}}
        int written = 0;
        {{- range .Fields}}
            {{- $fieldSpec := fieldSpecifier .}}
            {{- if or (and (eq $fieldSpec "") (ne $regSpec "r")) (eq $fieldSpec "w")}}
                {{- if isVariableArray .}}
        written += bigendian::encode_varray(buf + written, this->{{.Name}}, this->{{.Name}}_size);
                {{- else}}
        written += bigendian::encode(buf + written, this->{{.Name}});
                {{- end}}
            {{- end}}
        {{- end}}
        return written;
    {{- end}}
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
    {{- $regSpec := registerSpecifier .}}
    {{- if eq $regSpec "w"}}
        return -1; // write-only register cannot receive read data
    {{- else}}
        int read = 0;
        {{- range .Fields}}
            {{- $fieldSpec := fieldSpecifier .}}
            {{- if or (and (eq $fieldSpec "") (ne $regSpec "w")) (eq $fieldSpec "r")}}
                {{- if isVariableArray .}}
        read += bigendian::decode_varray(this->{{.Name}}, 1024, buf + read, this->{{.Name}}_size);
                {{- else}}
        read += bigendian::decode(this->{{.Name}}, buf + read);
                {{- end}}
            {{- end}}
        {{- end}}
        return read;
    {{- end}}
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
    {{- $regSpec := registerSpecifier .}}
    {{- if eq $regSpec "r"}}
        return -1; // read-only register cannot receive write data
    {{- else}}
        int read = 0;
        {{- range .Fields}}
            {{- $fieldSpec := fieldSpecifier .}}
            {{- if or (and (eq $fieldSpec "") (ne $regSpec "r")) (eq $fieldSpec "w")}}
                {{- if isVariableArray .}}
        read += bigendian::decode_varray(this->{{.Name}}, 1024, buf + read, this->{{.Name}}_size);
                {{- else}}
        read += bigendian::decode(this->{{.Name}}, buf + read);
                {{- end}}
            {{- end}}
        {{- end}}
        return read;
    {{- end}}
    }
};
{{end}}

} // namespace {{.Namespace}}
`))

// Generate generates C++ header code for Arduino from a Device AST
func Generate(device *parser.Device, namespace string) (string, error) {
	data := struct {
		Device    *parser.Device
		Namespace string
	}{
		Device:    device,
		Namespace: namespace,
	}

	var buf bytes.Buffer
	if err := headerTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// cppType converts a Pargus type to its C++ equivalent
func cppType(t parser.Type) string {
	switch v := t.(type) {
	case *parser.SimpleType:
		return simpleTypeToCpp(v.Name)
	case *parser.ArrayType:
		return arrayTypeToCpp(v)
	case *parser.BitField:
		return bitFieldToCpp(v)
	case *parser.TypeUnion:
		// Handle TypeUnion by checking which field is set
		if v.Bitfield != nil {
			return cppType(v.Bitfield)
		}
		if v.Array != nil {
			return cppType(v.Array)
		}
		if v.Simple != nil {
			return cppType(v.Simple)
		}
		return "void" // fallback
	default:
		return "void"
	}
}

// simpleTypeToCpp converts simple Pargus types to C++ types
func simpleTypeToCpp(pargusType string) string {
	switch pargusType {
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
		return "void"
	}
}

// arrayTypeToCpp converts array types to C++ array declarations
func arrayTypeToCpp(array *parser.ArrayType) string {
	elementType := cppType(array.Element)

	// Check if it's a variable-length array (size is an identifier, not a number)
	if len(array.Size) > 0 && (array.Size[0] >= 'a' && array.Size[0] <= 'z' || array.Size[0] >= 'A' && array.Size[0] <= 'Z') {
		// Variable-length array - return pointer type
		return elementType + "*"
	}

	// Fixed-size array - return just the element type, the template will add the array syntax
	return elementType
}

// bitFieldToCpp converts bit field types to C++ base type
func bitFieldToCpp(bitField *parser.BitField) string {
	// For bit fields, we return just the base type
	// Bit masks will be generated separately
	return simpleTypeToCpp(bitField.Base)
}

package generator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/dspasibenko/pargus/pkg/parser"
)

//
// C++ template
//

const cppTemplate = `
#ifndef __{{.Identifier}}__
#define __{{.Identifier}}__

#include <Arduino.h>
#include "bigendian.h"
 
{{- range .Doc}}
{{.}}
{{- end}}
namespace {{.Namespace}} {

// Register IDs
{{- range .Registers}}
static constexpr uint8_t Reg_{{.Name}}_ID = {{.Number}};
{{- end}}

{{- range .Registers}}
{{range .Doc}}{{.}}
{{end -}}
struct {{.Name}} {
{{- range .Constants}}
    {{- range .Doc}}
    {{.}}
    {{- end}}
    static constexpr {{.Type}} {{.Name}} = {{.Value}};
{{- end}}
{{- range .Fields}}
	{{- range .BitMasks}}
    {{.}}
    {{- end}}
    {{- range .Doc}}
    {{.}}
    {{- end}}
    {{.Decl}}{{if .Trailing}} {{.Trailing}}{{end}}
{{- end}}
    
    // Send read-only fields to wire (register read fields -> wire)
    int send_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
{{- range .Fields}}
{{- if .IsReadable}}
        {{range .SendReadWriteData}}{{.}}
        {{end -}}
{{- end}}
{{- end}}
        return offset;
    }

    // Send write-only fields to wire (register write fields -> wire)
    int send_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
{{- range .Fields}}
{{- if .IsWritable}}
        {{range .SendReadWriteData}}{{.}}
        {{end -}}
{{- end}}
{{- end}}
        return offset;
    }
    
    // Get read-only fields from wire (wire -> the register read fields)
    int receive_read_data(uint8_t* buf, size_t size) {
        int offset = 0;
{{- range .Fields}}
{{- if .IsReadable}}
        {{range .ReceiveReadWriteData}}{{.}}
        {{end -}}
{{- end}}
{{- end}}
        return offset;
    }

    // Get write-only fields from wire (wire -> the register writable fields)
    int receive_write_data(uint8_t* buf, size_t size) {
        int offset = 0;
{{- range .Fields}}
{{- if .IsWritable}}
        {{range .ReceiveReadWriteData}}{{.}}
        {{end -}}
{{- end}}
{{- end}}
        return offset;
    }
};
{{- end}}
} // namespace {{.Namespace}}
#endif // __{{.Identifier}}__
`

//
// Intermediate representation for template
//

type CppDevice struct {
	Doc        []string
	Namespace  string
	Identifier string
	Registers  []CppRegister
}

type CppRegister struct {
	Name      string
	Number    int
	Doc       []string
	Constants []CppConstant
	Fields    []CppField
}

type CppConstant struct {
	Doc   []string
	Name  string
	Type  string
	Value string
}

type CppField struct {
	Doc                  []string
	Name                 string
	BitMasks             []string
	Decl                 string
	IsReadable           bool
	IsWritable           bool
	SendReadWriteData    []string
	ReceiveReadWriteData []string
	Trailing             string
}

//
// Public entry
//

func GenerateCpp(dev *parser.Device, namespace, identifier string) (string, error) {
	tpl, err := template.New("cpp").Parse(cppTemplate)
	if err != nil {
		return "", err
	}

	out := CppDevice{Namespace: namespace, Identifier: identifier}
	out.Doc = flattenComments(dev.Doc)
	for _, reg := range dev.Registers {
		num, _ := strconv.ParseInt(reg.NumberStr, 0, 64)
		cr := CppRegister{
			Name:   reg.Name,
			Number: int(num),
			Doc:    flattenComments(reg.Doc),
		}

		// Process constants
		for _, c := range reg.Body.Constants() {
			cc := CppConstant{
				Doc:   flattenComments(c.Doc),
				Name:  c.Name,
				Type:  toCppTypes(c.Type.Name),
				Value: c.ValueStr,
			}
			cr.Constants = append(cr.Constants, cc)
		}

		for _, f := range reg.Body.Fields() {
			cf := CppField{
				Doc:        flattenComments(f.Doc),
				Name:       f.Name,
				Trailing:   safeString(f.TrailingComment),
				IsReadable: f.Specifier == "r" || f.Specifier == "",
				IsWritable: f.Specifier == "w" || f.Specifier == "",
			}

			switch {
			case f.Type.Simple != nil && f.Type.Simple.IsRegisterRef():
				refRegName := f.Type.Simple.Name
				cf.Decl = fmt.Sprintf("%s %s;", refRegName, f.Name)

				// For RegisterRef, we need to call different methods depending on read/write context
				// The template will filter by IsReadable/IsWritable
				// For readable fields in send_read_data/receive_read_data
				// For writable fields in send_write_data/receive_write_data
				if cf.IsReadable {
					cf.SendReadWriteData = append(cf.SendReadWriteData,
						fmt.Sprintf("offset += %s.send_read_data(buf + offset, size - offset);", f.Name))
					cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData,
						fmt.Sprintf("offset += %s.receive_read_data(buf + offset, size - offset);", f.Name))
				} else if cf.IsWritable {
					cf.SendReadWriteData = append(cf.SendReadWriteData,
						fmt.Sprintf("offset += %s.send_write_data(buf + offset, size - offset);", f.Name))
					cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData,
						fmt.Sprintf("offset += %s.receive_write_data(buf + offset, size - offset);", f.Name))
				}

			case f.Type.Bitfield != nil:
				base := toCppTypes(f.Type.Bitfield.Base)
				cf.Decl = fmt.Sprintf("%s %s;", base, f.Name)
				for _, bm := range f.Type.Bitfield.Bits {
					// Add bit member comments
					bmComments := flattenComments(bm.Doc)
					for _, comment := range bmComments {
						cf.BitMasks = append(cf.BitMasks, comment)
					}

					start, _ := strconv.Atoi(bm.Start)
					end := start
					if bm.End != nil {
						end, _ = strconv.Atoi(*bm.End)
					}
					mask := bitMask(start, end)

					// Add bit mask constant with range info in comment
					bitRange := bm.Start
					if bm.End != nil && *bm.End != bm.Start {
						bitRange = fmt.Sprintf("%s-%s", bm.Start, *bm.End)
					}
					cf.BitMasks = append(cf.BitMasks,
						fmt.Sprintf("// %s bit field (bits %s)", bm.Name, bitRange))
					cf.BitMasks = append(cf.BitMasks,
						fmt.Sprintf("static constexpr %s %s_%s_bm = 0x%X;",
							base, f.Name, bm.Name, mask))
				}
				cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("if (offset + sizeof(%s) <= size) {", base))
				cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    offset += bigendian::encode(buf + offset, %s);", f.Name))
				cf.SendReadWriteData = append(cf.SendReadWriteData, "}")
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("if (offset + sizeof(%s) <= size) {", base))
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    offset += bigendian::decode(%s, buf + offset);", f.Name))
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "}")

			case f.Type.Array != nil:
				elem := toCppTypes(f.Type.Array.Type.Name)
				if f.Type.Array.Size.Constant != nil {
					sz := *f.Type.Array.Size.Constant
					cf.Decl = fmt.Sprintf("%s %s[%s];", elem, f.Name, sz)
					cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("if (offset + sizeof(%s)*%s <= size) {", elem, sz))
					cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    offset += bigendian::encode(buf + offset, %s);", f.Name))
					cf.SendReadWriteData = append(cf.SendReadWriteData, "}")
					cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("if (offset + sizeof(%s)*%s <= size) {", elem, sz))
					cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    offset += bigendian::decode(%s, buf + offset);", f.Name))
					cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "}")
				} else {
					cf.Decl = fmt.Sprintf("%s* %s;", elem, f.Name)
					szFieldName := *f.Type.Array.Size.Variable
					field, bm := reg.FindFieldByName(szFieldName, len(cr.Fields))
					if bm != nil {
						// this is the bit mask field
						cf.SendReadWriteData = append(cf.SendReadWriteData, "{")
						cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    %s elems = (%s&%s)>>%d;", toCppTypes(field.Type.Bitfield.Base),
							field.Name, fmt.Sprintf("%s_%s_bm", field.Name, bm.Name), bm.StartBit()))
						cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    if (offset + sizeof(%s)*elems <= size) {", elem))
						cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("         offset += bigendian::encode_varray(buf + offset, %s, elems);", f.Name))
						cf.SendReadWriteData = append(cf.SendReadWriteData, "    }")
						cf.SendReadWriteData = append(cf.SendReadWriteData, "}")

						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "{")
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    %s elems = (%s&%s)>>%d;", toCppTypes(field.Type.Bitfield.Base),
							field.Name, fmt.Sprintf("%s_%s_bm", field.Name, bm.Name), bm.StartBit()))
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    if (offset + sizeof(%s)*elems <= size) {", elem))
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("         offset += bigendian::decode_varray(%s, buf + offset, elems);", f.Name))
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "    }")
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "}")
					} else {
						// this is the regular field
						cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("if (offset + sizeof(%s)*%s <= size) {", elem, field.Name))
						cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    offset += bigendian::encode_varray(buf + offset, %s, %s);", f.Name, field.Name))
						cf.SendReadWriteData = append(cf.SendReadWriteData, "}")
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("if (offset + sizeof(%s)*%s <= size) {", elem, field.Name))
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    offset += bigendian::decode_varray(%s, buf + offset, %s);", f.Name, field.Name))
						cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "}")

					}
				}

			case f.Type.Simple != nil:
				elem := toCppTypes(f.Type.Simple.Name)
				cf.Decl = fmt.Sprintf("%s %s;", elem, f.Name)
				cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("if (offset + sizeof(%s) <= size) {", elem))
				cf.SendReadWriteData = append(cf.SendReadWriteData, fmt.Sprintf("    offset += bigendian::encode(buf + offset, %s);", f.Name))
				cf.SendReadWriteData = append(cf.SendReadWriteData, "}")
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("if (offset + sizeof(%s) <= size) {", elem))
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, fmt.Sprintf("    offset += bigendian::decode(%s, buf + offset);", f.Name))
				cf.ReceiveReadWriteData = append(cf.ReceiveReadWriteData, "}")

			default:
				cf.Decl = fmt.Sprintf("/* unsupported field %s */", f.Name)
			}

			cr.Fields = append(cr.Fields, cf)
		}
		out.Registers = append(out.Registers, cr)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, out); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()) + "\n", nil
}

//
// Helpers
//

func toCppTypes(typ string) string {
	switch typ {
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
		return typ + "error"
	}
}

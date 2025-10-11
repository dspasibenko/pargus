package generator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/dspasibenko/pargus/pkg/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const goTemplate = `
package {{.Package}}

import (
    "encoding/binary"
)

{{- range .Doc}}
{{.}}
{{- end}}

{{- range .Registers}}
{{range .Doc}}{{.}}
{{end -}}
{{ $regName := .Name }}
type {{.Name}} struct {
{{- range .Fields}}
    {{- range .Doc}}
    {{.}}
    {{- end}}
    {{.Decl}} {{if .Trailing}} {{.Trailing}}{{end}}
{{- end}}
}

{{- range .Constants}}
{{range .Doc}}{{.}}
{{end -}}
const {{.Name}} {{.Type}} = {{.Value}}
{{- end}}

{{- range .Fields}}
{{- range .BitMasks}}
{{.}}
{{- end}}
{{- end}}

func (r *{{.Name}}) SendReadData(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsReadable}}
    {{range .SendReadWriteData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

func (r *{{.Name}}) SendWriteData(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsWritable}}
    {{range .SendReadWriteData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

func (r *{{.Name}}) ReceiveReadData(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsReadable}}
    {{range .ReceiveReadWriteData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

func (r *{{.Name}}) ReceiveWriteData(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsWritable}}
    {{range .ReceiveReadWriteData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

{{- range .Fields}}
// Getter and Setter for {{.Name}}
func (r *{{$regName}}) Get{{.CapitalizedName}}() {{.Type}} {
    return r.{{.Name}}
}

func (r *{{$regName}}) Set{{.CapitalizedName}}(v {{.Type}}) {
    r.{{.Name}} = v
}
{{- end}}

{{- end}}

func putUint8(b []byte, v uint8) {
	_ = b[0]
	b[0] = v
}

func putUint16(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

func putUint32(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

func putUint64(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}

func getUint8(b []byte) uint8 {
	return b[0]
}

func getUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func getUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func getUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
`

type GoDevice struct {
	Doc       []string
	Package   string
	Registers []GoRegister
}

type GoRegister struct {
	Name      string
	Doc       []string
	Constants []GoConstant
	Fields    []GoField
}

type GoConstant struct {
	Doc   []string
	Name  string
	Type  string
	Value string
}

type GoField struct {
	Doc                  []string
	Name                 string
	CapitalizedName      string
	Decl                 string
	Type                 string
	BitMasks             []string
	IsReadable           bool
	IsWritable           bool
	SendReadWriteData    []string
	ReceiveReadWriteData []string
	Trailing             string
}

func GenerateGo(dev *parser.Device, pkg string) (string, error) {
	tpl, err := template.New("go").Parse(goTemplate)
	if err != nil {
		return "", err
	}

	out := GoDevice{Package: pkg}
	out.Doc = flattenComments(dev.Doc)

	for _, reg := range dev.Registers {
		gr := GoRegister{
			Name: reg.Name,
			Doc:  flattenComments(reg.Doc),
		}

		// Process constants
		for _, c := range reg.Body.Constants() {
			gc := GoConstant{
				Doc:   flattenComments(c.Doc),
				Name:  fmt.Sprintf("%s_%s", reg.Name, c.Name),
				Type:  toGoTypes(c.Type.Name),
				Value: c.ValueStr,
			}
			gr.Constants = append(gr.Constants, gc)
		}

		for _, f := range reg.Body.Fields() {
			gf := GoField{
				Doc:             flattenComments(f.Doc),
				Name:            f.Name,
				CapitalizedName: cases.Title(language.English).String(f.Name),
				Trailing:        safeString(f.TrailingComment),
				IsReadable:      f.Specifier == "r" || f.Specifier == "",
				IsWritable:      f.Specifier == "w" || f.Specifier == "",
			}

			switch {
			case f.Type.Simple != nil && f.Type.Simple.IsRegisterRef():
				refRegName := f.Type.Simple.Name
				gf.Type = refRegName
				gf.Decl = fmt.Sprintf("%s %s", f.Name, refRegName)

				// For RegisterRef, call different methods depending on read/write context
				if gf.IsReadable {
					gf.SendReadWriteData = append(gf.SendReadWriteData,
						fmt.Sprintf("offset += r.%s.SendReadData(buf[offset:])", f.Name))
					gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
						fmt.Sprintf("offset += r.%s.ReceiveReadData(buf[offset:])", f.Name))
				} else if gf.IsWritable {
					gf.SendReadWriteData = append(gf.SendReadWriteData,
						fmt.Sprintf("offset += r.%s.SendWriteData(buf[offset:])", f.Name))
					gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
						fmt.Sprintf("offset += r.%s.ReceiveWriteData(buf[offset:])", f.Name))
				}

			case f.Type.Bitfield != nil:
				base := toGoTypes(f.Type.Bitfield.Base)
				gf.Type = base
				gf.Decl = fmt.Sprintf("%s %s", f.Name, base)
				for _, bm := range f.Type.Bitfield.Bits {
					// Add bit member comments
					bmComments := flattenComments(bm.Doc)
					for _, comment := range bmComments {
						gf.BitMasks = append(gf.BitMasks, comment)
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
					gf.BitMasks = append(gf.BitMasks,
						fmt.Sprintf("// %s bit field (bits %s)", bm.Name, bitRange))
					gf.BitMasks = append(gf.BitMasks,
						fmt.Sprintf("const %s_%s_%s_bm %s = 0x%X", reg.Name,
							f.Name, bm.Name, base, mask))
				}
				size := typeSize(base)
				gf.SendReadWriteData = append(gf.SendReadWriteData,
					fmt.Sprintf("putUint%d(buf[offset:], uint%d(r.%s))", size*8, size*8, f.Name),
					fmt.Sprintf("offset += %d", size))
				gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
					fmt.Sprintf("r.%s = %s(getUint%d(buf[offset:]))", f.Name, base, size*8),
					fmt.Sprintf("offset += %d", size))

			case f.Type.Array != nil && f.Type.Array.Size.Constant != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				sz := *f.Type.Array.Size.Constant
				gf.Type = fmt.Sprintf("[%s]%s", sz, elem)
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)
				gf.SendReadWriteData = append(gf.SendReadWriteData,
					fmt.Sprintf("for i := 0; i < %s; i++ {", sz),
					fmt.Sprintf("    putUint%d(buf[offset:], uint%d(r.%s[i]))", elemSize*8, elemSize*8, f.Name),
					fmt.Sprintf("    offset += %d", elemSize),
					"}")
				gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
					fmt.Sprintf("for i := 0; i < %s; i++ {", sz),
					fmt.Sprintf("    r.%s[i] = %s(getUint%d(buf[offset:]))", f.Name, elem, elemSize*8),
					fmt.Sprintf("    offset += %d", elemSize),
					"}")

			case f.Type.Array != nil && f.Type.Array.Size.Variable != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				refField := *f.Type.Array.Size.Variable
				gf.Type = "[]" + elem
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)

				fld, bm := reg.FindFieldByName(refField, len(gr.Fields))
				gf.SendReadWriteData = []string{"{"}
				gf.ReceiveReadWriteData = []string{"{"}
				if bm != nil {
					gf.SendReadWriteData = append(gf.SendReadWriteData,
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", refField, reg.Name, fld.Name, bm.Name, bm.StartBit()),
					)
					gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", refField, reg.Name, fld.Name, bm.Name, bm.StartBit()),
					)
				} else {
					gf.SendReadWriteData = append(gf.SendReadWriteData,
						fmt.Sprintf("    elems := r.%s", refField),
					)
					gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
						fmt.Sprintf("    elems := r.%s", refField),
					)
				}
				gf.SendReadWriteData = append(gf.SendReadWriteData,
					"    for i := 0; i < int(elems); i++ {",
					fmt.Sprintf("        putUint%d(buf[offset:], uint%d(r.%s[i]))", elemSize*8, elemSize*8, f.Name),
					fmt.Sprintf("        offset += %d", elemSize),
					"    }")
				gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
					fmt.Sprintf("	r.%s = []%s{}", f.Name, elem),
					"    for i := 0; i < int(elems); i++ {",
					fmt.Sprintf("        r.%s = append(r.%s, %s(getUint%d(buf[offset:])))", f.Name, f.Name, elem, elemSize*8),
					fmt.Sprintf("        offset += %d", elemSize),
					"    }")
				gf.SendReadWriteData = append(gf.SendReadWriteData, "}")
				gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData, "}")

			case f.Type.Simple != nil:
				elem := toGoTypes(f.Type.Simple.Name)
				gf.Type = elem
				gf.Decl = fmt.Sprintf("%s %s", f.Name, elem)
				size := typeSize(elem)
				gf.SendReadWriteData = append(gf.SendReadWriteData,
					fmt.Sprintf("putUint%d(buf[offset:], uint%d(r.%s))", size*8, size*8, f.Name),
					fmt.Sprintf("offset += %d", size))
				gf.ReceiveReadWriteData = append(gf.ReceiveReadWriteData,
					fmt.Sprintf("r.%s = %s(getUint%d(buf[offset:]))", f.Name, elem, size*8),
					fmt.Sprintf("offset += %d", size))

			default:
				gf.Type = "interface{}"
				gf.Decl = fmt.Sprintf("// unsupported field %s", f.Name)
			}

			gr.Fields = append(gr.Fields, gf)
		}
		out.Registers = append(out.Registers, gr)
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

func toGoTypes(typ string) string {
	switch typ {
	case "int8":
		return "int8"
	case "uint8":
		return "uint8"
	case "int16":
		return "int16"
	case "uint16":
		return "uint16"
	case "int32":
		return "int32"
	case "uint32":
		return "uint32"
	case "int64":
		return "int64"
	case "uint64":
		return "uint64"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	default:
		return "interface{}"
	}
}

func typeSize(goType string) int {
	switch goType {
	case "int8", "uint8":
		return 1
	case "int16", "uint16":
		return 2
	case "int32", "uint32", "float32":
		return 4
	case "int64", "uint64", "float64":
		return 8
	default:
		return 0
	}
}

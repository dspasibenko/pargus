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
// This is auto-generated file. DO NOT EDIT. Use pargus compiler to regenerate it. 
package {{.Package}}

import (
    "encoding/binary"
)

{{- range .Doc}}
{{.}}
{{- end}}

{{- range .Registers}}{{ $regName := .Name }}
{{range .Doc}}{{.}}
{{end -}}
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
{{- end}}


{{- range .Registers}}
{{ $regName := .Name }}
// ================= {{.Name}} implementation =================
// The {{.Name}} register's ID
func (r *{{.Name}}) ID() uint8 {
	return {{.ID}}
}

// BufSize4Read returns the buffer size required for read fields serialization
func (r *{{.Name}}) BufSize4Read() int {
    size := {{.BufSize4ReadConst}}
{{- range .Fields}}
{{- if .IsReadable}}
{{- if .BufSize4ReadExpr}}
    size += {{.BufSize4ReadExpr}}
{{- end}}
{{- end}}
{{- end}}
    return size
}

// BufSize4Write returns the buffer size required for write fields serialization
func (r *{{.Name}}) BufSize4Write() int {
    size := {{.BufSize4WriteConst}}
{{- range .Fields}}
{{- if .IsWritable}}
{{- if .BufSize4WriteExpr}}
    size += {{.BufSize4WriteExpr}}
{{- end}}
{{- end}}
{{- end}}
    return size
}

// SerializeRead serializes read data to the wire buffer
func (r *{{.Name}}) SerializeRead(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsReadable}}
    {{range .SerializeData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

// SerializeWrite serializes write data to the wire buffer
func (r *{{.Name}}) SerializeWrite(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsWritable}}
    {{range .SerializeData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

// DeserializeRead deserializes read data into the register
func (r *{{.Name}}) DeserializeRead(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsReadable}}
    {{range .DeserializeData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

// DeserializeWrite deserializes write data into the register
func (r *{{.Name}}) DeserializeWrite(buf []byte) int {
    offset := 0
{{- range .Fields}}
{{- if .IsWritable}}
    {{range .DeserializeData}}{{.}}
    {{end -}}
{{- end}}
{{- end}}
    return offset
}

{{- range .Fields}}
// Get{{.CapitalizedName}} returns value for {{.Name}}
func (r *{{$regName}}) Get{{.CapitalizedName}}() {{.Type}} {
    return r.{{.Name}}
}

// Set{{.CapitalizedName}} sets value for {{.Name}}
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
	Name               string
	ID                 uint8
	Doc                []string
	Constants          []GoConstant
	Fields             []GoField
	BufSize4ReadConst  int
	BufSize4WriteConst int
}

type GoConstant struct {
	Doc   []string
	Name  string
	Type  string
	Value string
}

type GoField struct {
	Doc               []string
	Name              string
	CapitalizedName   string
	Decl              string
	Type              string
	BitMasks          []string
	IsReadable        bool
	IsWritable        bool
	SerializeData     []string
	DeserializeData   []string
	Trailing          string
	BufSize4ReadExpr  string // Expression for variable size (empty if constant)
	BufSize4WriteExpr string // Expression for variable size (empty if constant)
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
			ID:   uint8(reg.Number()),
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
					gf.SerializeData = append(gf.SerializeData,
						fmt.Sprintf("offset += r.%s.SerializeRead(buf[offset:])", f.Name))
					gf.DeserializeData = append(gf.DeserializeData,
						fmt.Sprintf("offset += r.%s.DeserializeRead(buf[offset:])", f.Name))
					gf.BufSize4ReadExpr = fmt.Sprintf("r.%s.BufSize4Read()", f.Name)
				}
				if gf.IsWritable {
					gf.SerializeData = append(gf.SerializeData,
						fmt.Sprintf("offset += r.%s.SerializeWrite(buf[offset:])", f.Name))
					gf.DeserializeData = append(gf.DeserializeData,
						fmt.Sprintf("offset += r.%s.DeserializeWrite(buf[offset:])", f.Name))
					gf.BufSize4WriteExpr = fmt.Sprintf("r.%s.BufSize4Write()", f.Name)
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
				gf.SerializeData = append(gf.SerializeData,
					fmt.Sprintf("putUint%d(buf[offset:], uint%d(r.%s))", size*8, size*8, f.Name),
					fmt.Sprintf("offset += %d", size))
				gf.DeserializeData = append(gf.DeserializeData,
					fmt.Sprintf("r.%s = %s(getUint%d(buf[offset:]))", f.Name, base, size*8),
					fmt.Sprintf("offset += %d", size))

				// Bitfield buffer size is constant - add directly to register
				if gf.IsReadable {
					gr.BufSize4ReadConst += size
				}
				if gf.IsWritable {
					gr.BufSize4WriteConst += size
				}

			case f.Type.Array != nil && f.Type.Array.Size.Constant != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				sz := *f.Type.Array.Size.Constant
				gf.Type = fmt.Sprintf("[%s]%s", sz, elem)
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)
				gf.SerializeData = append(gf.SerializeData,
					fmt.Sprintf("for i := 0; i < %s; i++ {", sz),
					fmt.Sprintf("    putUint%d(buf[offset:], uint%d(r.%s[i]))", elemSize*8, elemSize*8, f.Name),
					fmt.Sprintf("    offset += %d", elemSize),
					"}")
				gf.DeserializeData = append(gf.DeserializeData,
					fmt.Sprintf("for i := 0; i < %s; i++ {", sz),
					fmt.Sprintf("    r.%s[i] = %s(getUint%d(buf[offset:]))", f.Name, elem, elemSize*8),
					fmt.Sprintf("    offset += %d", elemSize),
					"}")

				// Constant array buffer size: array size * element size - add directly to register
				szInt, _ := strconv.Atoi(sz)
				bufSizeConst := szInt * elemSize
				if gf.IsReadable {
					gr.BufSize4ReadConst += bufSizeConst
				}
				if gf.IsWritable {
					gr.BufSize4WriteConst += bufSizeConst
				}

			case f.Type.Array != nil && f.Type.Array.Size.Variable != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				refField := *f.Type.Array.Size.Variable
				gf.Type = "[]" + elem
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)

				fld, bm := reg.FindFieldByName(refField, len(gr.Fields))
				gf.SerializeData = []string{"{"}
				gf.DeserializeData = []string{"{"}

				var bufSizeExpr string
				if bm != nil {
					gf.SerializeData = append(gf.SerializeData,
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", refField, reg.Name, fld.Name, bm.Name, bm.StartBit()),
					)
					gf.DeserializeData = append(gf.DeserializeData,
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", refField, reg.Name, fld.Name, bm.Name, bm.StartBit()),
					)
					// Variable array buffer size: element size * bitfield value
					bufSizeExpr = fmt.Sprintf("(int((r.%s&%s_%s_%s_bm)>>%d) * %d)",
						refField, reg.Name, fld.Name, bm.Name, bm.StartBit(), elemSize)
				} else {
					gf.SerializeData = append(gf.SerializeData,
						fmt.Sprintf("    elems := r.%s", refField),
					)
					gf.DeserializeData = append(gf.DeserializeData,
						fmt.Sprintf("    elems := r.%s", refField),
					)
					// Variable array buffer size: element size * reference field
					bufSizeExpr = fmt.Sprintf("(int(r.%s) * %d)", refField, elemSize)
				}
				gf.SerializeData = append(gf.SerializeData,
					"    for i := 0; i < int(elems); i++ {",
					fmt.Sprintf("        putUint%d(buf[offset:], uint%d(r.%s[i]))", elemSize*8, elemSize*8, f.Name),
					fmt.Sprintf("        offset += %d", elemSize),
					"    }")
				gf.DeserializeData = append(gf.DeserializeData,
					fmt.Sprintf("	r.%s = []%s{}", f.Name, elem),
					"    for i := 0; i < int(elems); i++ {",
					fmt.Sprintf("        r.%s = append(r.%s, %s(getUint%d(buf[offset:])))", f.Name, f.Name, elem, elemSize*8),
					fmt.Sprintf("        offset += %d", elemSize),
					"    }")
				gf.SerializeData = append(gf.SerializeData, "}")
				gf.DeserializeData = append(gf.DeserializeData, "}")

				if gf.IsReadable {
					gf.BufSize4ReadExpr = bufSizeExpr
				}
				if gf.IsWritable {
					gf.BufSize4WriteExpr = bufSizeExpr
				}

			case f.Type.Simple != nil:
				elem := toGoTypes(f.Type.Simple.Name)
				gf.Type = elem
				gf.Decl = fmt.Sprintf("%s %s", f.Name, elem)
				size := typeSize(elem)
				gf.SerializeData = append(gf.SerializeData,
					fmt.Sprintf("putUint%d(buf[offset:], uint%d(r.%s))", size*8, size*8, f.Name),
					fmt.Sprintf("offset += %d", size))
				gf.DeserializeData = append(gf.DeserializeData,
					fmt.Sprintf("r.%s = %s(getUint%d(buf[offset:]))", f.Name, elem, size*8),
					fmt.Sprintf("offset += %d", size))

				// Simple type buffer size is constant - add directly to register
				if gf.IsReadable {
					gr.BufSize4ReadConst += size
				}
				if gf.IsWritable {
					gr.BufSize4WriteConst += size
				}

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

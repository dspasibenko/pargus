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
    "fmt"
)

{{- range .Doc}}
{{.}}
{{- end}}

{{- range .Registers}}{{ $regName := .Name }}
{{range .Doc}}{{.}}
{{end -}}
type {{.Name}} struct {
{{- range .fields}}
    {{- range .Doc}}
    {{.}}
    {{- end}}
    {{.Decl}} {{if .Trailing}} {{.Trailing}}{{end}}
{{- end}}
}

{{- range .Constants}}
{{range .Doc}}{{.}}
{{end -}}
const {{.Name}} {{.tp}} = {{.Value}}
{{- end}}

{{- range .fields}}
{{- range .BitMasks}}
{{.}}
{{- end}}
{{- end}}
{{- end}}


{{- range .Registers}}
{{ $regName := .Name }}
// ================= {{.Name}} implementation =================
// The {{.Name}} register's id
func (r *{{.Name}}) id() uint8 {
	return {{.id}}
}

// BufSize4Read returns the buffer size required for read fields serialization
func (r *{{.Name}}) BufSize4Read() int {
    size := {{.BufSize4ReadConst}}
{{- range .fields}}
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
{{- range .fields}}
{{- if .IsWritable}}
{{- if .BufSize4WriteExpr}}
    size += {{.BufSize4WriteExpr}}
{{- end}}
{{- end}}
{{- end}}
    return size
}

// Check validates the consistency of variable-length arrays with their size fields
func (r *{{.Name}}) Check() error {
{{- range .fields}}
{{- range .ConsistencyChecks}}
    {{.}}
{{- end}}
{{- end}}
    return nil
}

// SerializeRead serializes read data to the wire buffer
func (r *{{.Name}}) SerializeRead(buf []byte) (int, error) {
    if err := r.Check(); err != nil {
        return 0, err
    }
    offset := 0
{{- range .fields}}{{- if .SerializeReadData}}
    {{range .SerializeReadData}}{{.}}
    {{end -}}
{{- end}}{{- end}}
    return offset, nil
}

// SerializeWrite serializes write data to the wire buffer
func (r *{{.Name}}) SerializeWrite(buf []byte) (int, error) {
    if err := r.Check(); err != nil {
        return 0, err
    }
    offset := 0
{{- range .fields}}{{- if .SerializeWriteData}}
    {{range .SerializeWriteData}}{{.}}
    {{end -}}
{{- end}}{{- end}}
    return offset, nil
}

// DeserializeRead deserializes read data into the register
func (r *{{.Name}}) DeserializeRead(buf []byte) (int, error) {
    offset := 0
{{- range .fields}}{{- if .DeserializeReadData}}
    {{range .DeserializeReadData}}{{.}}
    {{end -}}
{{- end}}{{- end}}
    return offset, nil
}

// DeserializeWrite deserializes write data into the register
func (r *{{.Name}}) DeserializeWrite(buf []byte) (int, error) {
    offset := 0
{{- range .fields}}{{- if .DeserializeWriteData}}
    {{range .DeserializeWriteData}}{{.}}
    {{end -}}
{{- end}}{{- end}}
    return offset, nil
}

{{- range .fields}}
// Get{{.CapitalizedName}} returns value for {{.Name}}
func (r *{{$regName}}) Get{{.CapitalizedName}}() {{.tp}} {
    return r.{{.Name}}
}

// Set{{.CapitalizedName}} sets value for {{.Name}}
func (r *{{$regName}}) Set{{.CapitalizedName}}(v {{.tp}}) {
    r.{{.Name}} = v
}
{{- end}}

{{- end}}

type Integer interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func putNumber[T Integer](b []byte, v T) error {
	size := binary.Size(v)
	if len(b) < size {
		return fmt.Errorf("buffer too small: need %d bytes, have %d", size, len(b))
	}
	
	switch size {
	case 1:
		b[0] = byte(v)
	case 2:
		binary.BigEndian.PutUint16(b, uint16(v))
	case 4:
		binary.BigEndian.PutUint32(b, uint32(v))
	case 8:
		binary.BigEndian.PutUint64(b, uint64(v))
	default:
		return fmt.Errorf("unsupported type size: %d", size)
	}
	return nil
}

func getNumber[T Integer](b []byte, res *T) error {
	size := binary.Size(*res)
	if len(b) < size {
		return fmt.Errorf("buffer too small: need %d bytes, have %d", size, len(b))
	}
	
	switch size {
	case 1:
		*res = T(b[0])
	case 2:
		*res = T(binary.BigEndian.Uint16(b))
	case 4:
		*res = T(binary.BigEndian.Uint32(b))
	case 8:
		*res = T(binary.BigEndian.Uint64(b))
	default:
		return fmt.Errorf("unsupported type size: %d", size)
	}
	return nil
}

func putSlice[T Integer](b []byte, s []T) error {
	if len(s) == 0 {
		return nil
	}
	
	size := binary.Size(s[0])
	totalSize := size * len(s)
	if len(b) < totalSize {
		return fmt.Errorf("buffer too small: need %d bytes, have %d", totalSize, len(b))
	}
	
	switch size {
	case 1:
		for i, val := range s {
			b[i] = byte(val)
		}
	case 2:
		for _, val := range s {
			binary.BigEndian.PutUint16(b, uint16(val))
			b = b[2:]
		}
	case 4:
		for _, val := range s {
			binary.BigEndian.PutUint32(b, uint32(val))
			b = b[4:]
		}
	case 8:
		for _, val := range s {
			binary.BigEndian.PutUint64(b, uint64(val))
			b = b[8:]
		}
	default:
		return fmt.Errorf("unsupported type size: %d", size)
	}
	return nil
}

func getSlice[T Integer](b []byte, s []T) error {
	if len(s) == 0 {
		return nil
	}
	
	size := binary.Size(s[0])
	totalSize := size * len(s)
	if len(b) < totalSize {
		return fmt.Errorf("buffer too small: need %d bytes, have %d", totalSize, len(b))
	}
	
	switch size {
	case 1:
		for i := range s {
			s[i] = T(b[i])
		}
	case 2:
		for i := range s {
			s[i] = T(binary.BigEndian.Uint16(b))
			b = b[2:]
		}
	case 4:
		for i := range s {
			s[i] = T(binary.BigEndian.Uint32(b))
			b = b[4:]
		}
	case 8:
		for i := range s {
			s[i] = T(binary.BigEndian.Uint64(b))
			b = b[8:]
		}
	default:
		return fmt.Errorf("unsupported type size: %d", size)
	}
	return nil
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
	Doc                  []string
	Name                 string
	CapitalizedName      string
	Decl                 string
	Type                 string
	BitMasks             []string
	IsReadable           bool
	IsWritable           bool
	SerializeReadData    []string // Code for SerializeRead function
	SerializeWriteData   []string // Code for SerializeWrite function
	DeserializeReadData  []string // Code for DeserializeRead function
	DeserializeWriteData []string // Code for DeserializeWrite function
	Trailing             string
	BufSize4ReadExpr     string   // Expression for variable size (empty if constant)
	BufSize4WriteExpr    string   // Expression for variable size (empty if constant)
	ConsistencyChecks    []string // Checks for variable-length arrays
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

				// For RegisterRef, populate the appropriate contexts
				if gf.IsReadable {
					gf.SerializeReadData = append(gf.SerializeReadData,
						fmt.Sprintf("if n, err := r.%s.SerializeRead(buf[offset:]); err != nil {", f.Name),
						"    return offset, err",
						"} else {",
						"    offset += n",
						"}")
					gf.DeserializeReadData = append(gf.DeserializeReadData,
						fmt.Sprintf("if n, err := r.%s.DeserializeRead(buf[offset:]); err != nil {", f.Name),
						"    return offset, err",
						"} else {",
						"    offset += n",
						"}")
					gf.BufSize4ReadExpr = fmt.Sprintf("r.%s.BufSize4Read()", f.Name)
				}
				if gf.IsWritable {
					gf.SerializeWriteData = append(gf.SerializeWriteData,
						fmt.Sprintf("if n, err := r.%s.SerializeWrite(buf[offset:]); err != nil {", f.Name),
						"    return offset, err",
						"} else {",
						"    offset += n",
						"}")
					gf.DeserializeWriteData = append(gf.DeserializeWriteData,
						fmt.Sprintf("if n, err := r.%s.DeserializeWrite(buf[offset:]); err != nil {", f.Name),
						"    return offset, err",
						"} else {",
						"    offset += n",
						"}")
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
				serCode := []string{
					fmt.Sprintf("if err := putNumber(buf[offset:], r.%s); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %d", size),
				}
				deserCode := []string{
					fmt.Sprintf("if err := getNumber(buf[offset:], &r.%s); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %d", size),
				}

				// Bitfield code is the same for read and write contexts
				if gf.IsReadable {
					gf.SerializeReadData = append(gf.SerializeReadData, serCode...)
					gf.DeserializeReadData = append(gf.DeserializeReadData, deserCode...)
					gr.BufSize4ReadConst += size
				}
				if gf.IsWritable {
					gf.SerializeWriteData = append(gf.SerializeWriteData, serCode...)
					gf.DeserializeWriteData = append(gf.DeserializeWriteData, deserCode...)
					gr.BufSize4WriteConst += size
				}

			case f.Type.Array != nil && f.Type.Array.Size.Constant != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				sz := *f.Type.Array.Size.Constant
				gf.Type = fmt.Sprintf("[%s]%s", sz, elem)
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)
				serCode := []string{
					fmt.Sprintf("if err := putSlice(buf[offset:], r.%s[:]); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %s * %d", sz, elemSize),
				}
				deserCode := []string{
					fmt.Sprintf("if err := getSlice(buf[offset:], r.%s[:]); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %s * %d", sz, elemSize),
				}

				// Constant array buffer size: array size * element size - add directly to register
				szInt, _ := strconv.Atoi(sz)
				bufSizeConst := szInt * elemSize
				if gf.IsReadable {
					gf.SerializeReadData = append(gf.SerializeReadData, serCode...)
					gf.DeserializeReadData = append(gf.DeserializeReadData, deserCode...)
					gr.BufSize4ReadConst += bufSizeConst
				}
				if gf.IsWritable {
					gf.SerializeWriteData = append(gf.SerializeWriteData, serCode...)
					gf.DeserializeWriteData = append(gf.DeserializeWriteData, deserCode...)
					gr.BufSize4WriteConst += bufSizeConst
				}

			case f.Type.Array != nil && f.Type.Array.Size.Variable != nil:
				elem := toGoTypes(f.Type.Array.Type.Name)
				refField := *f.Type.Array.Size.Variable
				gf.Type = "[]" + elem
				gf.Decl = fmt.Sprintf("%s %s", f.Name, gf.Type)
				elemSize := typeSize(elem)

				fld, bm := reg.FindFieldByName(refField, len(gr.Fields))

				var bufSizeExpr string
				var serCode, deserCode []string

				if bm != nil {
					serCode = []string{
						"{",
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", fld.Name, reg.Name, fld.Name, bm.Name, bm.StartBit()),
						fmt.Sprintf("    if err := putSlice(buf[offset:], r.%s); err != nil {", f.Name),
						"        return offset, err",
						"    }",
						fmt.Sprintf("    offset += int(elems) * %d", elemSize),
						"}",
					}
					deserCode = []string{
						"{",
						fmt.Sprintf("    elems := (r.%s&%s_%s_%s_bm)>>%d", fld.Name, reg.Name, fld.Name, bm.Name, bm.StartBit()),
						fmt.Sprintf("    r.%s = make([]%s, int(elems))", f.Name, elem),
						fmt.Sprintf("    if err := getSlice(buf[offset:], r.%s); err != nil {", f.Name),
						"        return offset, err",
						"    }",
						fmt.Sprintf("    offset += int(elems) * %d", elemSize),
						"}",
					}
					// Variable array buffer size: element size * bitfield value
					bufSizeExpr = fmt.Sprintf("(int((r.%s&%s_%s_%s_bm)>>%d) * %d)",
						fld.Name, reg.Name, fld.Name, bm.Name, bm.StartBit(), elemSize)
				} else {
					serCode = []string{
						"{",
						fmt.Sprintf("    elems := r.%s", refField),
						fmt.Sprintf("    if err := putSlice(buf[offset:], r.%s); err != nil {", f.Name),
						"        return offset, err",
						"    }",
						fmt.Sprintf("    offset += int(elems) * %d", elemSize),
						"}",
					}
					deserCode = []string{
						"{",
						fmt.Sprintf("    elems := r.%s", refField),
						fmt.Sprintf("    r.%s = make([]%s, int(elems))", f.Name, elem),
						fmt.Sprintf("    if err := getSlice(buf[offset:], r.%s); err != nil {", f.Name),
						"        return offset, err",
						"    }",
						fmt.Sprintf("    offset += int(elems) * %d", elemSize),
						"}",
					}
					// Variable array buffer size: element size * reference field
					bufSizeExpr = fmt.Sprintf("(int(r.%s) * %d)", refField, elemSize)
				}

				if gf.IsReadable {
					gf.SerializeReadData = append(gf.SerializeReadData, serCode...)
					gf.DeserializeReadData = append(gf.DeserializeReadData, deserCode...)
				}
				if gf.IsWritable {
					gf.SerializeWriteData = append(gf.SerializeWriteData, serCode...)
					gf.DeserializeWriteData = append(gf.DeserializeWriteData, deserCode...)
				}

				// Generate consistency checks for variable-length arrays
				if bm != nil {
					gf.ConsistencyChecks = append(gf.ConsistencyChecks,
						fmt.Sprintf("if len(r.%s) != int((r.%s&%s_%s_%s_bm)>>%d) {",
							f.Name, fld.Name, reg.Name, fld.Name, bm.Name, bm.StartBit()),
						fmt.Sprintf("    return fmt.Errorf(\"array %s length (%%d) does not match field %s value (%%d)\", len(r.%s), int((r.%s&%s_%s_%s_bm)>>%d))",
							f.Name, refField, f.Name, fld.Name, reg.Name, fld.Name, bm.Name, bm.StartBit()),
						"}")
				} else {
					gf.ConsistencyChecks = append(gf.ConsistencyChecks,
						fmt.Sprintf("if len(r.%s) != int(r.%s) {", f.Name, refField),
						fmt.Sprintf("    return fmt.Errorf(\"array %s length (%%d) does not match field %s value (%%d)\", len(r.%s), int(r.%s))",
							f.Name, refField, f.Name, refField),
						"}")
				}

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
				serCode := []string{
					fmt.Sprintf("if err := putNumber(buf[offset:], r.%s); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %d", size),
				}
				deserCode := []string{
					fmt.Sprintf("if err := getNumber(buf[offset:], &r.%s); err != nil {", f.Name),
					"    return offset, err",
					"}",
					fmt.Sprintf("offset += %d", size),
				}

				// Simple type buffer size is constant - add directly to register
				if gf.IsReadable {
					gf.SerializeReadData = append(gf.SerializeReadData, serCode...)
					gf.DeserializeReadData = append(gf.DeserializeReadData, deserCode...)
					gr.BufSize4ReadConst += size
				}
				if gf.IsWritable {
					gf.SerializeWriteData = append(gf.SerializeWriteData, serCode...)
					gf.DeserializeWriteData = append(gf.DeserializeWriteData, deserCode...)
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

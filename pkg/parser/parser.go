package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/dspasibenko/pargus/pkg/golibs/cast"
)

//
// AST
//

// A comment may take several lines in a row, separated by empty lines
type CommentGroup struct {
	Elements []*CommentElement `(@@)*`
}

type CommentElement struct {
	Comment   *string `@Comment`
	EmptyLine *string `| @EmptyLine`
}

type Device struct {
	Pos       lexer.Position
	Doc       *CommentGroup `@@?`
	Name      string        `"device" @Ident`
	Registers []*Register   `@@*`
}

type Register struct {
	Pos       lexer.Position
	Doc       *CommentGroup `@@?`
	Name      string        `"register" @Ident`
	NumberStr string        `"(" @Int ")"`
	Specifier string        `( ":" @("r"|"w") )?`
	Body      *RegisterBody `@@`
}

type RegisterBody struct {
	Items []*BodyItem `"{" ( @@ )* "}" ";"`
}

type BodyItem struct {
	Constant *Constant `@@`
	Field    *Field    `| @@`
}

type Constant struct {
	Pos      lexer.Position
	Doc      *CommentGroup `@@?`
	Name     string        `"const" @Ident "="`
	Type     SimpleType    `@@`
	ValueStr string        `"(" @Int ")" ";"`
}

type Field struct {
	Pos             lexer.Position
	Doc             *CommentGroup `@@?` // leading comments
	Name            string        `@Ident`
	Specifier       string        `( ":" @("r"|"w") )?`
	Type            *TypeUnion    `@@`
	TrailingComment *string       `@End`
}

//
// Type system
//

type Type interface {
	isType()
}

type SimpleType struct {
	Name string `@("int8"|"uint8"|"int16"|"uint16"|"int32"|"uint32"|"int64"|"uint64"|"float32"|"float64")`
}

type ArrayType struct {
	Size ArraySize  `"[" @@ "]"`
	Type SimpleType `@@`
}

type ArraySize struct {
	Constant *string `@Int`
	Variable *string `| @Ident`
}

type BitField struct {
	Base string      `@("uint8"|"uint16"|"uint32"|"uint64")`
	Bits []BitMember `"{" @@ ("," @@)* "}"`
}

type BitMember struct {
	Doc   *CommentGroup `@@?`
	Name  string        `@Ident ":"`
	Start string        `@Int`
	End   *string       `( "-" @Int )?`
}

//
// Union wrapper for interface
//

type TypeUnion struct {
	Bitfield *BitField   `  @@`
	Array    *ArrayType  `| @@`
	Simple   *SimpleType `| @@`
}

func (*SimpleType) isType() {}
func (*ArrayType) isType()  {}
func (*BitField) isType()   {}
func (*TypeUnion) isType()  {} // for compatibility

//
// Parser
//

var parser = participle.MustBuild[Device](
	participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
		{"End", `;([ \t]+//[^\r\n]*)?`},
		{"Comment", `//[^\r\n]*`},
		{"EmptyLine", `\n\s*\n`},
		{"Keyword", `\b(const|device|register)\b`},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_-]*`},
		{"Int", `0[xX][0-9a-fA-F]+|0[bB][01]+|\d+`},
		{"Punct", `[{}();:,\[\]=\-]`},
		{"Whitespace", `\s+`},
	})),
	participle.Elide("Whitespace"),
	participle.Union[Type](&SimpleType{}, &ArrayType{}, &BitField{}),
	participle.UseLookahead(4),
)

func Parse(input string) (*Device, error) {
	// trim the input
	input = trimString(input)

	device, err := parser.ParseString("", input)
	if err != nil {
		return nil, err
	}

	// Process trailing comments - extract comment part from TrailingComment tokens
	for _, register := range device.Registers {
		for _, field := range register.Body.Fields() {
			if field.TrailingComment == nil {
				continue
			}
			commentStart := strings.Index(*field.TrailingComment, "//")
			if commentStart == -1 {
				field.TrailingComment = cast.StringPtr("")
				continue
			}
			extractedComment := (*field.TrailingComment)[commentStart:]
			field.TrailingComment = cast.StringPtr(extractedComment)
		}
	}

	// Validate register numbers are unique
	registerNumbers := make(map[int64]bool)
	for _, r := range device.Registers {
		val := r.Number()
		if registerNumbers[val] {
			return nil, fmt.Errorf("duplicate register number %d", val)
		}
		registerNumbers[val] = true

		// Validate field specifiers compatibility with register specifier
		if err := r.validateAndUpdateFieldSpecifiers(); err != nil {
			return nil, err
		}

		// Validate bit fields
		if err := r.validateBitFields(); err != nil {
			return nil, err
		}

		// Validate arrays
		if err := r.validateArrays(); err != nil {
			return nil, err
		}
	}

	return device, nil
}

func trimString(input string) string {
	// Split into lines
	lines := strings.Split(input, "\n")

	firstNonEmpty := -1
	// Find the last non-empty line
	lastNonEmpty := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lastNonEmpty = i
			if firstNonEmpty == -1 {
				firstNonEmpty = i
			}
		}
	}

	// If no non-empty lines found, return empty string
	if lastNonEmpty == -1 {
		return ""
	}

	// Return lines up to and including the last non-empty line
	return strings.Join(lines[:lastNonEmpty+1], "\n")
}

func (rb *RegisterBody) Constants() []*Constant {
	var constants []*Constant
	for _, item := range rb.Items {
		if item.Constant != nil {
			constants = append(constants, item.Constant)
		}
	}
	return constants
}

func (rb *RegisterBody) Fields() []*Field {
	var fields []*Field
	for _, item := range rb.Items {
		if item.Field != nil {
			fields = append(fields, item.Field)
		}
	}
	return fields
}

func (r *Register) Number() int64 {
	val, err := strconv.ParseInt(r.NumberStr, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid register number %s", r.NumberStr))
	}
	return val
}

// validateAndUpdateFieldSpecifiers checks if field specifiers are compatible with register specifier
func (r *Register) validateAndUpdateFieldSpecifiers() error {
	registerSpec := r.Specifier
	// If register has no specifier, field can have any specifier or none
	if registerSpec == "" {
		return nil
	}

	for _, field := range r.Body.Fields() {
		fieldSpec := field.Specifier

		// If field has no specifier, it inherits register specifier (valid)
		if fieldSpec == "" {
			field.Specifier = registerSpec // inherit register specifier
			continue
		}

		// Check compatibility
		if registerSpec == "r" && fieldSpec == "w" {
			return fmt.Errorf("field '%s' in register '%s' cannot be write-only because register is read-only", field.Name, r.Name)
		}

		if registerSpec == "w" && fieldSpec == "r" {
			return fmt.Errorf("field '%s' in register '%s' cannot be read-only because register is write-only", field.Name, r.Name)
		}
	}

	return nil
}

// validateBitFields validates that bit fields use only unsigned integer types
// and that bit ranges don't exceed the size of the base type
func (r *Register) validateBitFields() error {
	for _, field := range r.Body.Fields() {
		if field.Type.Bitfield != nil {
			bitField := field.Type.Bitfield

			// Check that base type is unsigned
			if !isUnsignedType(bitField.Base) {
				return fmt.Errorf("bit field '%s' in register '%s' must use unsigned integer type, got '%s'",
					field.Name, r.Name, bitField.Base)
			}

			// Get the size of the base type in bits
			baseTypeBits := getTypeSizeInBits(bitField.Base)

			// Validate each bit member
			for _, bitMember := range bitField.Bits {
				endBit := bitMember.EndBit()

				// Check that bit range doesn't exceed base type size
				if endBit >= baseTypeBits {
					return fmt.Errorf("bit field '%s' in register '%s': bit range %s-%d exceeds size of base type '%s' (%d bits)",
						field.Name, r.Name, bitMember.Start, &endBit, bitField.Base, baseTypeBits)
				}

				// Check that start bit is not negative
				if bitMember.StartBit() < 0 {
					return fmt.Errorf("bit field '%s' in register '%s': bit position cannot be negative, got %s",
						field.Name, r.Name, bitMember.Start)
				}

				// Check that start <= end
				if bitMember.StartBit() > endBit {
					return fmt.Errorf("bit field '%s' in register '%s': start bit %s cannot be greater than end bit %d",
						field.Name, r.Name, bitMember.Start, endBit)
				}
			}
		}
	}

	return nil
}

func (bm *BitMember) EndBit() int {
	if bm.End != nil {
		val, err := strconv.ParseInt(*bm.End, 0, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid bit member end %s", *bm.End))
		}
		return int(val)
	}
	return bm.StartBit()
}

func (bm *BitMember) StartBit() int {
	val, err := strconv.ParseInt(bm.Start, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid bit member start %s", bm.Start))
	}
	return int(val)
}

func (c *Constant) Value() int64 {
	val, err := strconv.ParseInt(c.ValueStr, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid constant value %s", c.ValueStr))
	}
	return val
}

// isUnsignedType checks if a type is an unsigned integer type
func isUnsignedType(typeName string) bool {
	switch typeName {
	case "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}

// getTypeSizeInBits returns the size of a type in bits
func getTypeSizeInBits(typeName string) int {
	switch typeName {
	case "uint8":
		return 8
	case "uint16":
		return 16
	case "uint32":
		return 32
	case "uint64":
		return 64
	default:
		return 0
	}
}

// FindFieldByName finds a field by name in the register, checking both regular fields and bitfield members
func (r *Register) FindFieldByName(fieldName string, currentFieldIndex int) (*Field, *BitMember) {
	// Check regular fields declared before current field
	fields := r.Body.Fields()
	for i := 0; i < currentFieldIndex; i++ {
		field := fields[i]

		// Check if it's a regular field with matching name
		if field.Name == fieldName && field.Type.Simple != nil {
			return field, nil
		}

		// Check if it's a bitfield with matching member names
		if field.Type.Bitfield != nil {
			for _, bitMember := range field.Type.Bitfield.Bits {
				// Check if it's a bitfield reference (fieldName_bitMemberName)
				bitFieldRefName := field.Name + "_" + bitMember.Name
				if fieldName == bitFieldRefName {
					return field, &bitMember
				}
			}
		}
	}

	return nil, nil
}

// validateArrays validates that variable-length arrays use unsigned integer types for size
// and that referenced fields are declared before the array
func (r *Register) validateArrays() error {
	fields := r.Body.Fields()
	for i, field := range fields {
		if field.Type.Array == nil {
			continue
		}
		arrayType := field.Type.Array

		if arrayType.Size.Variable == nil {
			// this is a constant-length array
			continue
		}

		// Check if it's a variable-length array with field reference
		fieldName := cast.String(arrayType.Size.Variable, "")

		// This is a field reference - check if the referenced field exists and is declared before this array
		exists, _ := r.FindFieldByName(fieldName, i)
		if exists == nil {
			return fmt.Errorf("variable-length array '%s' in register '%s' references undefined field '%s'",
				field.Name, r.Name, fieldName)
		}
	}
	return nil
}

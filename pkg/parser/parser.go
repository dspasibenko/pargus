package parser

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
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
	Number    int           `"(" @Int ")"`
	Specifier *string       `( ":" @("r"|"w") )?`
	Fields    []*Field      `"{" ( @@ )* "}" ";"`
}

type Field struct {
	Pos             lexer.Position
	Doc             *CommentGroup `@@?` // leading comments
	Name            string        `@Ident`
	Specifier       *string       `( ":" @("r"|"w") )?`
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
	Size    *ArraySize `"[" @@ "]"`
	Element *TypeUnion `@@`
}

type ArraySize struct {
	Constant *int    `@Int`
	Type     *string `| @("uint8"|"uint16"|"uint32"|"uint64")`
	Variable *string `| @Ident`
}

type BitField struct {
	Base string       `@("int8"|"uint8"|"int16"|"uint16"|"int32"|"uint32"|"int64"|"uint64")`
	Bits []*BitMember `"{" @@ ("," @@)* "}"`
}

type BitMember struct {
	Doc   *CommentGroup `@@?`
	Name  string        `@Ident ":"`
	Start int           `@Int`
	End   *int          `( "-" @Int )?`
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
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_-]*`},
		{"Int", `\d+`},
		{"Punct", `[{}();:,\[\]-]`},
		{"Whitespace", `\s+`},
	})),
	participle.Elide("Whitespace"),
)

func Parse(input string) (*Device, error) {
	// Remove trailing empty lines from input before parsing
	input = removeTrailingEmptyLinesFromString(input)

	device, err := parser.ParseString("", input)
	if err != nil {
		return nil, err
	}

	// Process trailing comments - extract comment part from TrailingComment tokens
	for _, register := range device.Registers {
		for _, field := range register.Fields {
			if field.TrailingComment != nil && *field.TrailingComment != ";" && strings.Contains(*field.TrailingComment, "//") {
				// Extract comment part from semicolon + comment
				commentStart := strings.Index(*field.TrailingComment, "//")
				if commentStart != -1 {
					extractedComment := (*field.TrailingComment)[commentStart:]
					field.TrailingComment = &extractedComment
				}
			} else if field.TrailingComment != nil && *field.TrailingComment == ";" {
				// No trailing comment, set to empty string
				empty := ""
				field.TrailingComment = &empty
			}
		}
	}

	// Fix comment distribution - move trailing comments to leading comments of next field
	for _, register := range device.Registers {
		fixCommentDistribution(register)
	}

	// Remove trailing empty lines from device and registers
	removeTrailingEmptyLines(device)

	// Validate register numbers are unique
	registerNumbers := make(map[int]bool)
	for _, register := range device.Registers {
		if registerNumbers[register.Number] {
			return nil, fmt.Errorf("duplicate register number %d", register.Number)
		}
		registerNumbers[register.Number] = true

		// Validate field specifiers compatibility with register specifier
		if err := validateAndUpdateFieldSpecifiers(register); err != nil {
			return nil, err
		}

		// Validate bit fields
		if err := validateBitFields(register); err != nil {
			return nil, err
		}

		// Validate arrays
		if err := validateArrays(register); err != nil {
			return nil, err
		}
	}

	return device, nil
}

// fixCommentDistribution - trailing comments should stay with their fields
func fixCommentDistribution(register *Register) {
	// Do nothing - trailing comments should remain with their fields
	// This function is kept for compatibility but doesn't move comments
}

func removeTrailingEmptyLines(device *Device) {
	// Remove trailing empty lines from device Doc
	if device.Doc != nil {
		device.Doc.Elements = removeTrailingEmptyLinesFromElements(device.Doc.Elements)
	}

	// Remove trailing empty lines from each register
	for _, register := range device.Registers {
		// Remove trailing empty lines from register Doc
		if register.Doc != nil {
			register.Doc.Elements = removeTrailingEmptyLinesFromElements(register.Doc.Elements)
		}

		// Remove trailing empty lines from each field
		for _, field := range register.Fields {
			if field.Doc != nil {
				field.Doc.Elements = removeTrailingEmptyLinesFromElements(field.Doc.Elements)
			}
		}
	}
}

func removeTrailingEmptyLinesFromString(input string) string {
	// Split into lines
	lines := strings.Split(input, "\n")

	// Find the last non-empty line
	lastNonEmpty := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lastNonEmpty = i
		}
	}

	// If no non-empty lines found, return empty string
	if lastNonEmpty == -1 {
		return ""
	}

	// Return lines up to and including the last non-empty line
	return strings.Join(lines[:lastNonEmpty+1], "\n")
}

func removeTrailingEmptyLinesFromElements(elements []*CommentElement) []*CommentElement {
	// Find the last non-empty line element
	lastNonEmpty := -1
	for i, element := range elements {
		if element.Comment != nil {
			lastNonEmpty = i
		}
	}

	// If no comments found, return empty slice
	if lastNonEmpty == -1 {
		return []*CommentElement{}
	}

	// Return elements up to and including the last non-empty line
	return elements[:lastNonEmpty+1]
}

// validateAndUpdateFieldSpecifiers checks if field specifiers are compatible with register specifier
func validateAndUpdateFieldSpecifiers(register *Register) error {
	registerSpec := register.Specifier
	// If register has no specifier, field can have any specifier or none
	if registerSpec == nil {
		return nil
	}

	for _, field := range register.Fields {
		fieldSpec := field.Specifier

		// If field has no specifier, it inherits register specifier (valid)
		if fieldSpec == nil {
			field.Specifier = registerSpec // inherit register specifier
			continue
		}

		// Check compatibility
		if *registerSpec == "r" && *fieldSpec == "w" {
			return fmt.Errorf("field '%s' in register '%s' cannot be write-only because register is read-only", field.Name, register.Name)
		}

		if *registerSpec == "w" && *fieldSpec == "r" {
			return fmt.Errorf("field '%s' in register '%s' cannot be read-only because register is write-only", field.Name, register.Name)
		}
	}

	return nil
}

// validateBitFields validates that bit fields use only unsigned integer types
// and that bit ranges don't exceed the size of the base type
func validateBitFields(register *Register) error {
	for _, field := range register.Fields {
		if field.Type.Bitfield != nil {
			bitField := field.Type.Bitfield

			// Check that base type is unsigned
			if !isUnsignedType(bitField.Base) {
				return fmt.Errorf("bit field '%s' in register '%s' must use unsigned integer type, got '%s'",
					field.Name, register.Name, bitField.Base)
			}

			// Get the size of the base type in bits
			baseTypeBits := getTypeSizeInBits(bitField.Base)

			// Validate each bit member
			for _, bitMember := range bitField.Bits {
				var endBit int
				if bitMember.End != nil {
					endBit = *bitMember.End
				} else {
					endBit = bitMember.Start
				}

				// Check that bit range doesn't exceed base type size
				if endBit >= baseTypeBits {
					return fmt.Errorf("bit field '%s' in register '%s': bit range %d-%d exceeds size of base type '%s' (%d bits)",
						field.Name, register.Name, bitMember.Start, endBit, bitField.Base, baseTypeBits)
				}

				// Check that start bit is not negative
				if bitMember.Start < 0 {
					return fmt.Errorf("bit field '%s' in register '%s': bit position cannot be negative, got %d",
						field.Name, register.Name, bitMember.Start)
				}

				// Check that start <= end
				if bitMember.Start > endBit {
					return fmt.Errorf("bit field '%s' in register '%s': start bit %d cannot be greater than end bit %d",
						field.Name, register.Name, bitMember.Start, endBit)
				}
			}
		}
	}

	return nil
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

// findFieldByName finds a field by name in the register, checking both regular fields and bitfield members
func findFieldByName(register *Register, fieldName string, currentFieldIndex int) (bool, string) {
	// Check regular fields declared before current field
	for i := 0; i < currentFieldIndex; i++ {
		field := register.Fields[i]

		// Check if it's a regular field with matching name
		if field.Name == fieldName {
			return true, "field"
		}

		// Check if it's a bitfield with matching member names
		if field.Type.Bitfield != nil {
			for _, bitMember := range field.Type.Bitfield.Bits {
				if bitMember.Name == fieldName {
					return true, "bitfield"
				}
				// Check if it's a bitfield reference (fieldName_bitMemberName)
				bitFieldRefName := field.Name + "_" + bitMember.Name
				if fieldName == bitFieldRefName {
					return true, "bitfield_ref"
				}
				// Check if it's a bitmask reference (fieldName_bitMemberName_bm)
				bitmaskName := field.Name + "_" + bitMember.Name + "_bm"
				if fieldName == bitmaskName {
					return true, "bitmask"
				}
			}
		}
	}

	return false, ""
}

// validateArrays validates that variable-length arrays use unsigned integer types for size
// and that referenced fields are declared before the array
func validateArrays(register *Register) error {
	for i, field := range register.Fields {
		if field.Type.Array != nil {
			arrayType := field.Type.Array

			// Check if it's a variable-length array with type size
			if arrayType.Size.Type != nil {
				sizeType := *arrayType.Size.Type

				// Check that size type is unsigned integer
				if !isUnsignedType(sizeType) {
					return fmt.Errorf("variable-length array '%s' in register '%s' must use unsigned integer type for size, got '%s'",
						field.Name, register.Name, sizeType)
				}
			}

			// Check if it's a variable-length array with field reference
			if arrayType.Size.Variable != nil {
				fieldName := *arrayType.Size.Variable

				// Check if this is actually a type name (uint8, uint16, etc.) or a field reference
				if isUnsignedType(fieldName) {
					// This is a type name, not a field reference - this should be handled by Type field
					// This case should not happen if the parser is working correctly
					return fmt.Errorf("variable-length array '%s' in register '%s' has type '%s' in Variable field instead of Type field",
						field.Name, register.Name, fieldName)
				}

				// This is a field reference - check if the referenced field exists and is declared before this array
				exists, _ := findFieldByName(register, fieldName, i)
				if !exists {
					return fmt.Errorf("variable-length array '%s' in register '%s' references undefined field '%s'",
						field.Name, register.Name, fieldName)
				}

				// The field exists and is declared before the array, which is valid
				// Note: We don't need to check the type here as the field could be a bitfield
				// and the actual size will be determined at runtime
			}
		}
	}

	return nil
}

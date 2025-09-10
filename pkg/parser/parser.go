package parser

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Device struct {
	Pos       lexer.Position
	Comment   *string     `@Comment?`
	Name      string      `"device" @Ident`
	Registers []*Register `@@*`
}

type Register struct {
	Pos       lexer.Position
	Comment   *string  `@Comment?`
	Name      string   `"register" @Ident`
	Number    int      `"(" @Int ")"`
	Specifier *string  `( ":" @("r"|"w") )?`
	Fields    []*Field `"{" ( @@ ";" )* "}" ";"`
}

type Field struct {
	Pos        lexer.Position
	Comment    *string    `@Comment?`
	Name       string     `@Ident`
	Specifier  *string    `( ":" @("r"|"w") )?`
	Type       *TypeUnion `@@`
	EndComment *string    `@Comment?`
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
	Size    string     `"[" (@Int | @Ident) "]"`
	Element *TypeUnion `@@`
}

type BitField struct {
	Base string       `@("uint8"|"uint16"|"uint32"|"uint64")`
	Bits []*BitMember `"{" @@ ("," @@)* "}"`
}

type BitMember struct {
	Comment *string `@Comment?`
	Name    string  `@Ident ":"`
	Start   int     `@Int`
	End     *int    `( "-" @Int )?`
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

var parser = participle.MustBuild[Device](
	participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(//[^\r\n]*(\r?\n\s*)*)*//[^\r\n]*`},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"Int", `\d+`},
		{"Punct", `[{}();:,\[\]-]`},
		{"Whitespace", `\s+`},
	})),
	participle.Elide("Whitespace"),
)

func Parse(input string) (*Device, error) {
	device, err := parser.ParseString("", input)
	if err != nil {
		return nil, err
	}

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
	}

	return device, nil
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

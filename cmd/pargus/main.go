package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dspasibenko/pargus/pkg/generator"
	"github.com/dspasibenko/pargus/pkg/parser"
)

func main() {
	var (
		outputFile     = flag.String("output", "", "Output file (default: input.h for C++, input.go for Go)")
		outputShort    = flag.String("o", "", "Output file (short form)")
		namespace      = flag.String("namespace", "", "C++ namespace name (required for C++)")
		namespaceShort = flag.String("n", "", "C++ namespace name (short form, required for C++)")
		packageName    = flag.String("package", "", "Go package name (required for Go)")
		packageShort   = flag.String("p", "", "Go package name (short form, required for Go)")
		generatorType  = flag.String("type", "cpp", "Generator type: cpp or go")
		generatorShort = flag.String("t", "cpp", "Generator type: cpp or go (short form)")
		help           = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.pa\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate C++ code:\n")
		fmt.Fprintf(os.Stderr, "  %s -t cpp -n MyNamespace -o output.h input.pa\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate Go code:\n")
		fmt.Fprintf(os.Stderr, "  %s -t go -p mypackage -o output.go input.pa\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Get generator type
	genType := *generatorType
	if *generatorShort != "" {
		genType = *generatorShort
	}
	if genType != "cpp" && genType != "go" {
		fmt.Fprintf(os.Stderr, "Error: generator type must be 'cpp' or 'go'\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check for required parameters based on generator type
	var ns, pkg string
	if genType == "cpp" {
		ns = *namespace
		if ns == "" {
			ns = *namespaceShort
		}
		if ns == "" {
			fmt.Fprintf(os.Stderr, "Error: namespace parameter is required for C++ generator\n")
			flag.Usage()
			os.Exit(1)
		}
	} else {
		pkg = *packageName
		if pkg == "" {
			pkg = *packageShort
		}
		if pkg == "" {
			fmt.Fprintf(os.Stderr, "Error: package parameter is required for Go generator\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	// Check for output file parameter
	output := *outputFile
	if output == "" {
		output = *outputShort
	}

	// Get input file from command line arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one input file is allowed\n")
		flag.Usage()
		os.Exit(1)
	}

	inputFile := args[0]

	// Set default output file if not specified
	if output == "" {
		ext := filepath.Ext(inputFile)
		base := filepath.Base(inputFile)
		if ext != "" {
			base = base[:len(base)-len(ext)]
		}
		if genType == "cpp" {
			output = base + ".h"
		} else {
			output = base + ".go"
		}
	}

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	// Parse the input
	device, err := parser.Parse(string(inputData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
		os.Exit(1)
	}

	// Generate code
	var code string
	if genType == "cpp" {
		code, err = generator.GenerateCpp(device, ns, strings.ReplaceAll(output, ".", "_"))
	} else {
		code, err = generator.GenerateGo(device, pkg)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	err = os.WriteFile(output, []byte(code), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", output, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s\n", output)
}

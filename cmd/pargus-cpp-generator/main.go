package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dspasibenko/pargus/pkg/generator"
	"github.com/dspasibenko/pargus/pkg/parser"
)

func main() {
	var (
		outputFile = flag.String("output", "", "Output header file (default: input.h)")
		outputShort = flag.String("o", "", "Output header file (short form)")
		namespace = flag.String("namespace", "", "C++ namespace name (required)")
		namespaceShort = flag.String("n", "", "C++ namespace name (short form, required)")
		help = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.pa\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -n MyNamespace -o output.h input.pa\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Check for required namespace parameter
	ns := *namespace
	if ns == "" {
		ns = *namespaceShort
	}
	if ns == "" {
		fmt.Fprintf(os.Stderr, "Error: namespace parameter is required\n")
		flag.Usage()
		os.Exit(1)
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
		output = base + ".h"
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

	// Create generator
	gen, err := generator.NewArduinoCppGenerator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	// Generate code
	code, err := gen.Generate(device, ns)
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

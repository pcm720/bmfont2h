package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pcm720/bmfont2h/bmfont"
)

// Parses BMFont text font descriptor and emits C header file containing font metadata and page files converted to C char array

func main() {
	if len(os.Args) < 2 {
		fmt.Println("bmfont2h — a BMFont to C header converter\nUsage:\n\tbmfont2h <input file> <output file>")
		return
	}
	var err error

	// Emit C header to output file
	outPath := ""
	if len(os.Args) == 3 {
		outPath, err = filepath.Abs(os.Args[2])
		if err != nil {
			log.Fatalf("failed to parse output file path: %s", err)
		}
	}

	// Get absolute filepath and chdir to the font folder
	inPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalf("can't get absolute path for the input file: %s", err)
	}

	// Parse the descriptor
	parsedFont, err := bmfont.ParseDescriptor(inPath)
	if err != nil {
		log.Fatalf("failed to parse font descriptor: %s", err)
	}

	if info, err := os.Stat(outPath); (outPath == "") || (err == nil && info.IsDir()) {
		outPath = filepath.Join(outPath, strings.ToLower(parsedFont.Name)+".h")
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to open output file: %s", err)
	}
	parsedFont.EmitCFont(outFile)
	outFile.Close()
}

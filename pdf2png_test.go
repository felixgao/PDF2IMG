package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"testing"
)

func TestConvertPDFToPNG(t *testing.T) {
	// Read the PDF file
	pdfData, err := ioutil.ReadFile("/Users/ggao/Downloads/ATT00001.pdf")
	if err != nil {
		t.Fatalf("failed to read PDF file: %v", err)
	}

	// Set up the conversion options
	options := ConvertOptions{
		PDFFile:     pdfData,
		Resolution:  150,
		PageIndices: []int{1, 2},
	}

	// Convert PDF to PNG
	compressedData, err := convertPDFToPNG(options)
	if err != nil {
		t.Fatalf("failed to convert PDF to PNG: %v", err)
	}

	// Check if the output is not empty
	if len(compressedData) == 0 {
		t.Fatal("empty compressed output")
	}
	println(len(compressedData))
	ioutil.WriteFile("/Users/ggao/Downloads/ATT00001.zip", compressedData, 0777)

	// Open the compressed data as a zip archive
	zipReader, err := zip.NewReader(bytes.NewReader(compressedData), int64(len(compressedData)))
	if err != nil {
		t.Fatalf("failed to open compressed file: %v", err)
	}

	// Validate the number of PNG files in the archive
	expectedNumFiles := len(options.PageIndices) // Change it to the expected number of PNG files
	if len(zipReader.File) != expectedNumFiles {
		t.Fatalf("unexpected number of PNG files in the archive: got %d, want %d", len(zipReader.File), expectedNumFiles)
	}

	// Extract and validate each PNG file (optional)
	for _, file := range zipReader.File {
		// Extract the PNG file
		pngFile, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open PNG file: %v", err)
		}

		// Read the PNG file data
		pngData, err := ioutil.ReadAll(pngFile)
		if err != nil {
			t.Fatalf("failed to read PNG file: %v", err)
		}

		if len(pngData) == 0 {
			t.Fatal("empty PNG file")
		}

		// Close the PNG file
		pngFile.Close()
	}
}

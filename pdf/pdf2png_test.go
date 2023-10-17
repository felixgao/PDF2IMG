package pdf

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func validateResult(compressedData []byte, expectedNumFiles int, exportOptions ExportOptions, tb testing.TB) {
	// Check if the output is not empty
	if len(compressedData) == 0 {
		tb.Fatal("empty compressed output")
	}

	// Open the compressed data as a zip archive
	zipReader, err := zip.NewReader(bytes.NewReader(compressedData), int64(len(compressedData)))
	if err != nil {
		tb.Fatalf("failed to open compressed file: %v", err)
	}

	filetype := strings.ToUpper(exportOptions.Format)

	// Validate the number of PNG files in the archive
	if len(zipReader.File) != expectedNumFiles {
		tb.Fatalf("unexpected number of %s files in the archive: got %d, want %d", filetype, len(zipReader.File), expectedNumFiles)
	}

	// Extract and validate each image file (optional)
	for _, file := range zipReader.File {
		// Extract the PNG file
		imgFile, err := file.Open()
		if err != nil {
			tb.Fatalf("failed to open %s file: %v", filetype, err)
		}

		// Read the image file data
		imgData, err := io.ReadAll(imgFile)
		if err != nil {
			tb.Fatalf("failed to read %s file: %v", filetype, err)
		}

		if len(imgData) == 0 {
			tb.Fatalf("empty Image file: %s", file.Name)
		}

		// Close the PNG file
		imgFile.Close()
	}
}

func TestConvertPDFToImage(t *testing.T) {

	// Read the PDF file
	pdfData, err := os.ReadFile("/Users/ggao/Downloads/ATT00001.pdf")
	if err != nil {
		t.Fatalf("failed to read PDF file: %v", err)
	}

	// Set up the conversion options
	options := ConvertOptions{
		PDFFile:     pdfData,
		PageIndices: []int{1, 2},
	}

	exportOptions := ExportOptions{
		Resolution: 300,
		Format:     "tiff",
		Quality:    80,
	}

	// Convert PDF to PNG
	compressedData, err := ConvertPDFToImage(options, exportOptions)
	if err != nil {
		t.Fatalf("failed to convert PDF to PNG: %v", err)
	}
	// Validate the result
	validateResult(compressedData, len(options.PageIndices), exportOptions, t)

	os.WriteFile("/Users/ggao/Downloads/ATT00001.zip", compressedData, 0777)

}

type FileInput struct {
	Name    string
	Indices []int
}

func setup() {
	vips.LoggingSettings(nil, vips.LogLevelInfo)
	vips.Startup(nil)
}

func BenchmarkGetPDFPageCount(b *testing.B) {
	inputTable := []FileInput{
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2, 3, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2, 3, 4}},
	}
	fmt.Println("Setup")
	setup()
	fmt.Println("Setup Done")
	fmt.Println("Start Benchmarking")

	for _, input := range inputTable {
		pdfData, err := os.ReadFile(input.Name)
		if err != nil {
			b.Fatalf("failed to read PDF file: %v", err)
		}

		b.RunParallel(func(pb *testing.PB) {
			b.ReportAllocs()
			b.ResetTimer()
			// Convert PDF to Image
			for pb.Next() {
				_, err := GetPDFPageCount(pdfData)
				if err != nil {
					b.Fatalf("failed to get PDF page count: %v", err)
				}
			}
		})
	}
}

func BenchmarkConvertPDFToImage(b *testing.B) {
	inputTable := []FileInput{
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{2}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{3}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 3}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{2, 3}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{2, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{3, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2, 3}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{2, 3, 4}},
		{Name: "/Users/ggao/Downloads/ATT00001.pdf", Indices: []int{1, 2, 3, 4}},
	}
	fmt.Println("Setup")
	setup()
	fmt.Println("Setup Done")
	fmt.Println("Start Benchmarking")

	for _, input := range inputTable {
		pdfData, err := os.ReadFile(input.Name)
		if err != nil {
			b.Fatalf("failed to read PDF file: %v", err)
		}

		// Set up the conversion options
		options := ConvertOptions{
			PDFFile:     pdfData,
			PageIndices: input.Indices,
		}
		exportOptions := ExportOptions{
			Resolution: 300,
			Format:     "tiff",
			Quality:    100,
		}

		b.RunParallel(func(pb *testing.PB) {
			b.ReportAllocs()
			b.ResetTimer()
			// Convert PDF to Image
			for pb.Next() {
				compressedData, _ := ConvertPDFToImage(options, exportOptions)
				validateResult(compressedData, len(options.PageIndices), exportOptions, b)
			}
		})
	}
}

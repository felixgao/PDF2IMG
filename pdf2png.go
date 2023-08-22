package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
)

type ConvertOptions struct {
	PDFFile     []byte
	PageIndices []int
}

type ExportOptions struct {
	Resolution int    `default:"300"`
	Format     string `default:"png"`
	Quality    int    `default:"100"`
}

type ImageResult struct {
	Image     []byte
	Index     int
	Extension string
}

var imageTypeMap = map[vips.ImageType]string{
	vips.ImageTypeJPEG: "jpg",
	vips.ImageTypePNG:  "png",
	vips.ImageTypeTIFF: "tiff",
}

var imageExtensionMap = map[string]vips.ImageType{
	"jpg":  vips.ImageTypeJPEG,
	"png":  vips.ImageTypePNG,
	"tiff": vips.ImageTypeTIFF,
}

func export(image *vips.ImageRef, exportOption ExportOptions) (string, []byte, *vips.ImageMetadata, error) {
	var format = imageExtensionMap[exportOption.Format]

	switch format {
	case vips.ImageTypePNG:
		ep := vips.NewPngExportParams()
		ext := imageTypeMap[format]
		imgBytes, imgMeta, err := image.ExportPng(ep)
		return ext, imgBytes, imgMeta, err
	case vips.ImageTypeTIFF:
		ep := vips.NewTiffExportParams()
		ext := imageTypeMap[format]
		if exportOption.Quality > 0 {
			ep.Compression = vips.TiffCompressionLzw
			ep.Quality = exportOption.Quality
		}
		imgBytes, imgMeta, err := image.ExportTiff(ep)
		return ext, imgBytes, imgMeta, err
	default:
		ext := imageTypeMap[vips.ImageTypeJPEG]
		ep := vips.NewJpegExportParams()
		if exportOption.Quality > 0 {
			ep.Quality = exportOption.Quality
		}
		imgBytes, imgMeta, err := image.ExportJpeg(ep)
		return ext, imgBytes, imgMeta, err
	}
}

func getPDFPageCount(pdfFile []byte) (int, error) {
	// Load the PDF file using vips
	pdfImportParams := vips.NewImportParams()
	pdfImportParams.Density.Set(35)
	pdfImportParams.Page.Set(0)
	pdfImportParams.NumPages.Set(1)
	tmp, err := vips.LoadImageFromBuffer(pdfFile, pdfImportParams)
	if err != nil {
		// unable to load PDF file
		return 0, fmt.Errorf("failed to load PDF file: %s", err.Error())
	}
	defer tmp.Close()
	return tmp.Pages(), nil
}

func convertPDFToImage(convertOptions ConvertOptions, exportOptions ExportOptions) ([]byte, error) {

	// Load the PDF file using vips
	// Get total pages of document

	pageCount, err := getPDFPageCount(convertOptions.PDFFile)
	if err != nil {
		// unable to load PDF file to get the page count
		return nil, err
	}

	// Validate the page indices
	for _, pageIndex := range convertOptions.PageIndices {
		if pageIndex < 1 || pageIndex > pageCount {
			return nil, fmt.Errorf("invalid page index: %d", pageIndex)
		}
	}

	// Create a new zip buffer
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// Start a Pipeline of Goroutines to convert PDF pages to PNG images
	page_count := len(convertOptions.PageIndices)
	var wg sync.WaitGroup
	imageChan := make(chan *ImageResult, page_count)
	wg.Add(page_count)
	// Iterate over the specified page indices
	for _, pageIndex := range convertOptions.PageIndices {
		go func(pageIndex int, pdfFile []byte) {
			defer wg.Done()

			// Load the PDF file using vips with options
			pdfImportParams := vips.NewImportParams()
			pdfImportParams.Density.Set(exportOptions.Resolution)
			pdfImportParams.Page.Set(pageIndex)
			pdfImportParams.NumPages.Set(1)

			// Render the PDF page to an image, lock the cirtical section for vips library access
			pageImage, err := vips.LoadImageFromBuffer(convertOptions.PDFFile, pdfImportParams)
			if err != nil {
				fmt.Printf("failed to render PDF page: %s\n", err.Error())
				return
			}
			defer pageImage.Close()

			extension, imgBuf, _, err := export(pageImage, exportOptions)
			if err != nil {
				fmt.Printf("failed to convert image to %s format: %s\n", extension, err.Error())
				return
			}

			result := ImageResult{
				Image:     imgBuf,
				Index:     pageIndex,
				Extension: extension,
			}

			// Send the result to the channel
			imageChan <- &result
		}(pageIndex, convertOptions.PDFFile)
	}

	// start a Goroutine to wait for all workers to finish
	go func() {
		// Wait for all Goroutines to finish
		wg.Wait()
		close(imageChan)

	}()
	// End of Pipeline

	// Iterate over the received PNG images

	for result := range imageChan {
		// Access the page index and image from the ImageResult struct
		pageIndex := result.Index
		pageImage := result.Image
		pageExtension := result.Extension

		// Create a new PNG file in the zip archive
		fileName := fmt.Sprintf("/page_%d.%s", pageIndex, pageExtension)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			fmt.Printf("failed to create %s file in zip: %s\n", pageExtension, err.Error())
			continue
		}

		// Write the PNG image data to the zip file
		_, err = fileWriter.Write(pageImage)
		if err != nil {
			fmt.Printf("failed to write %s data to zip: %s\n", pageExtension, err.Error())
		}
	}

	err = zipWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush zip writer: %s", err.Error())
	}
	zipWriter.Close()

	// Return the zip file contents
	return zipBuffer.Bytes(), nil
}

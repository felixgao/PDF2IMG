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
	Resolution  int
}

type ImageResult struct {
	Image []byte
	Index int
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

func convertPDFToPNG(options ConvertOptions) ([]byte, error) {

	// Load the PDF file using vips
	// Get total pages of document

	pageCount, err := getPDFPageCount(options.PDFFile)
	if err != nil {
		// unable to load PDF file to get the page count
		return nil, err
	}

	println("pageCount: ", pageCount)

	// Validate the page indices
	for _, pageIndex := range options.PageIndices {
		if pageIndex < 1 || pageIndex > pageCount {
			return nil, fmt.Errorf("invalid page index: %d", pageIndex)
		}
	}

	// Create a new zip buffer
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// Start a Pipeline of Goroutines to convert PDF pages to PNG images
	var wg sync.WaitGroup
	imageChan := make(chan *ImageResult, len(options.PageIndices))
	wg.Add(len(options.PageIndices))
	// Iterate over the specified page indices
	for _, pageIndex := range options.PageIndices {
		go func(pageIndex int, pdfFile []byte, resolution int) {
			defer wg.Done()

			// Load the PDF file using vips with options
			pdfImportParams := vips.NewImportParams()
			pdfImportParams.Density.Set(options.Resolution)
			pdfImportParams.Page.Set(pageIndex)
			pdfImportParams.NumPages.Set(1)

			// Render the PDF page to an image, lock the cirtical section for vips library access
			pageImage, err := vips.LoadImageFromBuffer(options.PDFFile, pdfImportParams)
			if err != nil {
				fmt.Printf("failed to render PDF page: %s\n", err.Error())
				return
			}
			defer pageImage.Close()

			ep := vips.NewPngExportParams()
			pngBuf, _, err := pageImage.ExportPng(ep)
			if err != nil {
				fmt.Printf("failed to convert image to PNG format: %s\n", err.Error())
				return
			}

			result := ImageResult{
				Image: pngBuf,
				Index: pageIndex,
			}

			// Send the result to the channel
			imageChan <- &result
		}(pageIndex, options.PDFFile, options.Resolution)
	}

	// start a Goroutine to write the PNG images to the zip file
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

		// Create a new PNG file in the zip archive
		fileName := fmt.Sprintf("/page_%d.png", pageIndex)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			fmt.Printf("failed to create PNG file in zip: %s\n", err.Error())
			continue
		}

		// Write the PNG image data to the zip file
		_, err = fileWriter.Write(pageImage)
		if err != nil {
			fmt.Printf("failed to write PNG data to zip: %s\n", err.Error())
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

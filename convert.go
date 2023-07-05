package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func convertHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Set the maximum file size to 10MB
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the uploaded PDF file from the form
	file, _, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Failed to get PDF file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the PDF content into memory
	pdfContent, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read PDF content", http.StatusInternalServerError)
		return
	}

	// Parse the page indices parameter
	pageIndicesParam := r.FormValue("pages")
	pageIndices, err := parsePageIndices(pageIndicesParam, 50)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid page indices: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Set the default resolution to 300 dpi if not specified
	resolutionParam := r.FormValue("resolution")
	resolution, err := strconv.Atoi(resolutionParam)
	if err != nil || resolution <= 0 {
		resolution = 300
	}

	// Convert the specified pages to PNG and add them to the zip file

	data, err := convertPDFToPNG(ConvertOptions{
		PDFFile:     pdfContent,
		PageIndices: pageIndices,
		Resolution:  resolution,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert pages %v to PNG: %s", pageIndices, err.Error()), http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=converted_images.zip")

	// Write the zip file contents to the response
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, "Failed to write zip file contents", http.StatusInternalServerError)
		return
	}
}

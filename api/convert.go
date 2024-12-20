package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/felixgao/pdf_to_png/pdf"
	"github.com/felixgao/pdf_to_png/util"
)

// TODO: add the end points to a router group /api
func RegisterConvertHandlers(handler *gin.Engine) {
	handler.POST("/convert", convertHandler)
	handler.POST("/api/convert", convertHandler)
}

// TODO: this could be moved to a middleware file or file checker util file
// ex. mideelwares/middlewares.go
func pdfCheckerMiddleware(c *gin.Context) {
	// Multipart form
	form, _ := c.MultipartForm()
	files := form.File["file[]"]

	if len(files) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No PDF file found",
		})
		return
	}

	if len(files) > 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Too many PDF files found, only one is allowed",
		})
		return
	}

	// check if content type is set correctly
	// not sure getting the file like this will resulted in reading the file.
	file := files[0]
	if file.Header.Get("Content-Type") != "application/pdf" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Content-Type is not a PDF",
		})
		return
	}

	c.Next()
}

// TODO might need to figure out how to turn this into a middleware
// hint: ioutil.NopCloser and ioutil.ReadAll(c.Request.Body)
func DetectContentType(c *gin.Context, f io.ReadSeeker) string {
	fileHeader := make([]byte, 512)
	if _, err := f.Read(fileHeader); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Error appears when operating file: " + err.Error(),
		})
		return ""
	}

	filetype := http.DetectContentType(fileHeader)

	if _, err := f.Seek(0, 0); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Error appears when operating file: " + err.Error(),
		})
		return ""
	}

	return filetype
}

// This is a function that helps parsing the request body multiple times.
// Warning: this will create a new buffer and will create addtional memory usage.
func getRequestBody(c *gin.Context) []byte {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var rawBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		bodyBytes = []byte("{}")
	}

	return bodyBytes
}

// @Summary Converting PDF to PNG
// @Tags Convert
// @Produce application/octet-stream
// @Success 200
// @Router /convert [post]
func convertHandler(c *gin.Context) {
	// Setup tracing and metrics
	var tracer = otel.Tracer("pdf2img")
	var meter = otel.Meter("pdf2img")
	ctx, childSpan := tracer.Start(c.Request.Context(), "parameter-check-span")
	duration, _ := meter.Int64Histogram("request_duration")
	counter, _ := meter.Int64Counter("request_count")
	startTime := time.Now()
	opts := []attribute.KeyValue{}

	// Multipart form
	form, _ := c.MultipartForm()
	files := form.File["file[]"]
	pdf_file := files[0]

	// Get the uploaded PDF file from the form
	f, openErr := pdf_file.Open()
	if openErr != nil {
		opts = append(opts, attribute.Key("ConvertError").String("PDF Open Error"))
		counter.Add(ctx, 1, metric.WithAttributes(opts...))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to open PDF file from form",
		})
		return
	}
	defer f.Close()
	file_type := DetectContentType(c, f)
	if file_type != "application/pdf" {
		opts = append(opts, attribute.Key("ConvertError").String("Wrong Content Type"))
		counter.Add(ctx, 1, metric.WithAttributes(opts...))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Content-Type is not a application/pdf",
		})
		return
	}
	// Read the PDF content into memory
	pdfContent, err := io.ReadAll(f)
	if err != nil {
		opts = append(opts, attribute.Key("ConvertError").String("Encrypted PDF"))
		counter.Add(ctx, 1, metric.WithAttributes(opts...))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to read PDF content, check the file is not encrypted",
		})
		return
	}

	pageCount, err := pdf.GetPDFPageCount(pdfContent)
	if err != nil {
		opts = append(opts, attribute.Key("ConvertError").String("Missing Page Count"))
		counter.Add(ctx, 1, metric.WithAttributes(opts...))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to get PDF page count",
		})
		return
	}

	// Parse the page indices parameter
	pageIndicesParam := c.PostForm("pages")
	// TODO: get the total page of the PDF file
	pageIndices, err := util.ParsePageIndices(pageIndicesParam, pageCount)
	if err != nil {
		opts = append(opts, attribute.Key("ConvertError").String("Invalid PDF Page Indices"))
		counter.Add(ctx, 1, metric.WithAttributes(opts...))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Invalid page indices(%s): %s", pageIndicesParam, err.Error()),
		})
		return
	}

	// Set the default resolution to 300 dpi if not specified
	resolutionParam := c.PostForm("resolution")
	resolution, err := strconv.Atoi(resolutionParam)
	if err != nil || resolution <= 0 || resolution > 300 {
		resolution = 300
		log.Println("resolution is not set or exceeds the range (100-300), using default value 300")
	}

	exportParam := c.PostForm("export")
	exportFileType, ok := pdf.ImageExtensionMap[exportParam]
	if !ok {
		exportFileType = pdf.ImageExtensionMap["jpg"]
		log.Println("export is not set, using default value jpg")
	}

	// Log the export options and page indices
	log.Printf("Export Type: %v; Page Indices: %v; Resolution: %v", pdf.ImageTypeMap[exportFileType], pageIndices, resolution)
	// Convert the specified pages to PNG and add them to the zip file
	exportOptions := pdf.ExportOptions{
		Resolution: resolution,
		Format:     exportParam,
		Quality:    100,
	}
	childSpan.End()

	_, childSpan = tracer.Start(c.Request.Context(), "conversion-span")
	byteFile, err := pdf.ConvertPDFToImage(pdf.ConvertOptions{
		PDFFile:     pdfContent,
		PageIndices: pageIndices,
	}, exportOptions)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed to convert pages %v to Image (%v): %s", pageIndices, exportOptions, err.Error()),
		})
		return
	}
	childSpan.End()

	opts = append(opts, attribute.Key("ConvertSuccess").String("true"))
	duration.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(opts...))
	counter.Add(ctx, 1, metric.WithAttributes(opts...))
	// write the zip file to the response
	fileName := util.FileNameWithoutExt(pdf_file.Filename)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", fileName))
	c.Data(http.StatusOK, "application/octet-stream", byteFile)

}

// TODO: use find_trim to reduce the size of the image
// https://www.libvips.org/API/current/libvips-arithmetic.html#vips-find-trim

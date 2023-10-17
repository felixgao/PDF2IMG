package api

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestConvertRoute(t *testing.T) {
	router := gin.Default()
	RegisterConvertHandlers(router)

	// Create a request body with pdf file and form params
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("pages", "1-3,5,7,8")
	writer.WriteField("resolution", "150")
	writer.WriteField("export", "png")

	part, _ := writer.CreateFormFile("file[]", "fidelity.pdf")
	pdfContent, err := os.ReadFile("/Users/ggao/Desktop/fidelity.pdf")
	if err != nil {
		t.Fatalf("failed to read PDF file: %v", err)
	}
	part.Write(pdfContent)
	writer.Close()

	w := httptest.NewRecorder()
	//Since we registerd multiple POST route, one of them is good enough to test
	req, _ := http.NewRequest("POST", "/convert", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	// Check the test results
	assert.Equal(t, http.StatusOK, w.Code)
	if w.Code != http.StatusOK {
		t.Fatal("Error Message: ", w.Body.String())
	}
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/octet-stream", contentType)
	fileName := w.Header().Get("Content-Disposition")
	assert.Equal(t, "attachment; filename=fidelity.zip", fileName)
	// Read response body
	zipBytes, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	log.Printf("zipBytes size: %v", len(zipBytes))
	// Check if the output is not empty
	if len(zipBytes) == 0 {
		t.Fatal("empty compressed output")
	}
	// write the zip file to disk for manual inspection
	os.WriteFile("/Users/ggao/Downloads/fidelity.zip", zipBytes, 0644)
}

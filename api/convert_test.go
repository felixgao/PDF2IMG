package api

import (
	"bytes"
	"io"
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
	writer.WriteField("pages", "1-3")
	writer.WriteField("resolution", "150")
	writer.WriteField("export", "png")

	part, _ := writer.CreateFormFile("file[]", "ATT00001.pdf")
	pdfContent, err := os.ReadFile("/Users/ggao/Downloads/ATT00001.pdf")
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
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/octet-stream", contentType)
	fileName := w.Header().Get("Content-Disposition")
	assert.Equal(t, "attachment; filename=ATT00001.zip", fileName)
	// Read response body
	zipBytes, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	// Check if the output is not empty
	if len(zipBytes) == 0 {
		t.Fatal("empty compressed output")
	}

}

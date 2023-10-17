package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/felixgao/pdf_to_png/testutil"
)

func TestPingRoute(t *testing.T) {
	router := gin.Default()
	RegisterHealthCheckHandlers(router)

	// expected result from the service
	body := gin.H{
		"message": "ok",
	}

	w := httptest.NewRecorder()
	//Since we registerd multiple GET route, one of them is good enough to test
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	testutil.AssertMatchJSON(t, body, w.Body.String())
}

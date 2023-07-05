package main

import (
	"log"
	"net/http"

	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)
	defer vips.Shutdown()

	http.HandleFunc("/convert", convertHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

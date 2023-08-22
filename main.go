package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)
	defer vips.Shutdown()

	// defining router
	mux := http.NewServeMux()
	// starting server
	fmt.Println("Server is running at 127.0.0.1:8080")

	mux.HandleFunc("/convert", convertHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

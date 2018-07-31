package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("statics"))
	http.Handle("/", fs)

	log.Println("Listening...")
	http.ListenAndServe(":9000", nil)
}

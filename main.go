package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Server!")
	})

	const port string = ":8080"

	fmt.Println("Server Listening on port: ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("error starting server", err)
	}
}

package handlers

import "net/http"

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Server is Working!"))
}

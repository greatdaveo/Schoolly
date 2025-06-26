package handlers

import "net/http"

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Execs Routes is Working!"))
}

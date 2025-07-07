package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {
	tRouter := teachersRouter()
	sRouter := studentsRouter()

	tRouter.Handle("/", sRouter)
	return tRouter

	// mux := http.NewServeMux()

	// mux.HandleFunc("GET /", handlers.RootHandler)

	// mux.HandleFunc("GET /execs/", handlers.ExecsHandler)

	// return mux
}

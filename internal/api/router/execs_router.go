package router

import (
	"net/http"

	"github.com/greatdaveo/Schoolly/internal/api/handlers"
)

func execRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
	mux.HandleFunc("POST /execs", handlers.AddExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.EditMultipleExecsHandler)

	mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.EditExecSingleDataHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteOneExecHandler)
	
	return mux
}

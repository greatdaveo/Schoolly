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
	mux.HandleFunc("POST /execs/{id}/update-password", handlers.UpdatePassword)

	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	// mux.HandleFunc("POST /execs/forgot-password", handlers.AddStudentHandler)
	// mux.HandleFunc("POST /execs/reset-password/reset/{resetcode}", handlers.AddStudentHandler)

	return mux
}

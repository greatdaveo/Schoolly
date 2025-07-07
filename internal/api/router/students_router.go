package router

import (
	"net/http"

	"github.com/greatdaveo/Schoolly/internal/api/handlers"
)

func studentsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.EditMultipleStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("PUT /students/{id}", handlers.EditStudentHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.EditStudentSingleDataHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudentHandler)

	return mux
}

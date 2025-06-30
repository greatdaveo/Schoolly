package router

import (
	"net/http"

	"github.com/greatdaveo/Schoolly/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("GET /teachers/", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers/", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers/", handlers.EditMultipleTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.EditTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.EditTeacherSingleDataHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)

	mux.HandleFunc("/students/", handlers.StudentsHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}

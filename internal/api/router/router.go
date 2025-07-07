package router

import (
	"net/http"

	"github.com/greatdaveo/Schoolly/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.RootHandler)

	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers", handlers.EditMultipleTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.EditTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.EditTeacherSingleDataHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteOneTeacherHandler)

	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetStudentsForATeacher)
	mux.HandleFunc("GET /teachers/{id}/studentcount", handlers.CountStudentsForATeacher)

	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.EditMultipleStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("PUT /students/{id}", handlers.EditStudentHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.EditStudentSingleDataHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudentHandler)

	mux.HandleFunc("GET /execs/", handlers.ExecsHandler)

	return mux
}

package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {
	eRouter := execRouter()
	tRouter := teachersRouter()
	sRouter := studentsRouter()

	sRouter.Handle("/", eRouter)
	tRouter.Handle("/", sRouter)
	return tRouter
}

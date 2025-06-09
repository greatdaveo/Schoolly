package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/greatdaveo/Schoolly/internal/api/middlewares"
	mw "github.com/greatdaveo/Schoolly/internal/api/middlewares"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Server is Working!"))
}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Hello GET Method on Teachers Route is Working!"))
		return
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Students Routes is Working!"))
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Query: ", r.URL.Query().Get("name"))
	// To parse from data (necessary for x-www-form-urlencoded)
	err := r.ParseForm()
	if err != nil {
		return
	}

	fmt.Println("Form from POST methods: ", r.Form)
	w.Write([]byte("Hello Execs Routes is Working!"))
}

func main() {

	const port string = ":3000"

	// To load the cert file
	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)

	mux.HandleFunc("/teachers", teacherHandler)

	mux.HandleFunc("/students", studentsHandler)

	mux.HandleFunc("/execs", execsHandler)

	fmt.Println("Server Listening on port: ", port)

	// To create a TLS custom server
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Initialize the rate limiter
	rl := mw.NewRateLimiter(5, time.Minute)

	hppOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		WhiteList:                   []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	secureMux := mw.Hpp(hppOptions)(rl.RateLimiterMiddleware(mw.Compression(mw.ResponseTimeMiddleWare((middlewares.SecurityHeaders(mw.Cors(mux)))))))

	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(cert, key)

	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}

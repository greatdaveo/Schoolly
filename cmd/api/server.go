package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/greatdaveo/Schoolly/internal/api/middlewares"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Server is Working!"))
	fmt.Println("Hello Server is Working!")
}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Hello GET Method on Teachers Route is Working!"))
		fmt.Println("Hello Teachers Route is Working")
		return
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Students Routes is Working!"))
	fmt.Println("Hello Students Routes is Working!")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Execs Routes is Working!"))
	fmt.Println("Hello Execs Routes is Working!")
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

	server := &http.Server{
		Addr: port,
		// Handler:   middlewares.SecurityHeaders(mux),
		Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(cert, key)

	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}

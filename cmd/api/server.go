package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/server/router"
	mw "github.com/greatdaveo/Schoolly/internal/api/middlewares"
	"github.com/greatdaveo/Schoolly/internal/api/router"
)

func main() {

	const port string = ":3000"

	// To load the cert file
	cert := "cert.pem"
	key := "key.pem"

	fmt.Println("Server Listening on port: ", port)

	// To create a TLS custom server
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// // Initialize the rate limiter
	// rl := mw.NewRateLimiter(5, time.Minute)

	// hppOptions := mw.HPPOptions{
	// 	CheckQuery:                  true,
	// 	CheckBody:                   true,
	// 	CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	// 	WhiteList:                   []string{"sortBy", "sortOrder", "name", "age", "class"},
	// }

	// secureMux := mw.Cors(rl.RateLimiterMiddleware(mw.ResponseTimeMiddleWare(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOptions)(mux))))))
	// secureMux := utils.ApplyMiddlewares(mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleWare, rl.RateLimiterMiddleware, mw.Cors)
	router := router.Router()
	secureMux := mw.SecurityHeaders(router)

	// To create custom server
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

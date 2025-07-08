package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	mw "github.com/greatdaveo/Schoolly/internal/api/middlewares"
	"github.com/greatdaveo/Schoolly/internal/api/router"
	"github.com/greatdaveo/Schoolly/internal/models/repositories/sqlconnect"
	"github.com/greatdaveo/Schoolly/pkg/utils"
	"github.com/joho/godotenv"
)

func main() {
	// For .env
	err := godotenv.Load()
	if err != nil {
		return
	}

	// Database Connection
	_, err = sqlconnect.ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "❌ Database Connection Error ------ ")
		fmt.Println("❌ Database Connection Error ------ : ", err)
		return
	}

	// To load the cert file
	cert := "cert.pem"
	key := "key.pem"

	PORT := os.Getenv("API_PORT")

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
	router := router.MainRouter()
	secureMux := mw.SecurityHeaders(router)

	// To create custom server
	server := &http.Server{
		Addr:    PORT,
		Handler: secureMux,
		// Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server Listening on port: ", PORT)
	err = server.ListenAndServeTLS(cert, key)

	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}

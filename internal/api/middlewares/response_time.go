package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

// To track the duration taken to handle response from the client
func ResponseTimeMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("⏳ Received Request in ResponseTime Middleware")

		start := time.Now()

		// To create a custom ResponseWriter to capture the status code
		wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// To calculate the duration
		duration := time.Since(start)

		// This set by the API to make the client know the time the response took
		w.Header().Set("X-Response-Time", duration.String())

		next.ServeHTTP(wrappedWriter, r)

		duration = time.Since(start)
		// To Log the request details
		fmt.Printf(
			"Method: %s, URL: %s, Status: %d, Duration: %v\n",
			r.Method, r.URL, wrappedWriter.status, duration.String(),
		)

		fmt.Println("----------------------------------------------")

		fmt.Println("✅ Sent Response from ResponseTime Middleware")

	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

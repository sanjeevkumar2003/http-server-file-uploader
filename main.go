package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler, logfile *os.File) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		sourceIP := r.RemoteAddr
		userAgent := r.UserAgent()
		requestTime := time.Now().Format("2006-01-02 15:04:05")
		endpoint := r.URL.Path
		status := lrw.statusCode

		logLine := fmt.Sprintf("%s | %s | %s | %s | %d\n", sourceIP, userAgent, requestTime, endpoint, status)

		if _, err := logfile.WriteString(logLine); err != nil {
			log.Printf("Error writing log: %v", err)
		}
	})
}

func dateHandler(w http.ResponseWriter, r *http.Request) {
	currentDate := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, "Current Date and Time: %s", currentDate)
}

func main() {

	logfile, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logfile.Close()

	dateHandlerWithLogging := loggingMiddleware(http.HandlerFunc(dateHandler), logfile)

	http.Handle("/", dateHandlerWithLogging)

	fmt.Println("Server is running on :8083")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server Failed: %v", err)
	}
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Middleware struct {
	APIEndpoint string
	APIKey      string
	SourceName  string
}

func NewMiddleware() *Middleware {
	return &Middleware{
		APIEndpoint: "https://server-v3-7qxc7hlaka-uc.a.run.app/api/add",
		APIKey:      "your_api_key",
		SourceName:  "your_source_name",
	}
}

func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Read request body
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(strings.NewReader(string(requestBody))) // Reset r.Body
		}

		// Create a ResponseWriter to capture the response
		responseWriter := &ResponseCapture{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(responseWriter, r)

		elapsedTime := time.Since(startTime)

		headers := map[string]string{}
		for name, values := range r.Header {
			headers[name] = strings.Join(values, ",")
		}

		var requestBodyParsed interface{}
		if r.Header.Get("Content-Type") == "application/json" {
			json.Unmarshal(requestBody, &requestBodyParsed)
		}

		var responseContentParsed interface{}
		json.Unmarshal(responseWriter.body, &responseContentParsed)

		data := map[string]interface{}{
			"request": map[string]interface{}{
				"method":  r.Method,
				"path":    r.URL.String(),
				"headers": headers,
				"body":    requestBodyParsed,
			},
			"response": map[string]interface{}{
				"status_code": responseWriter.statusCode,
				"content":     responseContentParsed,
			},
			"elapsed_time": elapsedTime.Seconds(),
		}

		go m.sendDataToAPI(data)
	})
}

func (m *Middleware) sendDataToAPI(data map[string]interface{}) {
	if m.APIKey != "" && m.SourceName != "" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshalling data: %v", err)
			return
		}

		req, err := http.NewRequest("POST", m.APIEndpoint, strings.NewReader(string(jsonData)))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-access-token", m.APIKey)
		req.Header.Set("source-name", m.SourceName)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request to API: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to send data to API endpoint. Status code: %d", resp.StatusCode)
		} else {
			log.Println("Devzery: Success!")
		}
	} else {
		if m.APIKey == "" {
			log.Println("Devzery: No API Key")
		}
		if m.SourceName == "" {
			log.Println("Devzery: No Source Name")
		}
	}
}

type ResponseCapture struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *ResponseCapture) Write(b []byte) (int, error) {
	rc.body = append(rc.body, b...)
	return rc.ResponseWriter.Write(b)
}

func main() {
	r := mux.NewRouter()

	mw := NewMiddleware()
	r.Use(mw.LoggingMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}).Methods("GET")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

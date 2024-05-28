package loggingMiddleware

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Middleware struct {
	APIEndpoint string
	APIKey      string
	SourceName  string
}

type contextKey string

const ErrorKey contextKey = "ErrorKey"

func NewMiddleware(apiEndpoint, apiKey, sourceName string) *Middleware {
	return &Middleware{
		APIEndpoint: apiEndpoint,
		APIKey:      apiKey,
		SourceName:  sourceName,
	}
}

func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(strings.NewReader(string(requestBody)))
		}

		responseWriter := &ResponseCapture{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(responseWriter, r)

		elapsedTime := time.Since(startTime)

		headers := map[string]string{}
		for name, values := range r.Header {
			headers[name] = strings.Join(values, ",")
		}

		var requestBodyParsed interface{}
		contentType := r.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(contentType, "application/json"):
			if err := json.Unmarshal(requestBody, &requestBodyParsed); err != nil {
				log.Printf("Error unmarshalling request body: %v", err)

				requestBodyParsed = nil

			}
		case strings.HasPrefix(contentType, "multipart/form-data"):
			if err := r.ParseMultipartForm(1024 * 1024); err != nil {
				log.Printf("Error parsing multipart form: %v", err)

				requestBodyParsed = nil

			} else {
				requestBodyParsed = r.Form
			}
		case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)

				requestBodyParsed = nil

			} else {
				requestBodyParsed = r.Form
			}
		}

		var responseContentParsed interface{}
		if err := json.Unmarshal(responseWriter.body, &responseContentParsed); err != nil {
			log.Printf("Error unmarshalling response body: %v", err)
			responseContentParsed = nil
		}

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
	go func() {
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
				return
			}
			log.Println("Devzery: Success!")
		} else {
			if m.APIKey == "" {
				log.Println("Devzery: No API Key")
			}
			if m.SourceName == "" {
				log.Println("Devzery: No Source Name")
			}
		}
	}()
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

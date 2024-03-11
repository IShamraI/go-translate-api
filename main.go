package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/IShamraI/go-translate-api/internal/counter"
	googletranslate "github.com/IShamraI/go-translate-api/internal/google_translate"
)

var (
	cnt  *counter.Counter = counter.New()
	port int

	translateService *googletranslate.GTranslateService = googletranslate.NewRuEn()
)

type Response struct {
	Text  string `json:"text"`
	Error string `json:"error"`
}

func (r *Response) ToJson() []byte {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"error": "Error marshalling json"}`)
	}
	return jsonData
}

func init() {
	flag.IntVar(&port, "port", 8080, "Port number for the service to listen on")
	flag.Parse()
}

func splitText(text string) []string {
	var maxChunkSize = 5000
	var chunks []string
	sentences := strings.Split(text, ".")
	currentChunk := ""

	for _, sentence := range sentences {
		// Check if adding the current sentence to the current chunk would exceed the max size
		if len(currentChunk)+len(sentence)+1 > maxChunkSize {
			currentChunk += "."
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = ""
		}
		// Add the current sentence to the current chunk
		if currentChunk != "" {
			currentChunk += ". "
		}
		currentChunk += sentence
	}

	// Add the last chunk
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

func handleTextSplit(w http.ResponseWriter, r *http.Request) {
	cnt.AddRequest()
	log.Printf("Request %d: %s", cnt.Requests, r.URL.Path)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")

	if text == "" {
		http.Error(w, "Please provide text in the 'text' field.", http.StatusBadRequest)
		return
	}

	chunks := splitText(text)
	cnt.AddLetters(len(text))
	log.Printf("Chunks: %d", len(chunks))

	// Return the split chunks as a JSON array
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var response Response
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		tChunk, err := translateService.Translate(chunk)
		if err != nil {
			log.Printf("Error translating chunk: %s", err)
			response.Error = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			response.Text = ""
			break
		}
		response.Text += fmt.Sprintf(" %s", tChunk)
		time.Sleep(time.Duration(500) * time.Millisecond) // Add a 500ms delay between chunks
	}
	w.Write(response.ToJson())
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	cnt.AddRequest()
	log.Printf("Request %d: %s", cnt.Requests, r.URL.Path)
	// Return the statistics as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(cnt.ToJson())
}

func main() {
	http.HandleFunc("/split", handleTextSplit)
	http.HandleFunc("/stats", handleStats)
	address := "0.0.0.0:" + strconv.Itoa(port)
	log.Printf("Listening on %s", address)
	http.ListenAndServe(address, nil)
}

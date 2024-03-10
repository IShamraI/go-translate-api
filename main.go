package main

import (
    "bufio"
    "bytes"
    "flag"
    "net/http"
    "strings"
    "sync"
    "strconv"
)

var (
    requestCount int
    letterCount  int
    mutex        sync.Mutex
    port         int
)

func init() {
    flag.IntVar(&port, "port", 8080, "Port number for the service to listen on")
    flag.Parse()
}

func splitText(text string) []string {
    const maxChunkSize = 5000

    paragraphs := strings.Split(text, "\n\n") // Split text into paragraphs
    var chunks []string
    var currentChunk bytes.Buffer

    for _, para := range paragraphs {
        if len(para) <= maxChunkSize {
            currentChunk.WriteString(para + "\n\n") // Add paragraph to current chunk
            mutex.Lock()
            letterCount += len(para)
            mutex.Unlock()
        } else {
            scanner := bufio.NewScanner(bytes.NewBufferString(para))
            scanner.Split(bufio.ScanWords)
            var chunkBuffer bytes.Buffer
            chunkBuffer.Grow(maxChunkSize)

            for scanner.Scan() {
                word := scanner.Text()
                if len(chunkBuffer.String())+len(word)+1 <= maxChunkSize { // +1 to account for the space after the word
                    chunkBuffer.WriteString(word)
                    chunkBuffer.WriteString(" ")
                } else {
                    chunks = append(chunks, chunkBuffer.String())
                    chunkBuffer.Reset()
                    chunkBuffer.WriteString(word)
                    chunkBuffer.WriteString(" ")
                }
            }
            if chunkBuffer.Len() > 0 {
                chunks = append(chunks, chunkBuffer.String())
            }
        }
    }

    if currentChunk.Len() > 0 {
        chunks = append(chunks, currentChunk.String())
    }

    return chunks
}

func handleTextSplit(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    requestCount++
    
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

    // Return the split chunks as a JSON array
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("["))
    for i, chunk := range chunks {
        if i > 0 {
            w.Write([]byte(","))
        }
        w.Write([]byte(`"` + chunk + `"`))
    }
    w.Write([]byte("]"))
}

func handleStats(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    stats := map[string]int{
        "requestCount": requestCount,
        "letterCount":  letterCount,
    }

    // Return the statistics as JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"requestCount":` + strconv.Itoa(requestCount) + `,"letterCount":` + strconv.Itoa(letterCount) + `}`))
}

func main() {
    http.HandleFunc("/split", handleTextSplit)
    http.HandleFunc("/stats", handleStats)
    http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

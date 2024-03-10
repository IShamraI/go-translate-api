package counter

import (
	"encoding/json"
	"sync"
)

type Counter struct {
	Requests int `json:"requestCount"`
	Letters  int `json:"letterCount"`
	sync.Mutex
}

func (c *Counter) AddRequest() {
	c.Lock()
	c.Requests++
	c.Unlock()
}

func (c *Counter) AddLetters(letters int) {
	c.Lock()
	c.Letters += letters
	c.Unlock()
}

func (c *Counter) ToJson() []byte {
	jsonStats, err := json.Marshal(c)
	if err != nil {
		return []byte(`{"error": "Error marshalling json"}`)
	}
	return jsonStats
}

func (c *Counter) Reset() {
	c.Lock()
	c.Requests = 0
	c.Letters = 0
	c.Unlock()
}

func New() *Counter {
	return &Counter{Requests: 0, Letters: 0}
}

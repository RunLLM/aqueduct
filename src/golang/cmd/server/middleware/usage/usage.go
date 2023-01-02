package usage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	obfuscated string = "***"
	delimiter  string = "/"
	hashKey    string = "aqueduct"
	// This is the Grafana Loki server address
	logURL string = "http://34.25.83.71:9090/loki/api/v1/push"
)

type UsageStats struct {
	ID      string `json:"id"`
	Route   string `json:"route"`
	Latency int64  `json:"latency"`
}

func reportUsage(startTime time.Time, r *http.Request) {
	pathToken := strings.Split(r.URL.Path, delimiter)
	for i, token := range pathToken {
		if _, err := uuid.Parse(token); err == nil {
			pathToken[i] = obfuscated
		}
	}

	machineID, err := machineid.ProtectedID(hashKey)
	if err != nil {
		log.Errorf("Failed to generate device ID: %v", err)
		return
	}

	usage := UsageStats{
		ID:      machineID,
		Route:   strings.Join(pathToken, delimiter),
		Latency: time.Since(startTime).Milliseconds(),
	}

	log.Errorf("This request took %d ms, request URL path: %s, machine id: %s.", usage.Latency, usage.Route, usage.ID)

	payload, err := json.Marshal(usage)
	if err != nil {
		log.Errorf("Failed to marshal usage stats: %v", err)
		return
	}

	req, err := http.NewRequest("POST", logURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Errorf("Failed to initialize usage stats POST request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func WithUsageStats() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			defer func() {
				go reportUsage(startTime, r)
			}()

			h.ServeHTTP(w, r)
		})
	}
}

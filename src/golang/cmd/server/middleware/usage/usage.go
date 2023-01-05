package usage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func reportUsage(startTime time.Time, environment string, statusCode int, urlPath string) {
	pathToken := strings.Split(urlPath, delimiter)
	for i, token := range pathToken {
		if _, err := uuid.Parse(token); err == nil {
			pathToken[i] = obfuscated
		}
	}

	// This call generates a unique hash of the host device in a privacy-preserving fashion.
	// Details can be found here: https://github.com/denisbrodbeck/machineid
	machineID, err := machineid.ProtectedID(hashKey)
	if err != nil {
		log.Errorf("Failed to generate obfuscated device ID: %v", err)
		return
	}

	startTimeUnix := startTime.UnixNano()

	labels := Labels{
		SchemaVersion: schemaVersion,
		Environment:   environment,
		StatusCode:    statusCode,
		Route:         strings.Join(pathToken, delimiter),
	}

	payload := Payload{
		ID:      machineID,
		Latency: time.Since(startTime).Milliseconds(),
		Labels:  labels,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal payload: %v", err)
		return
	}

	streams := Streams{
		Streams: []Stream{
			{
				Labels: labels,
				Values: [][]string{
					{
						strconv.FormatInt(startTimeUnix, 10),
						string(payloadJson),
					},
				},
			},
		},
	}

	streamsJson, err := json.Marshal(streams)
	if err != nil {
		log.Errorf("Failed to marshal streams: %v", err)
		return
	}

	//log.Errorf("printing json log: %v", string(streamsJson))
	//log.Errorf("payload: %v", streams.Streams[0].Values[0][1])

	req, err := http.NewRequest("POST", logURL, bytes.NewBuffer(streamsJson))
	if err != nil {
		log.Errorf("Failed to initialize usage stats POST request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send request to loki: %v", err)
		panic(err)
	}
	defer resp.Body.Close()
}

func WithUsageStats(environment string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			startTime := time.Now()
			defer func() {
				go reportUsage(startTime, environment, ww.Status(), r.URL.Path)
			}()

			h.ServeHTTP(ww, r)
		})
	}
}

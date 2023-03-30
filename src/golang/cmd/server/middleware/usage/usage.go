package usage

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func reportUsage(
	startTime time.Time,
	environment string,
	statusCode int,
	urlPath string,
	machineID string,
) {
	pathToken := strings.Split(urlPath, delimiter)
	for i, token := range pathToken {
		if _, err := uuid.Parse(token); err == nil {
			pathToken[i] = obfuscated
		}
	}

	startTimeUnix := startTime.UnixNano()

	// Loki creates indexes for labels to speed up searching. Each label should have a bounded number
	// of distinct values to prevent the index from getting too large. That's why fields such as ID
	// and Latency should not be included as labels.
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
				// According to Loki's requirement, the first field should be the timestamp in ns,
				// and the second field should be the payload in its json form.
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
		return
	}
	defer resp.Body.Close()
}

func WithUsageStats(environment string) func(http.Handler) http.Handler {
	var machineID string
	// Determine codespace ID based on
	// https://docs.github.com/en/codespaces/developing-in-codespaces/default-environment-variables-for-your-codespace
	if os.Getenv(codespacesEnv) == codespacesEnvActiveValue {
		environment = "code_space"
		machineID = fmt.Sprintf("%x", md5.Sum([]byte(os.Getenv(codespaceNameEnv))))
	} else {
		// This call generates a unique hash of the host device in a privacy-preserving fashion.
		// Details can be found here: https://github.com/denisbrodbeck/machineid
		var err error
		machineID, err = machineid.ProtectedID(hashKey)
		if err != nil {
			log.Errorf("Failed to generate obfuscated device ID: %v", err)
			return func(h http.Handler) http.Handler { return h }
		}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The reason we need this wrapper is so that we can get the status of the response via
			// ww.Status().
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			startTime := time.Now()
			defer func() {
				go reportUsage(startTime, environment, ww.Status(), r.URL.Path, machineID)
			}()

			h.ServeHTTP(ww, r)
		})
	}
}

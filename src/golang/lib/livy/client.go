package livy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dropbox/godropbox/errors"
)

// LivyClient represents a client for connecting to a Livy server.
type LivyClient struct {
	LivyServerURL string
	Client        *http.Client
}

// NewLivyClient creates a new LivyClient.
func NewLivyClient(livyServerURL string) *LivyClient {
	return &LivyClient{
		LivyServerURL: livyServerURL,
		Client:        &http.Client{},
	}
}

// Creates a SparkSession on Spark Cluster.
func (c *LivyClient) CreateSession(sessionReq *CreateSessionRequest) (*Session, error) {
	url := fmt.Sprintf("%s/sessions", c.LivyServerURL)
	body, err := json.Marshal(sessionReq)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling session request.")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating session request.")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending session request.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Newf("Failed to create session: %s", resp.Status)
	}

	// Decode the response JSON into a Session struct
	var session Session
	err = json.NewDecoder(resp.Body).Decode(&session)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding session response.")
	}

	return &session, nil
}

// Gets a particular SparkSession given the ID.
func (c *LivyClient) GetSession(id int) (*Session, error) {
	url := fmt.Sprintf("%s/sessions/%d", c.LivyServerURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating get session request.")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending get session request.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("Failed to get session %d: %s", id, resp.Status)
	}

	// Decode the response JSON into a Session struct
	var session Session
	err = json.NewDecoder(resp.Body).Decode(&session)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding get session response.")
	}

	return &session, nil
}

// Gets all active sessions on Spark Cluster.
func (c *LivyClient) GetSessions() ([]Session, error) {
	url := fmt.Sprintf("%s/sessions", c.LivyServerURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating getSessions request.")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending getSessions request.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("Failed to get getSessions: %s", resp.Status)
	}

	// Decode the response JSON into a slice of Session structs
	var getSessionsResponse GetSessionsResponse
	err = json.NewDecoder(resp.Body).Decode(&getSessionsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding getSessions response.")
	}

	return getSessionsResponse.Sessions, nil
}

func (c *LivyClient) WaitForSession(sessionID int, timeout time.Duration) error {
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return errors.New("Timed out waiting for SparkSession to create.")
		}
		s, err := c.GetSession(sessionID)
		if err != nil {
			return errors.Wrap(err, "Error retrieving session while waiting for creation.")
		}
		if s.State == Idle {
			break
		}

		time.Sleep(time.Second * 1)
	}
	return nil
}

// RunStatement creates a new statement for a given batch
func (c *LivyClient) RunStatement(sessionID int, statement *StatementRequest) (*Statement, error) {
	url := fmt.Sprintf("%s/sessions/%d/statements", c.LivyServerURL, sessionID)
	body, err := json.Marshal(statement)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling statement request.")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating statement request.")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending statement request.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Newf("Error creating statment: %v", resp.Status)
	}

	var s Statement
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding statement response.")
	}

	return &s, nil
}

// GetStatement retrieves a statement by its ID
func (c *LivyClient) GetStatement(sessionID int, statementID int) (*Statement, error) {
	u := fmt.Sprintf("%s/sessions/%d/statements/%d", c.LivyServerURL, sessionID, statementID)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating get statement request.")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending get statement request.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("Error creating get statment: %v", resp.Status)
	}

	var s Statement
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding get statement response.")
	}

	return &s, nil
}

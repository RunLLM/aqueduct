package livy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	client *LivyClient
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewLivyClient(server.URL)

	return func() {
		server.Close()
	}
}

func TestCreateSession(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		respSession := &Session{
			ID:    1,
			State: Idle,
		}
		resp, _ := json.Marshal(respSession)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})

	// Call CreateSession with expected
	sessionReq := &CreateSessionRequest{
		Kind: "pyspark",
	}
	session, err := client.CreateSession(sessionReq)

	// Define expected result
	expectedSession := &Session{
		ID:    1,
		State: Idle,
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
}

func TestGetSession(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	expectedSession := &Session{
		ID:    1,
		State: Idle,
	}
	mux.HandleFunc("/sessions/1", func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(expectedSession)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	// Call GetSession with session ID 1
	session, err := client.GetSession(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
}

func TestRunStatement(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	expectedStatement := &Statement{
		ID:     1,
		State:  Running,
		Output: StatementOutput{},
	}

	mux.HandleFunc("/sessions/1/statements", func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(expectedStatement)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})

	// Call RunStatement with session ID 1 and code "print('hello world')"
	statementReq := &StatementRequest{
		Code: "print('hello world')",
	}
	statement, err := client.RunStatement(1, statementReq)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatement, statement)
}

func TestGetStatement(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	expectedStatement := &Statement{
		ID:    1,
		State: Available,
		Output: StatementOutput{
			Status:         OK,
			ExecutionCount: 1,
			Data:           map[string]interface{}{},
		},
	}

	mux.HandleFunc("/sessions/1/statements/1", func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(expectedStatement)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	// Call GetStatement with session ID 1 and statement ID 1
	statement, err := client.GetStatement(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatement, statement)
}

func TestGetSessions(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	// Define expected result
	expectedSessions := []*Session{
		{ID: 1, State: Starting},
		{ID: 2, State: Idle},
		{ID: 3, State: Busy},
	}

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(expectedSessions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	// Call GetSessions with expected
	sessions, err := client.GetSessions()

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
}

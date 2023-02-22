package livy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		jsonResp := `{"id": 1, "state": "idle"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(jsonResp))
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

	mux.HandleFunc("/sessions/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp := `{"id": 1, "state": "idle"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})
	expectedSession := &Session{
		ID:    1,
		State: Idle,
	}
	// Call GetSession with session ID 1
	session, err := client.GetSession(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
}

func TestRunStatement(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	mux.HandleFunc("/sessions/1/statements", func(w http.ResponseWriter, r *http.Request) {
		jsonResp := `{"id": 1, "state": "running", "output": {}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(jsonResp))
	})

	expectedStatement := &Statement{
		ID:     1,
		State:  Running,
		Output: StatementOutput{},
	}

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

	mux.HandleFunc("/sessions/1/statements/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp := `{"id": 1, "state": "available", "output": {"status": "ok", "execution_count": 1, "data": {}}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	expectedStatement := &Statement{
		ID:    1,
		State: Available,
		Output: StatementOutput{
			Status:         OK,
			ExecutionCount: 1,
			Data:           map[string]interface{}{},
		},
	}

	// Call GetStatement with session ID 1 and statement ID 1
	statement, err := client.GetStatement(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatement, statement)
}

func TestGetSessions(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	// Define expected result

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		jsonResp := `{"sessions": [{"id": 1, "state": "starting"}, {"id": 2, "state": "idle"}, {"id": 3, "state": "busy"}]}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	expectedSessions := []Session{
		{ID: 1, State: Starting},
		{ID: 2, State: Idle},
		{ID: 3, State: Busy},
	}

	// Call GetSessions with expected
	sessions, err := client.GetSessions()

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
}

func TestWaitForSession(t *testing.T) {
	cleanup := setup()
	defer cleanup()

	mux.HandleFunc("/sessions/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp := `{"id": 1, "state": "not_started"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	err := client.WaitForSession(1, time.Second*2)
	expectedErrorMsg := "Timed out waiting for SparkSession to create."

	assert.NotEmpty(t, err)
	assert.Containsf(t, err.Error(), expectedErrorMsg, "expected error containing %q, got %s", expectedErrorMsg, err)
}

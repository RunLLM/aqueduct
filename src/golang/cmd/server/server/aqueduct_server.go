package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/authentication"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/maintenance"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/request_id"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/usage"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/logging"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
)

const (
	accountOrganizationId = "aqueduct"

	// The maximum number of entries this cache can have.
	previewCacheSize = 200
)

var uiDir = path.Join(os.Getenv("HOME"), ".aqueduct", "ui")

type AqServer struct {
	Router   *chi.Mux
	Name     string
	Database database.Database
	*Repos

	// Only the following group of fields will be reinitialized when the server is restarted
	GithubManager github.Manager
	// TODO ENG-1483: Move JobManager from Server to Handlers
	JobManager job.JobManager
	AqEngine   engine.AqEngine
	AqPath     string

	// UnderMaintenance indicates whether the server is currently down for system maintenance.
	UnderMaintenance atomic.Value
	// RequestMutex's read lock is acquired and released by each request to indicate when there
	// are no more active requests.
	RequestMutex sync.RWMutex

	// The environment in which the server runs. This is for usage stats collection purpose.
	Environment       string
	DisableUsageStats bool
}

func NewAqServer(environment string, disableUsageStats bool) *AqServer {
	ctx := context.Background()
	aqPath := config.AqueductPath()

	// The database cannot be reinitialized when the server restarts, because the database is passed
	// to the middleware functions.
	db, err := database.NewSqliteDatabase(&database.SqliteConfig{
		File: path.Join(aqPath, database.SqliteDatabasePath),
	})
	if err != nil {
		log.Fatalf("Unable to initialize database: %v", err)
	}

	if err := requireSchemaVersion(
		context.Background(),
		models.CurrentSchemaVersion,
		sqlite.NewSchemaVersionRepo(),
		db,
	); err != nil {
		db.Close()
		log.Fatalf("Unable to confirm required database schema version: %v", err)
	}

	s := &AqServer{
		Router:            chi.NewRouter(),
		Database:          db,
		Repos:             CreateRepos(),
		UnderMaintenance:  atomic.Value{},
		RequestMutex:      sync.RWMutex{},
		Environment:       environment,
		DisableUsageStats: disableUsageStats,
	}
	s.UnderMaintenance.Store(false)

	// Initialize the other server fields
	if err := s.Init(); err != nil {
		db.Close()
		log.Fatalf("Unable to initialize server: %v", err)
	}

	allowedOrigins := []string{"*"}
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: GetAllHeaders(s),
		AllowedMethods: []string{"GET", "POST"},
	})
	s.Router.Use(corsMiddleware.Handler)
	s.Router.Use(middleware.Logger)

	// Register server handlers
	AddAllHandlers(s)

	log.Infof("Creating a user account and a builtin SQLite integration.")
	testUser, err := CreateTestAccount(
		ctx,
		s,
		config.APIKey(),
		accountOrganizationId,
	)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	demoConnected, err := CheckBuiltinIntegration(ctx, s, accountOrganizationId)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	if !demoConnected {
		err = ConnectBuiltinIntegration(ctx, testUser, s.IntegrationRepo, s.Database)
		if err != nil {
			db.Close()
			log.Fatal(err)
		}
	}

	err = s.initializeWorkflowCronJobs(ctx)
	if err != nil {
		db.Close()
		log.Fatalf("Failed to create cron jobs for existing workflows: %v", err)
	} else {
		log.Info("Successfully created cron jobs for existing workflows")
	}

	return s
}

// Init sets all of the fields of this AqServer that depend on server configuration.
func (s *AqServer) Init() error {
	aqPath := config.AqueductPath()

	githubManager := github.NewUnimplementedManager()

	jobManager, err := job.NewProcessJobManager(
		&job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
		},
	)
	if err != nil {
		return err
	}

	storageConfig := config.Storage()

	previewCacheManager, err := preview_cache.NewInMemoryPreviewCacheManager(
		&storageConfig,
		previewCacheSize,
	)
	if err != nil {
		return err
	}

	vault, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return err
	}

	if err := syncVaultWithStorage(
		vault,
		s.IntegrationRepo,
		s.Database,
	); err != nil {
		return err
	}

	eng, err := engine.NewAqEngine(
		s.Database,
		githubManager,
		previewCacheManager,
		aqPath,
		GetEngineRepos(s.Repos),
	)
	if err != nil {
		return err
	}

	s.GithubManager = githubManager
	s.JobManager = jobManager
	s.AqPath = aqPath
	s.AqEngine = eng

	return nil
}

func (s *AqServer) StartWorkflowRetentionJob(period string) error {
	name := job.WorkflowRetentionName
	ctx := context.Background()

	// Delete old CronJob if it exists
	err := s.JobManager.DeleteCronJob(ctx, name)
	if err != nil {
		return errors.Wrap(err, "Unable to delete existing workflow retention job")
	}

	spec := job.NewWorkflowRetentionJobSpec(
		s.Database.Config(),
		s.JobManager.Config(),
	)

	err = s.JobManager.DeployCronJob(
		ctx,
		name,
		period,
		spec,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to start workflow retention cron job")
	}
	return nil
}

func (s *AqServer) AddHandler(route string, handlerObj handler.Handler) {
	middleware := alice.New()

	if !s.DisableUsageStats {
		middleware = middleware.Append(usage.WithUsageStats(s.Environment))
	}

	if handlerObj.AuthMethod() == handler.ApiKeyAuthMethod {
		middleware = middleware.Append(
			maintenance.Check(&s.UnderMaintenance),
			request_id.WithRequestId(),
			authentication.RequireApiKey(s.UserRepo, s.Database),
		)
	} else {
		panic(errors.New("Auth method is not supported."))
	}

	s.Router.Method(
		string(handlerObj.Method()),
		route,
		middleware.ThenFunc(ExecuteHandler(s, handlerObj)),
	)
}

func convertToSet(arr []string) map[string]bool {
	set := make(map[string]bool, len(arr))
	for _, elem := range arr {
		set[elem] = true
	}
	return set
}

func (s *AqServer) Log(ctx context.Context, key string, req *http.Request, statusCode int, err error) {
	excludedHeaderFields := convertToSet([]string{
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Api-Key",
		"Connection",
		"Content-Type",
		"Origin",
		"User-Agent",
		"Referer",
	})

	logging.LogRoute(ctx, key, req, excludedHeaderFields, statusCode, logging.ServerComponent, s.Name, err)
}

func (s *AqServer) Run(expose bool, port int) {
	// When we configure the server to listen on ":<PORT>" (without specifying the ip), it exposes itself
	// to the public.
	ip := ""
	if !expose {
		ip = "localhost"
	}

	static := http.FileServer(http.Dir(uiDir))
	s.Router.Method("GET", "/dist/*", http.StripPrefix("/dist/", static))
	s.Router.Get("/*", IndexHandler())

	log.Infof("%s Starting HTTP server on port %d\n", time.Now().Format("2006-01-02 03:04:05 PM"), port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), s.Router))
}

func IndexHandler() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fmt.Sprintf("%s/index.html", uiDir))
	}

	return http.HandlerFunc(fn)
}

// Pause puts the server in system maintenance mode by blocking all new requests
// and waits for all active requests to finish.
// It is the responsibility of the caller to call s.Restart() to allow requests
// to be processed again once the system maintenance is complete.
func (s *AqServer) Pause() {
	s.UnderMaintenance.Store(true)
	s.RequestMutex.Lock()
}

// Restart restarts a server that was previously stopped via s.Pause().
func (s *AqServer) Restart() {
	if err := s.Init(); err != nil {
		log.Fatalf("Unable to restart server: %v", err)
	}
	s.RequestMutex.Unlock()
	s.UnderMaintenance.Store(false)
}

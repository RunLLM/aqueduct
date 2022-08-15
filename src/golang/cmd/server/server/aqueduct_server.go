package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/authentication"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/request_id"
	"github.com/aqueducthq/aqueduct/cmd/server/middleware/verification"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/logging"
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
	RequiredSchemaVersion = 15

	accountOrganizationId = "aqueduct"

	// The maximum number of entries this cache can have.
	previewCacheSize = 200
)

var uiDir = path.Join(os.Getenv("HOME"), ".aqueduct", "ui")

type AqServer struct {
	Router *chi.Mux

	Name          string
	Database      database.Database
	GithubManager github.Manager
	// TODO ENG-1483: Move JobManager from Server to Handlers
	JobManager job.JobManager
	Vault      vault.Vault
	AqEngine   engine.AqEngine
	AqPath     string
	*Readers
	*Writers
}

func NewAqServer() *AqServer {
	ctx := context.Background()
	aqPath := config.AqueductPath()
	db, err := database.NewSqliteDatabase(&database.SqliteConfig{
		File: path.Join(aqPath, database.SqliteDatabasePath),
	})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	githubManager := github.NewUnimplementedManager()

	jobManager, err := job.NewProcessJobManager(
		&job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
		},
	)
	if err != nil {
		db.Close()
		log.Fatal("Unable to create job manager: ", err)
	}

	vault, err := vault.NewFileVault(&vault.FileConfig{
		Directory:     path.Join(aqPath, vault.FileVaultDir),
		EncryptionKey: config.EncryptionKey(),
	})
	if err != nil {
		db.Close()
		log.Fatal("Unable to start vault: ", err)
	}

	readers, err := CreateReaders(db.Config())
	if err != nil {
		db.Close()
		log.Fatal("Unable to create readers: ", err)
	}

	writers, err := CreateWriters(db.Config())
	if err != nil {
		db.Close()
		log.Fatal("Unable to create writers: ", err)
	}

	previewCacheManager, err := preview_cache.NewInMemoryPreviewCacheManager(
		conf.StorageConfig,
		previewCacheSize,
	)
	if err != nil {
		log.Fatal("Unable to create preview artifact cache: ", err)
	}

	eng, err := engine.NewAqEngine(
		db,
		githubManager,
		previewCacheManager,
		vault,
		aqPath,
		GetEngineReaders(readers),
		GetEngineWriters(writers),
	)
	if err != nil {
		log.Fatal("Unable to create aqEngine: ", err)
	}

	s := &AqServer{
		Router:        chi.NewRouter(),
		Database:      db,
		GithubManager: github.NewUnimplementedManager(),
		JobManager:    jobManager,
		Vault:         vault,
		AqPath:        aqPath,
		AqEngine:      eng,
		Readers:       readers,
		Writers:       writers,
	}

	allowedOrigins := []string{"*"}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: GetAllHeaders(s),
		AllowedMethods: []string{"GET", "POST"},
	})
	s.Router.Use(corsMiddleware.Handler)
	s.Router.Use(middleware.Logger)
	AddAllHandlers(s)

	if err := collections.RequireSchemaVersion(
		context.Background(),
		RequiredSchemaVersion,
		s.SchemaVersionReader,
		db,
	); err != nil {
		db.Close()
		log.Fatalf("Found incompatible database schema version: %v", err)
	}

	log.Infof("Creating a user account and a builtin SQLite integration.")
	testUser, err := CreateTestAccount(
		ctx,
		s,
		"",
		"",
		"",
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
		err = ConnectBuiltinIntegration(ctx, testUser, s.IntegrationWriter, s.Database, s.Vault)
		if err != nil {
			db.Close()
			log.Fatal(err)
		}
	}

	err = s.initializeWorkflowCronJobs(ctx)
	if err != nil {
		log.Fatalf("Failed to create cron jobs for existing workflows: %v", err)
	} else {
		log.Info("Successfully created cron jobs for existing workflows")
	}

	return s
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
		s.Vault.Config(),
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
	var middleware alice.Chain
	if handlerObj.AuthMethod() == handler.ApiKeyAuthMethod {
		middleware = alice.New(
			request_id.WithRequestId(),
			authentication.RequireApiKey(s.UserReader, s.Database),
			verification.VerifyRequest(),
		)
	} else {
		panic(handler.ErrUnsupportedAuthMethod)
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

module github.com/aqueducthq/aqueduct

go 1.16

require (
	cloud.google.com/go/storage v1.25.0
	github.com/apache/airflow-client-go/airflow v0.0.0-20220509204651-4f1b26e4a5d0
	github.com/aws/aws-sdk-go v1.40.33
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dropbox/godropbox v0.0.0-20200228041828-52ad444d3502
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/cors v1.2.0
	github.com/go-co-op/gocron v1.13.0
	github.com/google/go-github/v40 v40.0.0
	github.com/google/uuid v1.3.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/hashicorp/golang-lru v0.5.1
	github.com/jackc/pgx/v4 v4.13.0
	github.com/justinas/alice v1.2.0
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2
	google.golang.org/api v0.88.0
	google.golang.org/grpc v1.48.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
)

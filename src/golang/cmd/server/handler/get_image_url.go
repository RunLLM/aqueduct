package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/container_registry"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /api/integration/container-registry/url
// Method: GET
// Params: None
// Request:
//
//		Headers:
//			`api-key`: user's API Key
//			`integration_id`: container registry integration ID
//	     `service`: name of the service to get the URL for
//	     `image_name`: name of the image to get the URL for
//
// Response: serialized `getImageURLResponse` which contains the url for the image.
type GetImageURLHandler struct {
	GetHandler

	Database database.Database

	ResourceRepo repos.Resource
}

type getImageURLArgs struct {
	*aq_context.AqContext
	integrationID uuid.UUID
	service       shared.Service
	imageName     string
}

type getImageURLResponse struct {
	Url string `json:"url"`
}

func (*GetImageURLHandler) Name() string {
	return "GetImageURL"
}

func (*GetImageURLHandler) Headers() []string {
	return []string{routes.IntegrationIDHeader, routes.ServiceHeader, routes.ImageNameHeader}
}

func (*GetImageURLHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIdStr := r.Header.Get(routes.IntegrationIDHeader)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error parsing integration ID.")
	}

	return &getImageURLArgs{
		AqContext:     aqContext,
		integrationID: integrationId,
		service:       shared.Service(r.Header.Get(routes.ServiceHeader)),
		imageName:     r.Header.Get(routes.ImageNameHeader),
	}, http.StatusOK, nil
}

func (h *GetImageURLHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getImageURLArgs)

	emptyResponse := getImageURLResponse{}

	if args.service != shared.ECR {
		return emptyResponse, http.StatusBadRequest, errors.Newf("Container registry service %s is not supported.", args.service)
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	authConf, err := auth.ReadConfigFromSecret(context.Background(), args.integrationID, vaultObject)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to read container registry config from vault.")
	}

	conf, err := lib_utils.ParseECRConfig(authConf)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to parse configuration.")
	}

	err = container_registry.ValidateECRImage(conf, args.imageName)
	if err != nil {
		return emptyResponse, http.StatusUnprocessableEntity, err
	}

	return getImageURLResponse{
		Url: fmt.Sprintf("%s/%s", strings.TrimPrefix(conf.ProxyEndpoint, "https://"), args.imageName),
	}, http.StatusOK, nil
}

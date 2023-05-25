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
	"golang.org/x/oauth2/google"
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

	IntegrationRepo repos.Integration
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

	if !(args.service == shared.ECR || args.service == shared.GAR) {
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

	if args.service == shared.ECR {
		conf, err := lib_utils.ParseECRConfig(authConf)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to parse configuration.")
		}

		err = container_registry.ValidateECRImage(conf, args.imageName)
		if err != nil {
			return emptyResponse, http.StatusUnauthorized, err
		}

		return getImageURLResponse{
			Url: fmt.Sprintf("%s/%s", strings.TrimPrefix(conf.ProxyEndpoint, "https://"), args.imageName),
		}, http.StatusOK, nil
	} else if args.service == shared.GAR {
		conf, err := lib_utils.ParseGARConfig(authConf)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to parse configuration.")
		}

		// Obtain an OAuth2 token
		creds, err := google.CredentialsFromJSON(ctx, []byte(conf.ServiceAccountKey), "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get credential from service account key.")
		}
		token, err := creds.TokenSource.Token()
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get oauth token.")
		}

		// Create a new HTTP client
		client := &http.Client{}

		fullUrl := strings.Split(args.imageName, ":")[0]
		tag := strings.Split(args.imageName, ":")[1]
		host := strings.Split(fullUrl, "/")[0]
		projectID := strings.Split(fullUrl, "/")[1]
		repo := strings.Split(fullUrl, "/")[2]
		image := strings.Split(fullUrl, "/")[3]

		// Create a new HTTP request
		req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/v2/%s/%s/%s/manifests/%s", host, projectID, repo, image, tag), nil)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to create HTTP request.")
		}

		// Add the Authorization header to the request
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

		// Send the HTTP request
		resp, err := client.Do(req)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to send request.")
		}

		// Check the HTTP status code
		if resp.StatusCode == http.StatusOK {
			return getImageURLResponse{
				Url: args.imageName,
			}, http.StatusOK, nil
		} else if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return emptyResponse, resp.StatusCode, errors.New("Unable to access the requested image. Please double check you have the correct permissions and the image exists")
		} else {
			return emptyResponse, resp.StatusCode, errors.Newf("Received unexpected status:", resp.StatusCode)
		}
	} else {
		return emptyResponse, http.StatusBadRequest, errors.Newf("Container registry service %s is not supported.", args.service)
	}
}

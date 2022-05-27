package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
)

type ListBuiltinFunctionsHandler struct {
	GetHandler

	StorageConfig *shared.StorageConfig
}

type listBuiltinFunctionsResponse struct {
	Metadata []map[string]interface{} `json:"metadata"`
}

const builtinFunctionMetadataKey = "metadata"

func (*ListBuiltinFunctionsHandler) Name() string {
	return "ListBuiltinFunctions"
}

func (*ListBuiltinFunctionsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return nil, http.StatusOK, nil
}

func (h *ListBuiltinFunctionsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	serializedMetadata, err := storage.NewStorage(h.StorageConfig).Get(ctx, builtinFunctionMetadataKey)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error retrieving builtin functions")
	}

	var metadata []map[string]interface{}
	err = json.Unmarshal(serializedMetadata, &metadata)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error retrieving builtin functions")
	}

	return &listBuiltinFunctionsResponse{Metadata: metadata}, http.StatusOK, nil
}

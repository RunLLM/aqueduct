package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	// Names of the files inside a zipped Function
	modelFile       = "model.py"
	modelPickleFile = "model.pkl"
)

type exportFunctionArgs struct {
	*aq_context.AqContext
	operatorID uuid.UUID
	// Whether to export only the user-friendly function code
	userFriendly bool
}

type exportFunctionResponse struct {
	fileName string
	program  *bytes.Buffer
}

// Route: /function/{operatorId}/export
// Method: GET
// Params: operatorId
// Request
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: a zip file for the function content.
type ExportFunctionHandlerDeprecated struct {
	GetHandler

	Database database.Database

	DAGRepo      repos.DAG
	OperatorRepo repos.Operator
}

func (*ExportFunctionHandlerDeprecated) Name() string {
	return "ExportFunction"
}

func (*ExportFunctionHandlerDeprecated) Headers() []string {
	return []string{
		routes.ExportFnUserFriendlyHeader,
	}
}

func (h *ExportFunctionHandlerDeprecated) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Error when parsing common args.")
	}

	operatorIDStr := chi.URLParam(r, routes.OperatorIdUrlParam)
	operatorID, err := uuid.Parse(operatorIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Newf("Invalid function ID %s", operatorIDStr)
	}

	userFriendly := request.ParseExportUserFriendlyFromRequest(r)

	ok, err := h.OperatorRepo.ValidateOrg(
		r.Context(),
		operatorID,
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during operator ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this operator.")
	}

	return &exportFunctionArgs{
		AqContext:    aqContext,
		operatorID:   operatorID,
		userFriendly: userFriendly,
	}, http.StatusOK, nil
}

func (h *ExportFunctionHandlerDeprecated) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*exportFunctionArgs)

	emptyResp := exportFunctionResponse{}

	operatorObject, err := h.OperatorRepo.Get(ctx, args.operatorID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to get operator from the database.")
	}

	var path string

	if operatorObject.Spec.IsFunction() {
		path = operatorObject.Spec.Function().StoragePath
	} else if operatorObject.Spec.IsMetric() {
		path = operatorObject.Spec.Metric().Function.StoragePath
	} else if operatorObject.Spec.IsCheck() {
		path = operatorObject.Spec.Check().Function.StoragePath
	} else {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Requested operator is neither a function nor a validation.")
	}

	// Retrieve the workflow dag id to get the storage config information.
	dags, err := h.DAGRepo.GetByOperator(ctx, operatorObject.ID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while retrieving workflow dags from the database.")
	}

	if len(dags) == 0 {
		return emptyResp, http.StatusInternalServerError, errors.New("Could not find workflow that contains this operator.")
	}

	// Note: for now we assume all workflow dags have the same storage config.
	// This assumption will stay true until we allow users to configure custom storage config to store stuff.
	storageConfig := dags[0].StorageConfig
	for _, workflowDag := range dags {
		if workflowDag.StorageConfig != storageConfig {
			return emptyResp, http.StatusInternalServerError, errors.New("Workflow Dags have mismatching storage config.")
		}
	}

	program, err := storage.NewStorage(&storageConfig).Get(ctx, path)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to get function from storage")
	}

	if args.userFriendly {
		// Only the user-friendly code should be returned
		program, err = extractUserReadableCode(program, operatorObject.Name)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to export function code")
		}
	}

	return &exportFunctionResponse{
		fileName: operatorObject.Name,
		program:  bytes.NewBuffer(program),
	}, http.StatusOK, nil
}

func (*ExportFunctionHandlerDeprecated) SendResponse(w http.ResponseWriter, interfaceResp interface{}) {
	resp := interfaceResp.(*exportFunctionResponse)
	response.SendSmallFileResponse(w, resp.fileName, resp.program)
}

// extractUserReadableCode takes the zipped function code and only returns a zipped file
// containing the human-readable parts, i.e. source code, requirements.txt, and python_version.txt
// If no source code is found, it simply returns the original contents.
// In all cases, it renames the top-level directory to `operatorName`.
func extractUserReadableCode(data []byte, operatorName string) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	// Check if there is a source file, since older SDK clients did not generate this file
	hasSourceFile := false
	for _, zipFile := range zipReader.File {
		sourceFileName := fmt.Sprintf("%s.py", operatorName)
		parts := strings.Split(zipFile.Name, "/")
		if len(parts) == 2 && parts[1] == sourceFileName {
			hasSourceFile = true
			break
		}
	}

	buf := new(bytes.Buffer) // This is where the new zipped file is written to
	zipWriter := zip.NewWriter(buf)

	for _, zipFile := range zipReader.File {
		parts := strings.Split(zipFile.Name, "/")
		if hasSourceFile &&
			len(parts) == 2 &&
			(parts[1] == modelFile || parts[1] == modelPickleFile) {
			// There is a source file so we can skip the files that are not user-friendly to read
			continue
		}

		// Generate a new file name using `operatorName`, because it is more user-friendly to read
		// than the current directory name that is a unique UUID
		parts[0] = operatorName
		zipFileName := strings.Join(parts, "/")

		if err := writeZipFile(zipWriter, zipFile, zipFileName); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// writeZipFile writes the contents of `zf` to `w` in a file called `zfName`
func writeZipFile(w *zip.Writer, zf *zip.File, zfName string) error {
	f, err := zf.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// Read content of `zf`
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	newZipFile, err := w.Create(zfName)
	if err != nil {
		return err
	}

	// Copy `content` into `newZipFile`
	_, err = newZipFile.Write(content)
	return err
}

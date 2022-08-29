package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// Names of the files inside a zipped Function
	sourceFile      = "source.py"
	modelFile       = "model.py"
	modelPickleFile = "model.pkl"
)

type exportFunctionArgs struct {
	*aq_context.AqContext
	operatorId uuid.UUID
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
type ExportFunctionHandler struct {
	GetHandler

	Database          database.Database
	OperatorReader    operator.Reader
	WorkflowDagReader workflow_dag.Reader
}

func (*ExportFunctionHandler) Name() string {
	return "ExportFunction"
}

func (h *ExportFunctionHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Error when parsing common args.")
	}

	operatorIdStr := chi.URLParam(r, routes.OperatorIdUrlParam)
	operatorId, err := uuid.Parse(operatorIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Newf("Invalid function ID %s", operatorIdStr)
	}

	ok, err := h.OperatorReader.ValidateOperatorOwnership(
		r.Context(),
		aqContext.OrganizationId,
		operatorId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during operator ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this operator.")
	}

	return &exportFunctionArgs{
		AqContext:  aqContext,
		operatorId: operatorId,
	}, http.StatusOK, nil
}

func (h *ExportFunctionHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*exportFunctionArgs)

	emptyResp := exportFunctionResponse{}

	operatorObject, err := h.OperatorReader.GetOperator(ctx, args.operatorId, h.Database)
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
	workflowDags, err := h.WorkflowDagReader.GetWorkflowDagsByOperatorId(ctx, operatorObject.Id, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while retrieving workflow dags from the database.")
	}

	if len(workflowDags) == 0 {
		return emptyResp, http.StatusInternalServerError, errors.New("Could not find workflow that contains this operator.")
	}

	// Note: for now we assume all workflow dags have the same storage config.
	// This assumption will stay true until we allow users to configure custom storage config to store stuff.
	storageConfig := workflowDags[0].StorageConfig
	for _, workflowDag := range workflowDags {
		if workflowDag.StorageConfig != storageConfig {
			return emptyResp, http.StatusInternalServerError, errors.New("Workflow Dags have mismatching storage config.")
		}
	}

	program, err := storage.NewStorage(&storageConfig).Get(ctx, path)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to get function from storage")
	}

	_, err = extractUserReadableCode(program)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to export function code")
	}

	return &exportFunctionResponse{
		fileName: operatorObject.Name,
		program:  bytes.NewBuffer(program),
	}, http.StatusOK, nil
}

func (*ExportFunctionHandler) SendResponse(w http.ResponseWriter, interfaceResp interface{}) {
	resp := interfaceResp.(*exportFunctionResponse)
	response.SendSmallFileResponse(w, resp.fileName, resp.program)
}

// extractUserReadableCode takes the zipped function code and only returns a zipped file
// containing the human-readable parts, i.e. source.py, requirements.txt, and python_version.txt
// If no source code (i.e. source.py) is found, it simply returns the original zipped file.
func extractUserReadableCode(data []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	// Check if there is a source.py file, since older SDK clients did not generate this file
	hasSourceFile := false
	for _, zipFile := range zipReader.File {
		if zipFile.Name == sourceFile {
			hasSourceFile = true
			break
		}
	}

	if !hasSourceFile {
		// There is no source.py file so we just return the original zipped file without human readable code
		return data, nil
	}

	buf := new(bytes.Buffer) // This is where the new zipped file is written to
	zipWriter := zip.NewWriter(buf)

	for _, zipFile := range zipReader.File {
		logrus.Warnf("File Name: %v", zipFile.Name)
		if zipFile.Name == modelFile || zipFile.Name == modelPickleFile {
			// These files are not human readable so we skip them
			continue
		}

		if err := writeZipFile(zipWriter, zipFile); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// writeZipFile writes the contents of `zf` to `w`
func writeZipFile(w *zip.Writer, zf *zip.File) error {
	f, err := zf.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// Read content of `zf`
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	newZipFile, err := w.Create(zf.Name)
	if err != nil {
		return err
	}

	// Copy `content` into `newZipFile`
	_, err = newZipFile.Write(content)
	return err
}

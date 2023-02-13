package _000019_add_serialization_value_to_param_op

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	serialization_type_key  = "serialization_type"
	value_key               = "val"
	spec_field              = "spec"
	pythonExecutorPackage   = "aqueduct_executor"
	typeInferencePythonPath = "migrators.parameter_val_type_inference_000019.main"
)

type Operator struct {
	Id     uuid.UUID `db:"id" json:"id"`
	OpSpec Spec      `db:"spec" json:"spec"`
}

type MigrationSpec struct {
	ParamType string `json:"param_type"`
	ParamVal  string `json:"param_val"`
	Op        string `json:"op"`
}

func getAllOperators(
	ctx context.Context,
	db database.Database,
) ([]Operator, error) {
	query := "SELECT id, spec FROM operator;"

	var response []Operator
	err := db.Query(ctx, &response, query)
	return response, err
}

func updateParamOperatorWithNewSpec(
	ctx context.Context,
	operator Operator,
	db database.Database,
) error {
	// If the serialization_type is present we should consider this already migrated
	if _, exists := operator.OpSpec.Param[serialization_type_key]; exists {
		return nil
	}

	param_val := operator.OpSpec.Param[value_key]

	migrationSpec := MigrationSpec{
		ParamType: "",
		ParamVal:  param_val,
		Op:        "encode",
	}

	specData, err := json.Marshal(migrationSpec)
	if err != nil {
		return err
	}

	// Launch the Python job to infer the type of the parameter value
	cmd := exec.Command(
		"python3",
		"-m",
		fmt.Sprintf("%s.%s", pythonExecutorPackage, typeInferencePythonPath),
		"--spec",
		base64.StdEncoding.EncodeToString(specData),
	)
	cmd.Env = os.Environ()

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Errorf("Error running Python migration job. Stdout: %s, Stderr: %s.", outb.String(), errb.String())
		return err
	}

	outputs := strings.Split(outb.String(), "\n")
	param_type := outputs[0]
	param_val = outputs[1]
	operator.OpSpec.Param[serialization_type_key] = param_type

	// We also change the param value to be a base64 encoding
	operator.OpSpec.Param[value_key] = param_val

	newParamSpec := &Spec{
		Type:  operator.OpSpec.Type,
		Param: operator.OpSpec.Param,
	}

	changes := map[string]interface{}{
		spec_field: newParamSpec,
	}

	return repos.UpdateRecord(ctx, changes, "operator", "id", operator.Id, db)
}

func updateParamOperatorWithOldSpec(
	ctx context.Context,
	operator Operator,
	db database.Database,
) error {
	// If the serialization_type is not present we should not change this
	if _, exists := operator.OpSpec.Param[serialization_type_key]; !exists {
		return nil
	}

	// If we want to migrate back to old version we delete the serialization type in the spec
	// And we also decode the encoded value and store that
	param_val := operator.OpSpec.Param[value_key]
	serialization_type := operator.OpSpec.Param[serialization_type_key]

	migrationSpec := MigrationSpec{
		ParamType: serialization_type,
		ParamVal:  param_val,
		Op:        "decode",
	}

	specData, err := json.Marshal(migrationSpec)
	if err != nil {
		return err
	}

	// Launch the Python job to infer the type of the parameter value
	cmd := exec.Command(
		"python3",
		"-m",
		fmt.Sprintf("%s.%s", pythonExecutorPackage, typeInferencePythonPath),
		"--spec",
		base64.StdEncoding.EncodeToString(specData),
	)
	cmd.Env = os.Environ()

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Errorf("Error running Python migration job. Stdout: %s, Stderr: %s.", outb.String(), errb.String())
		return err
	}

	outputs := strings.Split(outb.String(), "\n")
	decoded_val := outputs[0]

	delete(operator.OpSpec.Param, serialization_type_key)
	operator.OpSpec.Param[value_key] = decoded_val

	newParamSpec := &Spec{
		Type:  operator.OpSpec.Type,
		Param: operator.OpSpec.Param,
	}

	changes := map[string]interface{}{
		spec_field: newParamSpec,
	}

	return repos.UpdateRecord(ctx, changes, "operator", "id", operator.Id, db)
}

package artifact

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact/boolean"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact/float"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact/jsonable"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact/table"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/dropbox/godropbox/errors"
)

// This file covers all artifact specs.
//
// To add a new spec:
// - Add a new enum constant for `Type`
// - Add a new field in specUnion for the new spec struct
// - Implement 3 additional methods for top level Spec type:
//  - IsNewType() method to validate the type
//  - NewType() method to get the value of the type from private `spec` field
//  - NewSpecFromNewType() method to construct a spec from the new type

type Type string

const (
	TableType Type = "table"
	FloatType Type = "float"
	BoolType  Type = "boolean"
	JsonType  Type = "json"
)

type specUnion struct {
	Type  Type         `json:"type"`
	Table *table.Table `json:"table,omitempty"`

	// TODO(ENG-1119): The float artifact currently also represents integers.
	Float *float.Float   `json:"float,omitempty"`
	Bool  *boolean.Bool  `json:"bool,omitempty"`
	Json  *jsonable.Json `json:"jsonable,omitempty"`
}

type Spec struct {
	spec specUnion
}

func NewSpecFromTable(t table.Table) *Spec {
	return &Spec{
		spec: specUnion{Type: TableType, Table: &t},
	}
}

func NewSpecFromFloat(f float.Float) *Spec {
	return &Spec{
		spec: specUnion{Type: FloatType, Float: &f},
	}
}

func NewSpecFromBool(b boolean.Bool) *Spec {
	return &Spec{
		spec: specUnion{Type: BoolType, Bool: &b},
	}
}

func (s Spec) Type() Type {
	return s.spec.Type
}

func (s Spec) IsTable() bool {
	return s.Type() == TableType
}

func (s Spec) Table() *table.Table {
	if !s.IsTable() {
		return nil
	}

	return s.spec.Table
}

func (s Spec) IsFloat() bool {
	return s.Type() == FloatType
}

func (s Spec) Float() *float.Float {
	if !s.IsFloat() {
		return nil
	}

	return s.spec.Float
}

func (s Spec) IsBool() bool {
	return s.Type() == BoolType
}

func (s Spec) Bool() *boolean.Bool {
	if !s.IsBool() {
		return nil
	}

	return s.spec.Bool
}

func (s Spec) IsJson() bool {
	return s.Type() == JsonType
}

func (s Spec) Json() *jsonable.Json {
	if !s.IsBool() {
		return nil
	}

	return s.spec.Json
}

func (s Spec) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.spec)
}

func (s *Spec) UnmarshalJSON(rawMessage []byte) error {
	var spec specUnion
	err := json.Unmarshal(rawMessage, &spec)
	if err != nil {
		return err
	}

	// Overwrite the spec type based on the data.
	var typeCount int
	if spec.Table != nil {
		spec.Type = TableType
		typeCount++
	} else if spec.Float != nil {
		spec.Type = FloatType
		typeCount++
	} else if spec.Bool != nil {
		spec.Type = BoolType
		typeCount++
	} else if spec.Json != nil {
		spec.Type = JsonType
		typeCount++
	}

	if typeCount != 1 {
		return errors.Newf("Artifact Spec can only be of one type. Number of types: %d", typeCount)
	}

	s.spec = spec
	return nil
}

func (s *Spec) Value() (driver.Value, error) {
	return utils.ValueJsonB(*s)
}

func (s *Spec) Scan(value interface{}) error {
	return utils.ScanJsonB(value, s)
}

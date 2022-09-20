package _000019_add_serialization_value_to_param_op

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

var param_type = "param"

func Up(ctx context.Context, db database.Database) error {
	resultOperators, err := getAllOperators(ctx, db)
	if err != nil {
		return err
	}

	for _, operator := range resultOperators {
		if operator.OpSpec.Type == param_type {
			err := updateParamOperatorWithNewSpec(ctx, operator, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	resultOperators, err := getAllOperators(ctx, db)
	if err != nil {
		return err
	}

	for _, operator := range resultOperators {
		if operator.OpSpec.Type == param_type {
			err := updateParamOperatorWithOldSpec(ctx, operator, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

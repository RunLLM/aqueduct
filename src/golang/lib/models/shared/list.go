package shared

import (
	"database/sql/driver"
	"sort"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

func ErrReadOnlyFieldType() error {
	return errors.New("We cannot serialize read-only field type.")
}

type NullableList[T any] []T

func (l *NullableList[T]) Value() (driver.Value, error) {
	return nil, ErrReadOnlyFieldType()
}

func (l *NullableList[T]) Scan(value interface{}) error {
	if value == nil {
		*l = nil // explicitly assign empty value.
		return nil
	}

	return utils.ScanJSONB(value, l)
}

type NullableIndexedList[T any] []T

type indexedListItem[T any] struct {
	Value T   `json:"value"`
	Idx   int `json:"idx"`
}

func (l *NullableIndexedList[T]) Value() (driver.Value, error) {
	return nil, ErrReadOnlyFieldType()
}

func (l *NullableIndexedList[T]) Scan(value interface{}) error {
	var itemList []indexedListItem[T]

	if value != nil {
		err := utils.ScanJSONB(value, &itemList)
		if err != nil {
			return err
		}
	}

	if len(itemList) > 1 {
		sort.SliceStable(itemList, func(i int, j int) bool {
			return itemList[i].Idx < itemList[j].Idx
		})
	}

	values := make([]T, 0, len(itemList))
	for _, item := range itemList {
		values = append(values, item.Value)
	}

	*l = values
	return nil
}

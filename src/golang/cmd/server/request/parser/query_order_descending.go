package parser

import (
	"net/http"
	"strings"

	"github.com/dropbox/godropbox/errors"
)

type OrderDescendingQueryParser struct{}

func (OrderDescendingQueryParser) Parse(r *http.Request) (bool, error) {
	query := r.URL.Query()

	var err error
	orderDescending := true
	if orderDescendingVal := query.Get("order_descending"); len(orderDescendingVal) > 0 {
		orderDescendingVal = strings.ToLower(orderDescendingVal)
		if orderDescendingVal == "true" {
			return true, nil
		}

		if orderDescendingVal == "false" {
			return false, nil
		}

		return true, errors.Wrap(err, "Invalid order_descending value.")
	}

	return orderDescending, nil
}

package parser

import (
	"net/http"

	"github.com/dropbox/godropbox/errors"
)

type OrderByQueryParser struct{}

func (OrderByQueryParser) Parse(r *http.Request, tableColumns []string) (string, error) {
	query := r.URL.Query()

	var err error
	var orderBy string
	if orderByVal := query.Get("order_by"); len(orderByVal) > 0 {
		// Check is a field in table
		isColumn := false
		for _, column := range tableColumns {
			if column == orderByVal {
				isColumn = true
				break
			}
		}
		if !isColumn {
			return "", errors.Wrap(err, "Invalid order_by value.")
		}
		orderBy = orderByVal
	}

	return orderBy, nil
}

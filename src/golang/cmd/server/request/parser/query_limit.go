package parser

import (
	"net/http"
	"strconv"

	"github.com/dropbox/godropbox/errors"
)

type LimitQueryParser struct{}

func (LimitQueryParser) Parse(r *http.Request) (int, error) {
	query := r.URL.Query()

	var err error
	limit := -1
	if limitVal := query.Get("limit"); len(limitVal) > 0 {
		limit, err = strconv.Atoi(limitVal)
		if err != nil {
			return -1, errors.Wrap(err, "Invalid limit parameter.")
		}
	}

	return limit, nil
}

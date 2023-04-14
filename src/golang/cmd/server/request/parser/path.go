package parser

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type pathParser struct {
	URLParam string
}

func (p pathParser) Parse(r *http.Request) string {
	return chi.URLParam(r, p.URLParam)
}

package database

import (
	log "github.com/sirupsen/logrus"
)

// Helper function to log the query being executed and any args that are specified.
func logQuery(query string, args ...interface{}) {
	if args != nil {
		log.Infof("Executing query: %s with args: %v", query, args)
	} else {
		log.Infof("Executing query: %s", query)
	}
}

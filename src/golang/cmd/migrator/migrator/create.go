package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/aqueducthq/aqueduct/lib/errors"
	log "github.com/sirupsen/logrus"
)

type ScriptLanguage string

const (
	SqlScriptLanguage ScriptLanguage = "sql"
	GoScriptLanguage  ScriptLanguage = "go"

	migrationFilePath = "golang/cmd/migrator/versions"
)

type templateArgs struct {
	Version int64
	Dir     string
	Path    string
}

// Create creates a new directory in lib/migration with name.
// It creates migration script templates based on the language type,
// which should be sql or go. It returns an error, if any.
func Create(name string, language ScriptLanguage) error {
	version, err := nextVersion()
	if err != nil {
		return errors.Wrap(err, "Unable to determine next schema version.")
	}

	dir, path, err := createMigrationDir(version, name)
	if err != nil {
		return errors.Wrap(err, "Unable to create new dir for schema migration.")
	}

	// Delete the new directory if there is any error in the rest of the function
	defer func() {
		if err != nil {
			os.RemoveAll(path)
		}
	}()

	args := &templateArgs{
		Version: version,
		Dir:     dir,
		Path:    path,
	}

	var tmpl *template.Template

	switch language {
	case SqlScriptLanguage:
		tmpl = template.Must(template.New("sql migration").Parse(getSqlTemplate()))
	case GoScriptLanguage:
		tmpl = template.Must(template.New("go migration").Parse(getGoTemplate()))
	default:
		return errors.Newf("Unsupported script language specified: %v", language)
	}

	if err := createTemplate(tmpl, args); err != nil {
		return errors.Wrap(err, "Unable to create migration script file.")
	}

	log.Infof("Successfully created %s migration script templates in: %s", language, args.Path)

	return nil
}

// nextVersion returns the next available schema version in the sequence.
// It does this by checking the migration dirs that exist in internal/migration.
func nextVersion() (int64, error) {
	files, err := os.ReadDir(migrationFilePath)
	if err != nil {
		return -1, err
	}

	maxVersion := int64(0)
	for _, file := range files {
		if !file.IsDir() {
			// This is not a migration script directory
			continue
		}

		name := file.Name()
		s := strings.Split(name, "_")

		if len(s) < 2 {
			// This is not a migration script directory, since there is no preceding version
			continue
		}

		versionStr := s[0]
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			return -1, err
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}

// createMigrationDir creates a new directory for a schema migration with
// the specified version number and name. It returns the name of this
// directory, the full path of the directory, and an error, if any.
func createMigrationDir(version int64, name string) (string, string, error) {
	dir := fmt.Sprintf("%06d_%s", version, name)
	path := filepath.Join(migrationFilePath, dir)
	err := os.Mkdir(path, 0o755)
	return dir, path, err
}

// createTemplate creates the main.go file using the template specified
// and the args provided. It returns an error, if any.
func createTemplate(tmpl *template.Template, args *templateArgs) error {
	path := filepath.Join(args.Path, "main.go")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, *args); err != nil {
		return err
	}

	return nil
}

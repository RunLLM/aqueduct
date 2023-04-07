.DEFAULT_GOAL = all

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

migrator:
	cd golang/cmd/migrator && go install
	mkdir -p build # Make the build directory if it doesn't exist
	mv $$GOPATH/bin/migrator build/migrator

executor:
	cd golang/cmd/executor && go install
	mkdir -p build
	cp $$GOPATH/bin/executor build/executor

server:
	cd golang/cmd/server && go install
	mkdir -p build
	cp $$GOPATH/bin/server build/server

migrator-release:
	cd golang/cmd/migrator && go install -ldflags="-s -w" -trimpath
	mkdir -p build # Make the build directory if it doesn't exist
	cp $$GOPATH/bin/migrator build/migrator

executor-release:
	cd golang/cmd/executor && go install -ldflags="-s -w" -trimpath
	mkdir -p build
	cp $$GOPATH/bin/executor build/executor

server-release:
	cd golang/cmd/server && go install -ldflags="-s -w" -trimpath
	mkdir -p build
	cp $$GOPATH/bin/server build/server

release: server-release executor-release migrator-release

# Prerequisites:
# install RTK codegen in ~/: `npm i @rtk-query/codegen-openapi`
# install pydantic codegen: `pip install pip3 install datamodel-code-generator`
# install swagger codegen (for go): `brew install swagger-codegen`
rest-api:
	rm -rf api/codegen/go
	rm -rf api/codegen/python
	rm -rf api/codegen/rtk
	mkdir api/codegen/go
	mkdir api/codegen/python
	mkdir api/codegen/rtk
	cd api && swagger-codegen generate -i aqueduct.json -l go-server -o codegen/go/
	cd api && swagger-codegen generate -i aqueduct.json -l python -o codegen/python/
# pydantic: cd api && datamodel-codegen --input aqueduct.json --input-file-type jsonschema --output codegen/python/response.py
	cd ui/common && npx @rtk-query/codegen-openapi ../../api/rtk.json

test:
	cd golang && go test -v $$(go list ./...)

# Database integration tests
test-database:
	cd golang && go test -v github.com/aqueducthq/aqueduct/lib/repos/tests -database

lint-go:
	gofumpt -w golang/
	cd golang && golangci-lint run --disable-all -p format --fix

# Remove any compiled binaries.
clean:
	rm -rf build

.PHONY: 
	server executor migrator \
	server-release executor-release migrator-release release \
	test test-database lint-go rest-api clean \

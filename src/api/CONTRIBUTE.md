# Aqueduct OpenAPI 3.0 Specification
## Code Generation
Go to `~/aqueduct/src` and run `make rest-api`. This command generates 3 directories under `~/aqueduct/src/api/codegen`:
* go: All go models and go server code
* python: All python models and python client code
* rtk: All typescript models and RTK client
* TODO: copy proper generated file to `src/golang`, `src/ui` and `sdk/`.
* TODO: copy proper generated file to `gitbook`

## Adding a Schema (Model)
* Each object should be with its own `.json` file under `schema/`.
* Add new objects to `"component"` -> `"schemas"` section of `aqueduct.json`. This ensures all objects are built and reused in typescript build.
* In `aqueduct.json`, use `#/component/schemas/<Name>` to refer to objects rather than directly referring to file.

## Adding a Path
* Follow existing pattern
* Add new parameters in `parameter/` and refer in `aqueduct.json` using file path.
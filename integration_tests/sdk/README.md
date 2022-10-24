# SDK Integration Tests

These tests run the SDK against an Aqueduct backend. Each test is built to clean up after itself. If it creates a workflow, it will attempt to delete it in the end. Tests can be run in parallel.

## Usage

Running all the tests in this repo:
`API_KEY=<your api key> SERVER_ADDRESS=<your server's address> INTEGRATION=aqueduct_demo pytest . -rP`

Running all the tests in a single file:
- `<your env variables> pytest <path to test file> -rP`

Running a specific test:
- `<your env variables>  pytest . -rP -k '<specific test name>'`

Running tests in parallel, with concurrency 5:
- Install pytest-xdist
- `<your env variables> pytest . -rP -n 5`

# Backend Integration Tests

These tests are run against the Aqueduct backend to check the endpoints' reads and outputs are as expected.

The `setup_class` sets up all the workflows which are read by each `test_endpoint_{handler_name}` test. When all the tests in the suite are done, the workflows set up in `setup_class` are deleted in the `teardown_class`.

## Usage

Running all the tests in this repo:
`API_KEY=<your api key> SERVER_ADDRESS=<your server's address> pytest . -rP`

Running all the tests in a single file:
- `<your env variables> pytest <path to test file> -rP`

Running a specific test:
- `<your env variables>  pytest . -rP -k '<specific test name>'`

Running tests in parallel, with concurrency 5:
- Install pytest-xdist
- `<your env variables> pytest . -rP -n 5`

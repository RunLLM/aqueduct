# SDK Integration Tests

These tests run the SDK against an Aqueduct backend. Each test is built to clean up after itself. If it creates a workflow, it will attempt to delete it in the end. All tests can be run in parallel.

There are two different suites of SDK Integration Tests, each with their own purpose:
1) Aqueduct Tests: These tests cover all Aqueduct behavior, from a user's perspective. They set up DAGs and usually publish flows. 
Test cases are generic enough to be reusable across multiple types of data integrations and engines. They are found in the `aqueduct_tests/` folder.
2) Data Integration Tests: While Aqueduct tests is the defacto test suite for testing Aqueduct behavior, one disadvantage of such
powerful but generic test cases is that every data integration is different. Unlike compute, each data integration has its own set of
APIs, abilities, and limitations, and Aqueduct Tests are philosophically less suitable for providing such coverage. Data Integration tests are meant
to be focused and complete, instead of reusable. They should only use the SDK's Integration API to validate data movement to and from
our supported third-party integrations.

## Configuration
For these test suites to run, a configuration file must exist at `test-config.yml`. This file contains:
1) The apikey to access the server.
2) The server's address.
3) The connection configuration information for each of the data integrations to run against. The test suites
will automatically run against each of the data integrations specified in this file, unless `--data` is supplied.

## Usage

From this directory, to run the Aqueduct Tests:
`pytest aqueduct_tests/ -rP -vv`

To run the Data Integration Tests:
`pytest data_integration_tests/ -rP -vv`

Both these test suites share a collection of configuration flags:
* `--data`: The integration name of the data integration to run all tests against.
* `--engine`: The integration of the engine to compute all tests on.
* `--keep-flows`: If set, we will not delete any flows created by the test run. This is useful for debugging.
* `--deprecated`: Runs against any deprecated API that still exists in the SDK. Such code paths should be eventually deleted after some time, but this ensures backwards compatibility.

For additional markers/fixtures/flags, please inspect `conftest.py` in this directory. For test-specific configurations,
see `aqueduct_tests/conftest.py` and  `data_integration_tests/conftest.py`.

## Useful Pytest Flags 

Running all the tests in a single file:
- `pytest <path to test file> -rP -vv`

Running a specific test:
- `pytest <specific test directory> -rP -vv -k '<specific test name>'`

Running tests in parallel, with concurrency 5:
- Install pytest-xdist
- `pytest <specific test directory> -rP -vv -n 5`

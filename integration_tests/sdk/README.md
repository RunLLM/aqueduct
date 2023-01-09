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

## Usage

From this directory, to run the Aqueduct Tests:
`API_KEY=<your api key> SERVER_ADDRESS=<your server's address> INTEGRATION=aqueduct_demo pytest aqueduct_tests/ -rP -vv

To run the Data Integration Tests:
`API_KEY=<your api key> SERVER_ADDRESS=<your server's address> INTEGRATION=aqueduct_demo pytest data_integration_tests/ -rP -vv

The test suite also has a variety of other custom flags, please inspect the `conftest.py` files in both this directory and subdirectories
to find their descriptions.

Running all the tests in a single file:
- `<your env variables> pytest <path to test file> -rP -vv`

Running a specific test:
- `<your env variables>  pytest <specific test directory> -rP -vv -k '<specific test name>'`

Running tests in parallel, with concurrency 5:
- Install pytest-xdist
- `<your env variables> pytest <specific test directory> -rP -vv -n 5`

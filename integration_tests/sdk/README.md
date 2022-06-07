# SDK Integration Tests

The repo contains integration test cases that use our Python SDK against a cluster. Example models used are the
[sentiment model](https://github.com/spiralai/example-notebooks/tree/master/sentiment/vader) and the [churn model](https://github.com/spiralai/example-notebooks/tree/master/churn/churn_predictor).

Each test is built to clean up after itself. If it creates a workflow, it will attempt to delete it in the end.

## Usage

Running all the tests in this repo:
`EMAIL=<EMAIL> API_KEY=<API_KEY> GATEWAY_ADDRESS=<GATEWAY_ELB_ADDRESS> INTEGRATION=<CONNECTED_INTEGRATION_NAME> pytest . -rP`
Note that integration names can be found in the `integration` table of your cluster's Postgres instance.

Running all the tests in a single file:
`<...> pytest <path to test file> -s`

Running a specific test:
`<...>  pytest . -rP -k '<specific test name>'`

There are two complexity flags to toggle for the suite:
`--complex_models`: if set, we will always run real models like sentiment and churn. Otherwise, tests will default instead
to a dummy function, which is significantly faster to evaluate.
`--publish`: if set, our flow tests will run flow.publish() instead flow.test(). Again, leaving this flag out will speed
up test execution, at the expense of losing `publish()` coverage.

You can always run tests in parallel by installing pytest-xdist` and adding `-n <num_workers>` to your command.

Note that integration tests will default to using http. So test against a https cluster, use the `--https` flag. Eg.
`<...> pytest . -rP --https`
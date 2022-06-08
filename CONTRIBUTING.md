# Welcome to the Aqueduct contributing guide <!-- omit in toc -->

Thank you for becoming an Aqueduct contributor! Here you will find guides for setting up your development
environment and contributing to our codebase.

## Setting up development environment
To contribute to our UI, you need to install NodeJS with version greater than `v16`. See the
instruction [here](https://nodejs.org/en/download/) to download and install NodeJS.

To contribute to our backend server, you need to install Go with version greater than `v1.16`.
See the instruction [here](https://go.dev/dl/) to download and install Go. Note that our `go.mod`
uses Go version `1.16`, so if you are using a later version of Go, please ensure that your contribution
is backward compatible with `1.16`.

You also need Python `>=3.7` to contribute to our SDK and the backend Python connector package. See
the instruction [here](https://www.python.org/downloads/) to download and install Python.

## Contributing to the codebase

### Create a GitHub issue
Before writing code, please create a GitHub issue so that our team and external contributors are aware
that you will be working on this task. Follow the steps below:
1. Browse existing issues and verify there aren't issues that match your task.
2. Create a new issue following our issue template that describes the task.
3. Assign the issue to yourself so that we know you are activaly working on this issue.

### Development flow
Please follow the steps below to contribute:
1. Fork our GitHub repository.
2. Open a branch for your work.
3. Write code that implements the GitHub issue you created, or an issue that doesn't have an assignee.
In the latter case, please update the GitHub issue and assign yourself to the issue.
4. Make sure your code passes all tests (see the Testing section below for more details).
5. Rebase `main` with your branch before submitting a pull request following our PR template.
6. Open the pull request against our repo.

### Testing
Please write unit tests and integration tests for new feature contributions.

Note that before running any test, you need to make sure the test will run against your latest version
of the code. To ensure this, from the aqueduct root directory, run `python3 scripts/install_local.py`.

Then, you can go ahead and verify all the tests are passing:
1. From `./src`, run `make test` to ensure all Golang unit tests are passing.
2. If the PR made changes to the database APIs, from `./src`, run `make test-database` to run the
database integration tests.
3. Run `pip3 install pytest` and run `pytest . -rP -vv` from `./sdk` to ensure all SDK unit tests are passing.
4. See [here](https://github.com/aqueducthq/aqueduct/tree/main/integration_tests/sdk) for running the SDK integration tests.

### Review process
After you make the PR, a member of our team will review, suggest changes, and approve it.
We verify the following when reviewing the code:
1. The PR addresses the GitHub issue raised.
2. The functionality is correct.
3. If the PR includes new features, it contains the relevant tests.
4. The PR passes all the tests.
5. The code conforms to the style guide and fits well with the existing code.
6. The PR does not introduce security vulnerability.

Once all of the above are verified, we will approve and merge your contribution!
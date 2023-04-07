# Contributing to Aqueduct

Thank you for contributing to Aqueduct! This is a quick guide to setting up
your development environment and contributing to the codebase.

We welcome contributions from anyone who's excited about building prediction 
infrastructure for data scientists. Contributions can be both in the form of
actual code or in the form of bug reports and feature requests.

For general community guidelines, please see our [code of
conduct](CODE_OF_CONDUCT.md).

## Setting up your development environment

### UI

The Aqueduct UI is built with React and Typescript. To run the UI in 
development mode, you will need to have NodeJS installed. You can find the
version of NodeJS we use in `src/ui/.nvmrc`. NodeJS installations can be
unpredictable, so your development setup *might* work with a different version
of NodeJS, but for the most predictable results, please use the same version of
NodeJS. 

You can find installation instructions for NodeJS [here](https://nodejs.org/en/download/).

### Backend

Our backend server is primarily implemented in [Golang](https://go.dev). You can find the
version of Golang that we're currently using at the top of `src/golang/go.mod`.
You can find installation instructions for Golang [here](https://go.dev./dl).

Our backend also includes Python code to orchestrate operators and connect to
common data systems. We currently support Python versions 3.7 through 3.10.

## Contributing to the codebase

If you're fixing a simple bug, please feel free to submit a pull request
directly.  However, if you find yourself implementing a larger change or
working on a new feature, we'd love to hear from you before you dive into the
code -- this helps us know what's going on in the community and also gives
the project maintainers an opportunity to weigh in before you're off to the
races. You can create a new issue [here](https://github.com/aqueducthq/aqueduct/issues/new/choose).

### Development flow

Before you get started, fork the Aqueduct GitHub repo. We typically follow the
[GitHub flow](https://docs.github.com/en/get-started/quickstart/github-flow)
for development purposes. If you're not familiar with the GitHub flow, here's a
quick overview:

1. On your fork of Aqueduct, create a new branch for your work. 
2. Make whatever changes you're intending to make. 
3. Verify the correctness of your changes (and write tests!) -- see below for
   more on this.
4. Merge the latest commits from `main` branch on aqueducthq/aqueduct (or
   rebase your branch against `main`). 
5. Open a new pull request to merge your changes into `main`.
6. Run the integration tests by adding the `run_integration_test` label to your PR.

### Verifying your changes

There are two ways that you should verify the correctness of your changes:
running the application itself and testing.

#### Running the application

Once you've made changes in your local clone of the repository and are ready to
test it, the easiest way to test is to replace your local Aqueduct 
installation, which you can do by running `python3 scripts/install_local.py`.
This will compile your local changes both on the backend server and UI and
install them in `~/.aqueduct`. If you aren't making UI changes, then running
`python3 scripts/install_local.py -s -g -e` will save some time by only updating
the non-UI parts of the system.

Once this is done, you can run `aqueduct start` to access the app running with
the changes you've made.

We recommend creating some of the workflows in `sdk/examples` as a good testing
starting point.

#### Testing

Please write unit tests and integration tests for new feature contributions.

You can run the existing test suite with the following steps -- please make
sure to run `python3 scripts/install_local.py` first:

1. From `src`, run `make test` to ensure all Golang unit tests are passing.
2. If the PR made changes to the database APIs, run `make test-database` in `src` to run the
database integration tests.
3. Run `pip3 install pytest` and run `pytest aqueduct_tests/ -rP -vv` and `pytest data_integration_tests/ -rP -vv` from the `sdk` directory to ensure all SDK unit tests are passing.
4. See [here](https://github.com/aqueducthq/aqueduct/tree/main/integration_tests/sdk) for instructions on running the SDK integration tests.

### Reviewing code changes

We're constantly working to improve our CI process. Today, when you make a PR,
our GitHub Actions will ensure that your code changes pass basic testing
requirements as well as linting requirements. 

We will publish a code quality guide here soon, but as a part of the review
process, we will typically look for general code style & quality, completeness of the
feature or bugfix, security vulnerabilities, and tests.

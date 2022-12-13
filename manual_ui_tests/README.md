# Manual UI Tests
These scripts helps setting up a number of workflows to verify their rendered outcomes in UI.
Simply run from your root `aqueduct` directory with `python3 manual_ui_tests/initialize.py`. You can go to **Workflows** page and see if each workflow matches its description.

* View **Usage** for additional parameters
* Run through **Checklist** when you are doing a more rigorous checking. For example, during release.
    * The **Checklist** covers important UI features not necessarily related to deployed workflows. Like integrations.
* View **Keep It Up to Date** for contributing principles and details.

## Usage
To run with more flexibility, configure the following commandline flags:
* `--addr`: the server address
* `--integration`: the integration name. The integration need to be pre-configured before running the script.
* `--api-key`: the API key if different from `aqueduct.api_key()` by any reason.
* `-q`, `--quiet`: block commandline outputs.

## Checklist
* **Workflows** Page: There should be **4** workflows. **3** Succeeded and **1** Failed.
* **Workflow Details** Page: Each page should reflect the **workflow description**. Pay extra attention to **Workflow Status Bar** and any noted **sidesheets** in the description.
* **Integration** Page:
    * There should be **11** *Data* integrations and **4** *Compute* integrations.
    * If you are not using additional integration, `aqueduct_demo` should be the only available one.
* **Integration Details** Page:
    * There should be **4** workflows in **Workflows** section.
    * If you are using `aqueduct_demo`, there should be **8** tables in **Data** section.
* **Data** Page: There should be **1** data available.

## Keep It Up to Date
* The scripts and **Checklist** should be focused on UI features. We should use other tests to cover SDK / Backend specific features.
* Steps to add a workflow:
    * Name with one of `succeed_`, `warning_` and `fail_`.
    * Create a file with `<name>.py` under `workflows` directory. With the following:
        * `NAME` constant.
        * `DESCRIPTION` constant.
        * `deploy(client, integration)` function.
    * Update `initalize.py` by importing the new file and update `WORKFLOW_PKGS` constant.
    * Update **Workflows** and maybe **Data** section in **Checklist**.
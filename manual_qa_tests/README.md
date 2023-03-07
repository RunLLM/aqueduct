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
* `--data-integration`: the data integration name. The data integration need to be pre-configured before running the script.
* `--api-key`: the API key if different from `aqueduct.api_key()` by any reason.
* `--example-notebooks`: also run all example notebooks
* `--slack-token`: The Slack App bot token to integrate with slack notifications.
* `--slack-channel`: The channel to send Slack notifications.
* `--notification-level`: The notification threshold level. (e.g. 'Success' to receive all notifications, 'Warning' to recieve for warning / failed workflows, and 'Error' for failed workflows only.)

## Checklist
* **Workflows** Page: 
    * There should be **5** workflows. **4** Succeeded and **1** Failed if using `--example-notebooks`
    * There should be **5** workflows. **3** Succeeded and **2** Failed if **not** using `--example-notebooks`
* **Notifications**: There should be **2** notifications for failed workflow.
* **Workflow Details** Page: Each page should reflect the **workflow description**. Pay attention to any noted **sidesheets** behaviors in the description.
* **Integration** Page:
    * There should be **11** *Data* integrations, **5** *Compute* integrations, and **2** *Notifications* integrations.
    * If you are not using additional integration, `aqueduct_demo` should be the only available one.
* **Integration Details** Page:
    * In the **Workflows** section of the `aqueduct_demo` page:
        * There should be **10** workflows if using `--example-notebooks`
        * Ther should be **4** workflows if **not** using `--example-notebooks`
    * If you are using `aqueduct_demo`, there should be **8** tables in **Data** section.
* **Data** Page: There should be **5** data rows available.
* **Slack channel**:
    * There should be **16** new notifications.
    * Each notification should have the following aspects:
        * A title including the workflow's name and status
        * Workflow name
        * ID
        * Result ID
        * If the workflow has check failures, it should list all failed checks with correct error or warning state.
        * A link to the workflow result's UI page.

## Keep It Up to Date
* The scripts and **Checklist** should be focused on features that:
    * Requires E2E workflow deployments.
    * UI.
    * Any other human setup / evaluation.
* You should consider using other more automated tests to cover your need:
    * SDK integration tests.
    * Backend unittests.
* Steps to add a workflow:
    * Name the workflow with one of `succeed_`, `warning_` and `fail_` prefixes.
    * Create a file with `<name>.py` under `workflows` directory. With the following:
        * `NAME` constant.
        * `DESCRIPTION` constant.
        * `deploy(client, integration)` function.
    * Update `initalize.py` by importing the new file and update `WORKFLOW_PKGS` constant.
    * Update **Workflows**, **Notification** and maybe **Data** section in **Checklist**.
* Since workflow stop executing at failure, make sure fail workflow fails deterministically such that test outcome is predictable:
    * No operator that could execute in parallel to the failure one.
# Changelog

## 0.2.7
Released on March 22, 2023.

### Key Features
* [Beta] Aqueduct now has support for on-demand Kubernetes cluster creation and
    management on AWS. From the Aqueduct UI, you can connect Aqueduct to your
    AWS account via the cloud integration feature. Once connected, you can use
    this cloud integration to ask Aqueduct to automatically create a Kubernetes
    cluster for you. See the documentation
    [here](https://docs.aqueducthq.com/integrations/on-demand-resources/on-demand-aws-eks-clusters)
    for how to create an operator that uses on-demand Kubernetes.

### Enhancements
* Improves error handling to return more detailed error messages from errors
    occuring during execution.
* Improves error handling by surfacing errors that occur outside of the
    execution of an indvidual function as workflow-level errors; these errors
    could occur for example if a compute system was misconfigured.
* Improves handling of artifact name conflicts in the Python SDK. Explicitly
    named artifacts (using either the `outputs` argument to the `@op` decorator
    or the `.set_name()` function) will immediately flag and prevent conflicts
    in artifact names. Automatically named artifacts will error if multiple
    artifacts with the same name are included in a single `publish_flow` call.
* Displays all compute engines associated with a workflow on the workflows list
    page.
* Improves efficiency when previewing large objects on the UI by retrieving a
    sample of the data instead of the full data object and noting that the
    displayed data is a sample.

### Bugfixes
* Fixes bug where an S3 or GCS bucket being used as the Aqueduct artifact store
    could possibly be deleted from the UI.
* Fixes bug that caused navigation buttons to be misaligned with other buttons
    on the action bar on the workflow details page.
* Fixes bug where navigating to the next most recent run on the workflow
    details page would not work correctly.
* Fixes bug pending or errored metric would show as "Unknown" on the UI instead
    of with the correct status.
* Fixes bug where warning-level checks were being shown as failures on the
    workflows list page.
* Fixes bug where certain DAG layouts would continue to show a layout with
    overlapping and crossing edges.

### Note
* The parameterization of SQL queries may have unexpected behavior if you accidentally define a 
    parameter with the same name twice. The parameter value will be chosen at random in such a case.
    This bug will be fixed in the next release.

## 0.2.6

Released on March 14, 2023.

### Key Features
* Enables registering a workflow without immediately triggering a run of that
    workflow. When calling `publish_flow` the `run_now` parameter can be set to
    `False`, which will tell the Aqueduct server to wait until the next
    scheduled (or triggered) run. In the interim, only the DAG's structure will
    be shown without any execution metadata.

### Enhancements
* Adds error checking to ensure that a remote compute engine is paired with a
    remote artifact store.
* Adds error checking to ensure that a valid `kubeconfig` is provided when
    connecting Aqueduct to Kubernetes.
* Enables using `~` to refer to the home directory when specifying a path to
    AWS credentials or a `kubeconfig`.
* Adds icons to integration details views to indicate when an object store (AWS
    S3, GCS) are being used for metadata storage.
* Allows specifying the specific library version of CUDA when requesting GPU
    resources; v11.4.1 is the current default because that is what EKS clusters
    use by default.

### Bugfixes
* Fixes bug where certain errors occurring during task launch weren't clearly 
    surfaced in the Aqueduct stack.
* Fixes bug where, within a Python process, executing an operator in lazy mode
    precluded users from later executing it in eager mode (or vice versa).
* Fixes bug where notification count wasn't being shown.
* Fixes bug where the notice for an ongoing metadata migration would run off
    the end of the screen.
* Removes duplicate "History" header on metric details page.

## 0.2.5
Released on March 7, 2023.

### Key Features
* Users can now run Aqueduct workflows on Spark clusters on AWS EMR. With Apache Livy as an interface, Aqueduct can submit your code to your Spark cluster reliably and seamlessly. See our documentation [here](https://docs.aqueducthq.com/integrations/adding-an-integration/connecting-to-spark-emr.md).
* Redesigns node layout on DAG view to improve information presentation and
    better distinguish between different node types.

### Enhancements
* Enables specification of Snowflake role when connecting the Snowflake
    integration.
* Updates workflow details page header to be more compact and reduce
    information overload.
* Adds support for specifying compute engine and resource requirements when
    creating metrics & checks.

### Bugfixes
* Fixes issue where size of large rows on Snowflake was artificially limited.
* Resolves requirement mismatches that would occur on the latest versions of
    Ubuntu 22.
* Fixes layout issue where dates on metric history graph could have been cut
    off.
* Fixes bug where the integration details page for compute integrations would
    not list all workflows using that integration.
* Fixes bug where metrics plot failed to render when upstream operator was 
    cancelled.

## 0.2.4
Released on February 28, 2023.

### Enhancements
* Opens links to docs and feedback in new tabs rather than in the existing tab
    on the Aqueduct UI.
* When authoring a pipeline, allows reusing the same Python function multiple
    times in the same DAG.
* Improves the layout of the card displaying metadata storage information on
    the settings page.

### Bugfixes
* Fixes bug where changing the Aqueduct metadata storage layer when there was a
    previously-failed workflow would cause the data migration process to pause.
* Fixes bug where certain DAGs would render in a confusing fashion on the
    workflow details page. The algorithm for DAG layouts is now signifcantly
    more reliable.

## 0.2.3
Released on February 22, 2023.

### Enhancements
* Updates workflow and data table views to show overview of all executed checks
    rather than just one.
* Garbages collect Lambda-specific Docker images from the Aqueduct server's
    machine after the Lambda integration connection is finished.
* Improves performance of the Aqueduct serialization library by looking into
    collection types (lists, tuples) and using data-type specific serialization
    methods for each entry.
* On the Aqueduct settings page, adds details about what storage engine is
    being used for metadata and version snapshot storage.

### Bugfixes
* Fixes detail header alignment on artifact and operator details pages.
* Fixes bug where latest MariaDB and MySQL drivers were not bieng installed
    correctly on M1 Macs.
* Fixes bug where running the same function with multiple unnamed parameters
    more than once would fail.
* Fixes bug where Aqueduct Docker images running for save operators were missing
    dependencies for certain data types.

## 0.2.2
Released on February 14, 2023.

### Key Features
* Adds support for receiving Aqueduct notifications via email or in Slack
    workspaces. You can configure notification settings for your Aqueduct
    installation at large, and you can also customize notification settings
    per-workflow. Notifications can be configured to be sent for all workflow
    executions, only on warnings, or only on errors.
    * *Email*: You can connect Aqueduct to your email account and specify a
        list of email addresses as recipients. Each notification will trigger a
        separate email.
    * *Slack*: You can connect Aqueduct to your Slack workspace and specify a
        channel that Aqueduct should send notifications on. Each notification
        will send a separate message.

### Enhancements
* Adds support for specifying Snowflake schema when creating integration from
    UI.
* Adds support for executing an operator that has one or more parameters and 
    multiple outputs interactively. You can call the same function, and
    Aqueduct will automatically override previous implicitly created
    parameters. See [our
    documentation](https://docs.aqueducthq.com/parameters) for more details.
    ```python
    @op
    def fn(param):
      return param

    res = fn(1).get()
    >>> 1 # Creates a parameter named `param` for you automatically, with a default value of 1.

    res = fn(2).get()
    >>> 2 # Updates `param` to have a default value of 2.
    ```

### Bugfixes
* Fixes two bugs where Aqueduct server was retrieving full data objects from
    the Aqueduct metadata store to check for their existence. When working 
    with non-trivial data, this could cause serious performance issues.
* Fixes bug where object does not exist errors from S3 were mishandled, causing
    Aqueduct to surface incorrect errors.
* Fixes bug where pods that are marked as pending on Kubernetes were being
    treated as failed operators.
* Fixes bug where log and stack traces blocks didn't have proper formatting and
    backgrounds on the UI.
* Fixes bug that was causing full data objects to be retrieved repeatedly when
    loading metadata on the UI.
* Fixes bug where UI was previously treating not-yet-executed operators (for an
    in-progress workflow) as failed operators.
* Fixes bug where the SDK's `global_config` could not be changed to set
    Aqueduct as the compute engine.

## 0.2.1
Released on February 7, 2023.

### Key Features
* Allows customizing artifact names from the SDK in one of two ways.
    ```python
    # Method 1: Use the decorator
    @op(outputs=['sklearn model', 'churn predictions'])
    def train_and_predict_churn(features):
      # ...
      return model, predictions

    # Method 2: Use .set_name()
    @op
    def train_model(features):
      # ...
      return model

    # ...
    model = train_model(features)
    model.set_name('churn model')
    ```

### Enhancements
* Allows providing filepath to ServiceAccount key file when connecting to
    BigQuery from Aqueduct SDK.
* Improves form validation when connecting Databricks integration.
* Throughout the SDK, enables references to workflows using workflow name in
    addition to workflow ID.
* Puts upper bounds on Python package dependencies to prevent unexpected
    regressions (e.g., recent issues caused by SQLAlchemy 2.0).

### Bugfixes
* Fixes bug where errors were not being properly handled when an operator had
    multiple outputs. This was occurring because the return value didn't have
    the expected length.

## 0.2.0
Released on January 31, 2023

### Key Features
* [Beta] Aqueduct now supports running workflows on Databricks Spark clusters! 
    As of this release, you can now connect Aqueduct to a Databricks cluster 
    from the UI and use the Aqueduct decorator API to deploy workflows onto 
    those clusters. 
    * Databricks workflows can read data from Snowflake and AWS S3. Future
        releases support other data systems, including Delta Lake.
    * Currently, you cannot run a subset of a workflow on a Databricks cluster;
        the whole workflow must be run on Databricks.
    * We plan to add support for non-Databricks Spark clusters in the coming
        releases.

### Enhancements
* Allows workflows running on Airflow to be triggered upon the completion of
    other workflows. Note that the completion of an Airflow workflow cannot
    trigger the execution of another workflow because completion state is not
    synchronously tracked on Airflow. 
* Unifies color and size of status indicators throughout the UI.

### Bugfixes
* Fixes bug where internal server error was uncaught when retrieving operator
    results. 
* Fixes bug where workflow status bar had unnecessary backticks around objects.
* Fixes bug where access checks for AWS S3 buckets would fail with certain
    permissions that were in fact valid.
* Fixes bug where saving tables to relational databases with long column names
    (\> 255 characters) would fail.
* Fixes bug where SQLAlchemy version 2 introduced access issues with Pandas
    DataFrames. Our current solution is to require SQLAlchemy version 1.
* Fixes bug where listing tables in BigQuery required complex, brittle SQL 
    queries.
* Fixes bug where data listing page might crash on UI after the execution of a
    failed workflow.
* Fixes bug where status indicator on check and metric details was not being 
    properly displayed.
* Fixes bug where checks and metrics of failed workflow executions show no values.
* Fixes bug where after switching to cloud storage as the metadata store, new integration
    credentials weren't properly saved to cloud storage.
* Fixes bug where preview fails after switching to cloud storage as the metadata store.
* Fixes bug where failing metrics show as NaN in metric preview list on UI.

### Deprecations
* The `.save()` on Artifacts has been removed. As of
    [v0.1.6](https://github.com/aqueducthq/aqueduct/releases/tag/v0.1.6), the
    recommended method is to use the `.save()` API on integration objects.

## 0.1.11
Released on January 23, 2023

### Enhancements
* Upgrades workflow layout rendering tool to use the elkjs library.
* Shows the name of the options on the UI's menu sidesheet to improve clarity.
* Removes the Aqueduct logo on the UI's home page to reduce redundancy.

### Bugfixes
* Fixes bug where operator execution fails when running on Kubernetes. This was due to a time gap
    between launching a Kubernetes job and spinning up a pod, and our system wasn't accounting
    for this.
* Fixes bug where the workflow details page keeps re-rendering.
* Fixes bug where the Kubernetes logo doesn't show up on the UI.
* Fixes bug where the UI keeps hitting the notification route, which led to unnecessary overhead.
    This was caused by omitting an empty dependency array in one of our useEffect hooks.

## 0.1.10
Released on January 17, 2023

### Enhancements
* For workflows that are triggered at the end of other workflows, we now allow changing the
    triggering workflow from the UI's workflow settings dialog.
* Differentiate keys and values better on the UI; adds the use of different
    colored text to make it clear which is the key and which is the value when
    showing, for example, metric and check values.
* Improves presentation of non-success states of metrics and checks on workflow
    DAG. Rather than leaving the nodes empty as before, they now include icons
    that demonstrate the execution state (failed, pending, canceled).
* Adds redesigned search interface to workflow and data list pages. The search
    bar itself has been reduced in size, and a sort functionality has been
    added that allows users to select a column by which to sort the view.
* Orders integrations alphabetically on the integrations page to make them
    easier to find.

### Bugfixes
* Fixes bug where operator & artifact statuses were missing from details pages.
* Fixes bug where the header breadcrumbs did not show the title of the workflow
    on the metric details page.

## 0.1.9
Released on Januay 10, 2023.

### Key Features
* As of this release, Aqueduct has usage tracking. Usage tracking is fully
    anonymized and captures API routes, performance data, and error rates
    without revealing any specifics of your machine or deployment. For more
    details, check out our [documentation](https://docs.aqueducthq.com/usage).
* We now support cascading workflow triggers, which means a workflow can trigger another one at the end of its execution. You can specify that in our [python SDK](https://github.com/aqueducthq/aqueduct/blob/main/sdk/aqueduct/client.py#L373).

### Enhancements
* Makes the artifact, check, metric, and operator details pages full width.
* Shows the Aqueduct version number on the UI navigation bar.
* Hides previews when artifacts are canceled.
* Hides parameters in status bar.

### Bugfixes
* Fixes a number of UI bugs:
  * Resets workflow settings dialog content after close.
  * Aligns margins on right side of workflow details page.
  * Removes vestigial popover to access settings page.
  * Addresses regression where a workflow's saved objects were not being shown
      prior to workflow deletion.
  * Aligns the width of metric and check history items.
  * Updates the metadata views (workflows & data list pages) to differentiate 
      table headers from metadata rows.
  * Persists the number of rows shown per-page on metadata views between page refreshes.

## 0.1.8
Released on December 20, 2022.

### Enhancements
* Allows user to set compute engine in the operator's decorator.
* Reduces the number of significant figures for metrics on data list page and workflows list page to
    improve readability.

### Bugfixes
* Fixes a bug where the UI shows data section for compute integrations.
* Fixes a bug where previewing Mongo collection crashes.

## 0.1.7
Released on December 14, 2022.

### Bugfixes
* Fixes a bug where the Aqueduct installation script fails if the user doesn't have conda installed.

## 0.1.6
Released on December 13, 2022.

### Key Features
* Introduces new table views on the workflows and data pages that show rich
    metadata at a glance, including workflow and artifact status, data types,
    and associated metrics and checks!
* Adds support for integrating with conda. Once the user registers conda integration through the UI, 
    Aqueduct will create conda environments to run any newly created workflows to provide better 
    Python version and dependency management.

### Enhancements
* Introduces a new `save` API; now, to save an artifact, users can write the
    following. The original `table.save()` syntax still works but will be
    deprecated in a future release.
```python
db.save(table, 'my_new_table', update_mode='replace')
```
* Disallows creating multiple integrations with the same name.

### Bugfixes
* Fixes a bug where unused integration couldn't be deleted if historical
    workflow runs were associated with it.
* Fixes a bug where logs weren't being displayed on operator details page.
* Fixes a bug where saving multiple pieces of data to the same database would
    cause the workflow UI to crash.
* Fixes a bug where calling a metric or check with no inputs didn't raise a
    client-side error.
* Fixes a bug where metric history & graph was not sorted by time.
* Fixes a bug where where every click into a workflow DAG node reset the DAG
    visualization.
* Fixes a number of bugs that caused no notifications to be displayed on the
    UI.

## 0.1.5
Released on November 29, 2022.

### Key Features
* Enables operators running on Kubernetes to access GPUs and set RAM and CPU
    requirements. Note that using a GPU requires your Kubernetes cluster to
    already have GPU machines attached. See [our
    documentation](https://docs.aqueducthq.com/operators/configuring-resource-constraints) for more details.
```python
@op(resources={'num_cpus': 2, 'memory': '5Gb', 'gpu_resource_name': 'nvidia.com/gpu'})
def my_operator_with_many_resources():
  return 1
```
* Similarly, functions running on AWS Lambda can have memory requirement
    set using the syntax above; AWS Lambda does not support setting CPU requirement
    and it does not support GPUs.

### Enhancements
* Enables operator previews to execute using different integrations, including
    using the resource constraints described above.
* Allows for the execution engine to be set globally for a client instance. See
    more details
    [here](https://docs.aqueducthq.com/integrations/using-integrations/compute-integrations#setting-the-global-configuration):
```python
aq.global_config({'engine': 'my_k8s_integration'})
```

### Bugfixes
* Fixes bug where a Kubernetes pod that ran out of memory would fail silently.

## 0.1.4
Released on November 14, 2022.

### Enhancements
* Extends internal integration test framework to support automated testing
    against third-party compute engines.
* Significantly refactors internal data model implementations to improve
    readability and maintainability.

### Bugfixes
* Fixes bug where certain dividers on the navigation sidebar were too wide.
* Fixes bug where opening sidesheets would change page name.
* Fixes bug where function executor Dockerfiles had incorrect start script.
* Fixes bug that caused built-in metric and check functions to have different
    Python environments than regular operators. 

## 0.1.3
Released on November 7, 2022.

### Enhancements
* Surfaces errors with parameter validation in workflow status summary.
* Catches errors generated during `requirements.txt` installation and surfaces
    them eagerly; previously, these errors were ignored.
* Improves operator execution time by only importing `great_expectations` when
    it's being used; the library import is quite slow, so doing it on every
    operator was wasteful.
* Adjusts various font sizes in the UI to improve presentation.
* Adds MongoDB integration.
* Adds `engine` parameter to `global_config`, allowing users to specify a
    default compute engine; `engine` is also now an optional parameter to
    `publish_flow`:
    * If the `engine` argument to `publish_flow` is specified, it will override
        the `global_config`. Otherwise, the engine set in `global_config` will
        be used.
    * If neither the `engine` argument to `publish_flow` or `global_config` is
        set, the workflow will be executed on the default Aqueduct execution
        engine.

### Bugfixes
* Fixes bug where operator details button text overflowed.

## 0.1.2
Released on October 31, 2022.

### Enhancements
* Hides search bar on data viewing page when there are no artifacts.
* Adds support for variable length arguments (`*args`) in Aqueduct functions.

### Bugfixes
* Fixes a bug where updating the metadata of a paused workflow would fail.
* Fixes a bug where parameters were shown as having an upstream function that
    wasn't accessible from the UI.

## 0.1.1
Released on October 25, 2022.

### Enhancements
* Adds support for Tensorflow Keras models to type system.
* Allows users to chain multiple SQL queries in the extract operator.
* Automatically migrates all metadata and artifact snapshots when the user changes the storage layer.
* Re-enables downloading operator code.

### Bugfixes
* Fixes bug where artifact details view was not scrollable in drawer view.
* Fixes bugs where parameter nodes were rendered incorrectly.
* Fixes bug where search functionality was broken on the data page.

## 0.1.0
Released on October 18, 2022.

### Key Features
* Updates the UI to provide a simplified, more responsive layout and surface
    more information about workflow execution. 
    * Adds details pages for operators, artifacts, checks, and metrics which
        show the history and metadata (e.g., code preview, historical values)
        for the relevant object.
    * Replaces old sidesheets with preview of the details pages when clicking
        on a node in the workflow DAG.
    * Adds narrower, simplified navigation sidebar as well as breadcrumbs to
        simplify navigation.
    * Makes page layout more responsive to narrow screens.
* Adds Helm chart to deploy Aqueduct on Kubernetes servers; when running in
    Kubernetes, there's a new integration mechanism to connect Aqueduct to the
    current Kubernetes server that uses an in-cluster client rather than
    `kubeconfig` file.
* When switching Aqueduct metadata stores from local to cloud-hosted,
    automigrates all data to cloud storage.

### Enhancements
* Allows operators to have multiple output artifacts. You can specify the
    number of by using the `num_outputs` argument to the `@op` decorator.
```python
import aqueduct as aq

@aq.op(num_outputs=3)
def multi_output:
  return 1, 2, 3

a, b, c = multi_output()
```
* Enables modifying version history retention policy from the settings pane of
    the workflow page.
* Adds documentation link to menu sidebar.
* Detects when SDK and server version are mismatched and surfaces an error when
    creating SDK client.
* Allows `publish_flow` to accept both a single artifact or a list of multiple
    artifacts in the `artifacts` parameter.
* Moves retention policy parameter from `publish_flow` to `FlowConfig` object.

### Bugfixes
* Fixes bug where tuple return types in operators were not returned correctly.
* Sets minimum version requirements on `pydantic` and `typing-extensions`;
    older versions caused inexplicable and confusing bugs.
* Fixes bug where CSV upload dialog didn't show all rows in data upload
    preview.
* Fixes bug where parameters and checks were marked as canceled when there were
    invalid inputs.
* Fixes bug where Aqueduct logo was cut off on the welcome page on small
    screens.
* Fixes bug where long `stdout` or `stderr` logs were truncated on the UI.
* Fixes bug where SQLite inserts would fail because of an argument limit for
    older versions of SQLite.
* Fixes bug where running Aqueduct operators in temporary environments (e.g.,
    IPython interpreter, VSCode notebooks) would fail because the operator 
    source file would not be detectable.

## 0.0.16
Released on September 26, 2022.

### Enhancements
* Improves the readability of the operator logs printed from the SDK by omitting empty logs and
    making formatting uniform.
* Throws a more informative error message when a table artifact's column name is not of type string.
    Aqueduct currently cannot support DataFrame's with non-string type columns.

### Bugfixes
* Fixes bug where authentication errors caused by incorrect integration credentials were treated as
    system errors, which led to a confusing error message.
* Fixes bug introduced in the previous releases where the settings gear was hidden on the UI.
* Fixes a number of minor formatting and spacing issues on the UI.

## 0.0.15
Released on September 20, 2022.

### Key Features
* Adds support for running new workflows on AWS Lambda and Apache Airflow. Users can define
    workflows using the Aqueduct API but delegate the execution of those workflows onto these
    compute systems.
* Allows Aqueduct parameters to hold any Python object; parameters are also now implicitly 
    created when a Python object is passed into a decorated function.


### Enhancements
* Updates UI to describe database write operators as `save` operators instead of `load` operators to
    avoid confusion.
* Adds `describe` methods to all non-tabular artifact types.

### Bugfixes
* Fixes bug where stack traces and other messages in workflow status bar would
    overflow past edge of screen.
* Fixes bug where some workflows that should have been triggered on server
    start were being ignored due to inconsistent metadata.
* Fixes bug where newest workflow run wasn't shown after a run was manually
    triggered.

## 0.0.14
Released on September 12, 2022.

### Enhancements
* Enables searching through workflows list.
* Workflows are now displayed on the workflows page even before any runs have been created.
* Adds canceled state to operator lifecycle; when upstream operators fail, downstream operators and
    artifact are now marked as canceled rather than being marked as permanently in progress.
* Adds ability to connect new SQLite DB from UI.
* Redesigns integration viewing page to explicitly show DB tables rather than the previous select menu.

### Bugfixes
* Fixes bug where browser console throws error when there is no write operator in workflow DAG.
* Fixes bug where operators previously could not return `None`.

## 0.0.13
Released on September 6, 2022.

### Key Features
* Adds AWS Athena integration. You can now execute SQL queries against AWS Athena using the Aqueduct
    integration API. (Since Athena is a query service, we do not support saving data to Athena.)

### Enhancements
* Removes team and workflow notification categories and simplifies the presentation of the
    notifications pane to be a single box containing all notifications.
* Improves workflow metadata persistence: A newly created workflow will now show on the UI even
    before any runs are finished and persisted.
* Adds support for optionally lazily executing functions during workflow definition. You can also set
    the global configuration for all functions to be lazy by using `aqueduct.global_config({"lazy": True})`.
```python
@op
def my_op(input):
  # ... modify your data...
  return output

result = my_op.lazy(input) # This will not execute immediately.
result.get() # This will force execution of `my_op`.
```
* Enforces typing for saved data; only tabular data is now saveable to relational DBs.
* Makes exported function code human-readable. When you download the code for a function, it will
    include a file with the name of the operator, which will have the function's Python code.

### Bugfixes

None! :tada:

## 0.0.12
Released on August 25, 2022.

### Key Features
* Adds support for running workflows on Kubernetes. You can now register a Kubernetes integration
    from the UI by providing the cluster's kubeconfig file and publish workflows
    to run on Kubernetes by modifying the `config` argument in the SDK's `publish_flow` API. 
* Enables using Google Cloud Storage (GCS) as Aqueduct's metadata store. You can register GCS as a storage
    integration from the UI and store Aqueduct metadata in GCS.

### Enhancements
* Adds support for editing the authentication credentials of existing integrations from the UI.
* Adds support for deleting integrations from the UI.
* Adds support for deleting data created by Aqueduct when deleting a workflow; when deleting a workflow, 
    you will now see an option to select the objects created by this workflow. 
 
### Bugfixes

None! :tada:

## 0.0.11
Released on August 23, 2022.

### Important Note
* If you did a fresh installation of Aqueduct v0.0.10, you may have run into a bug that says our
    schema migrator did not run successfully. To fix this, run `aqueduct clear` and `pip3 install --upgrade aqueduct-ml`.
    You can then start the server via `aqueduct start` and everything should work again.

### Bugfixes
* Fixes a bug where a fresh installation of Aqueduct fails due to a bug in the schema migration process.

## 0.0.10
Released on August 22, 2022.

### Key Features
* Adds support for non-tabular data types; operators can now return any
    Python-serializable object. Under the hood, Aqueduct has special
    optimization for JSON blobs, images, and tables, in addition to supporting
    regular Python objects.
* Enables eager execution when defining workflow artifacts; artifacts are now
    immediately computed at definition time, before calling the `get` API, which
    surfaces potential errors earlier during workflow construction.

### Enhancements
* Caches previously computed function results to avoid repetitive
    recomputation. 
* Enables using AWS S3 as Aqueduct's metadata store; when connecting an S3
    integration, you can now optionally choose to store all Aqueduct metadata
    in AWS S3.

### Bugfixes
* Fixes a bug where the DAG view would ignore the selected version when
    refreshing the page.

## 0.0.9
Released on August 15, 2022.

### Enhancements
* Removes the system name prefix from integration connection form; users found
    this confusing because it was unclear you had to provide a name in addition
    to the prefix.
* Removes deprecated CLI commands, `aqueduct server` and `aqueduct ui`.
* Adds `__str__` method to SDK `TableArtifact` class to support
    pretty-printing.
* Adds support for authenticating with AWS S3 via pre-defined credentials
    files, including when authentication was done via AWS SSO.
    <img width="1683" alt="image" src="https://user-images.githubusercontent.com/867892/184670267-9666b842-7663-406e-adf0-65c2c5c90fc4.png">

### Bugfixes
* Fixes bug where Python requirements weren't properly installed when the client
    and the server ran on different machines.
* Fixes bug where Python stack traces were truncated when running imported
    Python functions.
* Fixes bug where errors generated when uploading a CSV to the Aqueduct demo
    database were formatted poorly and unreadable.
* Fixes bug where SDK client would indefinitely cache the list of connected
    integrations; if a user connected an integration after creating an SDK
    client, that integration would not have been accessible from the SDK
    client.

## 0.0.8
Released on August 8, 2022.

### Enhancements

* Uses `pip freeze` to detect and capture local Python requirements when an
    explicit set of requirements is not specified during function creation.
* Adds download bars to CLI to demonstrate progress when downloading files from
    S3. 
    <img 
         alt="Aqueduct now has progress bars when downloading compiled binaries from S3."
         src="https://user-images.githubusercontent.com/867892/182453985-d0f5408b-8858-46c5-a8bc-e4e198e092ee.png" 
         height="400px"
     />
* When running the Aqueduct server locally, the CLI now automatically opens a
    browser tab with the Aqueduct UI on `aqueduct start` and passes the local
    API key as a query parameter to automatically log in.
* When running on EC2 with `--expose`, detects and populates the public IP 
    address of the current machine in CLI output on `aqueduct start`.
* Makes the file format parameter in the S3 integration a string, so users can
    specify file format by passing in `"csv"`, `"json"`, etc.
* Improves the layout and readability of the integrations UI page by adding
    explicit cards for each integration and also labeling each one with its
    name. <br />
    <img 
         alt="The integrations page has been reorganized to have a border around each image and a corresponding label." 
         src="https://user-images.githubusercontent.com/867892/183465351-fe7724a3-049a-428c-acea-00413a5eea4e.png" 
         height="400px"
    />
* Allows users to create operators from existing functions without redefining
    the operator with a decorator -- using `aqueduct.to_operator`, an existing
    function can be converted into an Aqueduct operator.
* Reduces CLI log output by redirecting info and debug logs to a log file; adds
    a corresponding `--verbose` flag to the CLI so users can see log output in
    terminal if desired.
* Reorganizes integration management behind a dropdown menu, adding option to
    test whether the integration connection still works or not. <br />
    <img
         src="https://user-images.githubusercontent.com/867892/183466408-ffb9f69b-8080-4ce5-ae7e-884f11aae39b.png"
         height="200px"
         alt="A new organization for the integration details page adds an options dropdown next to the upload CSV button."
     />
* Adds "Workflows" section in the integration management page to show all workflows and operators associated with the integration.

### Bugfixes
* Fixes bug where interacting with the UI when the Aqueduct server was
    off resulted in an unhelpful error message ("Failed to fetch."). The fix explicitly
    detects whether the server is unreachable.
* Fixes bug where missing dependencies for integrations (e.g., requiring a
    Python package to access Postgres) were not explicitly surfaced to the user
    -- a cryptic import error message has been replaced with an explicit
    notification that a dependency needs to be installed.
* Fixes bug where metric nodes were misformatted.
* Fixes bug where loading large tables caused UI to significantly slow down
    because React was blindly rendering all cells -- using virtualized tables,
    the UI now only renders the data that is being shown on screen.

## 0.0.7
Released on August 1, 2022.

### Enhancements
* Upgrades to go-chi v5.
* Removes need to provide API key and server address when running client and server on same machine.
* Adds support for operators with no input data.

### Bugfixes
* Fixes bug where imported functions were not executed correctly.
* Improves CSV upload UI to make data preview accurate and more legible.
* Fixes bug where requirements.txt was not consistently used.
* Fixes bug where bottom sidesheet and DAG viewer were misaligned and improperly sized.

## 0.0.6
Released on July 25, 2022.

### Enhancements
* Prints error message as part of preview execution stack trace, not above it.

### Bugfixes
* Fixes bug where parameters argument to `head` function was unused.
* Fixes bug where menu sidebar didn't link to home page.
* Fixes bug where operator zipfiles weren't cleaned up after workflow creation.
* Fixes bug where S3 connection listed all objects in bucket, causing connection to be extremely slow.
* Fixes bug where error and warning checks aren't properly distinguished.

## 0.0.5
Released on July 14, 2022.

### Enhancements
* Makes password optional when creating a Postgres connection.
* Adds `describe` method to every relational integration.
* Improves log capture when executing user functions.
* Enables configuration of S3 storage backend for version snapshots and operator metadata.
* Displays workflow ID on workflow settings modal.
* Adds ability to fetch an individual artifact from a workflow run using the SDK.
* Supports reading multiple S3 files into a single Pandas DataFrame.
* Deprecates showing `pyplot` image in notebook on workflow creation; instead, provides link to UI.

### Bugfixes

None! :tada:

## 0.0.4
Released on July 7, 2022.

### Key Features
* Workflows can now have custom parameters! A workflow can have any numbers of parameters which can be used in Python operators or
  SQL queries. See [here](https://docs.aqueducthq.com/workflows/parameterizing-a-workflow) for more details.

### Enhancements
* Add SDK support for fetching and pretty-printing workflow and workflow run metadata.
* Hide success notifications by default to avoid repetitive notifications.
* Allow for custom port selection for the Postgres integration.
* Allow requirements.txt to be set on an operator-by-operator basis.
* Add ability to copy SDK initialization snippet from account page.
* Allows metrics to be integers in addition to floats.
* Adds syntax candy for `head` on `TableArtifact`s.

### Bugfixes
* Fix bug that showed undefined in search bar when data search returned empty results.
* Fix bug where integration passwords were shown in plaintext on request headers.
* Fix bug where schema metadata was improperly persisted.
* Fix bug that disallowed non-CSV file uploads.
* Fix bug that caused unnecessary repetitive calls to the DAG render API.
* Fix a number of minor UI bugs -- margins, button placement, etc.
* Deprecates use of ipynbname in the SDK, which prevented the SDK from running in some notebook environments.

## 0.0.3
Released on June 21, 2022.

### Key Features
* View what tables are present in an integration by clicking on the integration in the UI.
* View all data artifacts created by Aqueduct on the `/data` page on the UI.
* Add support for pre-defined metrics and checks, including lower & upper bounds and equality checks.
* Implement support for capturing low-level metrics, such as compute time, CPU usage, and memory usage, on a per-operator basis. 

### Enhancements
* API keys can now be retrieved from the SDK if running on the same machine as the Aqueduct server with `aqueduct.get_apikey()`.
* Add feature to automatically search for next available port when port 8080 is occupied.
* Users can upload custom data to the Aqueduct demo DB — navigate to the integrations page, click on the Aqueduct Demo database, and hit “Add CSV”.
* Allow users to optionally specify HTTP/S prefix when creating Aqueduct API client.
* Implements support for creating checks via [Great Expectations](https://greatexpectations.io/).
* Simplifies notifications interface by reducing redundant text.

### Bugfixes
* Fix bug where logs directory didn’t exist on upgraded installations. 
* Fix bug where account page wasn’t previously being displayed.

### Contributors

* [Kenneth Xu](https://github.com/kenxu95)
* [Vikram Sreekanti](https://github.com/vsreekanti)
* [Chenggang Wu](https://github.com/cw75)
* [Fanjia Yan](https://github.com/Fanjia-Yang)
* [Haris Choudhary](https://github.com/HarisChoudhary)
* [Andre Giron](https://github.com/agiron123)
* [Hari Subbaraj](https://github.com/hsubbaraj-spiral)
* [Eunice Chan](https://github.com/eunice-chan)
* [Saurav Chhatrapati](https://github.com/saurav-c)
* [Boyuan Deng](https://github.com/Boyuan-Deng)

## 0.0.2
Released on June 9, 2022.

### Enhancements
* Allows users to start both the backend server and UI with `aqueduct start`
* Removes NextJS from UI project, reverting to a vanilla React app packaged with Parcel
* Removes need for users to have `npm` installed by serving the UI from the same server as the backend
* Automatically ships common library as a transpiled module, removing need for explicit transpilation
* Allows users to retrieve package version by running `aqueduct version`
* Only binds server to `localhost` by default, removing requirement for firewall permissions
* Improves the thread safety of the job manager
* Allow users to execute annotated functions by calling `fn.local(args)`

### Bugfixes
* Fixes incorrect use of `typing` library for Python3.7
* Fixes inconsistency in DAG rendering which would previously cause page load jitter
* Fixes bug where bounds on metrics were mislabeled

### Contributors

* [Joey Gonzalez](https://github.com/jegonzal)
* [Kenneth Xu](https://github.com/kenxu95)
* [Vikram Sreekanti](https://github.com/vsreekanti)
* [Chenggang Wu](https://github.com/cw75)
* [Boyuan Deng](https://github.com/Boyuan-Deng)
* [Fanjia Yan](https://github.com/Fanjia-Yang)
* [Haris Choudhary](https://github.com/HarisChoudhary)
* [Andre Giron](https://github.com/agiron123)
* [Will Crosier](https://github.com/datadawg88)
* [Wei Chen](https://github.com/likawind)

## 0.0.1
Released on May 26, 2022.

Initial release of the Aqueduct open-source project.

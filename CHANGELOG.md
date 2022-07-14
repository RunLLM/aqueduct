# Changelog

## 0.0.5
Released on 7/14/2022

### Enhancments
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
Released on 7/7/2022

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
Released on 6/21/2022

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
Released on 6/9/2022

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
Released on 5/26/2022

Initial release of the Aqueduct open-source project.

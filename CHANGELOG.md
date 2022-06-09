# Changelog

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

## 0.0.1
Released on 5/26/2022

Initial release of the Aqueduct open-source project.
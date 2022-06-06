# Changelog

## 0.0.2

### Enhancements
* Allows users to start both the backend server and UI with `aqueduct start`
* Removes NextJS from UI project, reverting a React app
* Removes need for users to have `npm` installed 
* Automatically ships common library as a transpiled module, removing need for explicit transpilation
* Allows users to retrieve package version by running `aqueduct version`
* Only binds server to `localhost` by default, removing requirement for firewall permissions

### Bugfixes
* Fixes incorrect use of `typing` library for Python3.7
* Fixes inconsistency in DAG rendering which would previously cause page load jitter

### Contributors

## 0.0.1
Released on 5/26/2022

Initial release of the Aqueduct open-source project.
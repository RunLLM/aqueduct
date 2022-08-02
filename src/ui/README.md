# Aqueduct UI

The Aqueduct UI is implemented in React and Typescript. Once you've installed the Aqueduct
package, you can start the aqueduct server and a production build of the UI by running `aqueduct start`.

For information on how to run the UI in development mode, please refer to the Development Guide section below.

The Aqueduct UI code is organized into two directories -- a component library in
`common/` that is published to npm as a package called `@aqueducthq/common`, and an application server in `app/`
that consumes `@aqueducthq/common`. This package is also shared with the enterprise edition of Aqueduct.

If you're actively developing the UI, you will need create a file called ```.env.local``` in the root of `src/ui/app/`.

This file should contain the following:

```SERVER_ADDRESS=http://localhost:8080```

If you are exposing the ui at a public IP, make
sure to update the `SERVER_ADDRESS` field accordingly (e.g. 
`SERVER_ADDRESS=http://3.142.105.48:8080`).

Please note that you should include ```http://``` at the beginning of SERVER_ADDRESS for things to work correctly. 

## Development Guide

### Node and npm versions
We use [nvm](https://github.com/nvm-sh/nvm) to manage node.js and npm versions.

For convenience, we have added a .nvmrc file at the root of the ui folder to make sure all developers use the same
version of node.js.

To use the .nvmrc, from the ```src/ui```, simply run:

```nvm use```

NOTE for Linux Users: do NOT run npm under sudo when using nvm as a version manager. This causes a different version 
of npm to be used than the version that nvm is using.

### Installing Dependencies

Since ```common/``` is used throughout the ```app/``` project, you must first install dependencies in common and link
to the package with npm.

### 1. Install @aqueducthq/common's dependencies:

```cd src/ui/common```

```npm install```

The install command above will both install dependencies to node_modules and build a JavaScript bundle to be used by ```app```

```npm link```

The link command above will create a symlink to your local ```@aqueducthq/common``` package which will be used in ```app```.

### 2. Install app's dependencies:

```cd src/ui/app```

```npm link @aqueducthq/common```

The link command above will install the @aqueducthq/common package that was symlinked via ```npm link``` and will also install
the remaining dependencies for ```app```.

### Active development:
To enable local development with automatic reloading, open up two terminal windows and do the following:

In the first terminal window, navigate to ```src/ui/common``` and run:

```npm run start```

Finally, in the second terminal window navigate to ```src/ui/app``` and run:

```npm run start```

Now, you should see that your development server has started at http://localhost:1234, which is the default port that Parcel
uses.

You should now be able to edit files ```app``` and see your changes reflected live.

After editing files in ```common``` to see your latest changes, stop and start the server in ```app```. Since ```npm run start```
runs the ```common``` server in watch mode, can leave common running in watch mode to have changes built on file save.

### Release Process for @aqueducthq/common:
After modifying code in @aqueducthq/common, please update the version number in src/ui/common/package.json accordingly.

At the moment, we aren't following semantic versioning, but rather bumping the version number on each release of the
aqueduct client. 

When developing, please use the following pattern:
```v0.0.1-rc0```, ```v0.0.1-rc1```, etc. for changes in active devlopment that need to be shared with the enterprise version.

The suffix ```-rc0``` signifies that the current version is a release candidate for version 0.0.1. 

When finished developing, update the version number to ```0.0.1``` and release alongside other Aqueduct components.

### Future Work:
- Use npm/yarn workspaces for a better development workflow to handle getting latest changes from @aqueducthq/common.
- Better monorepo support for project. There are several build tools that we can use for this:
    - [Turborepo](https://turborepo.org/)
    - [Moon](https://moonrepo.dev/)
    - [Nx](https://nx.dev/)



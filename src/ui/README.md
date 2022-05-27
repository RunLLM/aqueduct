# Aqueduct UI

The Aqueduct UI is implemented in Next.js. Once you've installed the Aqueduct
package, you can start the UI by running `aqueduct ui`.

The code here is organized into two directories -- a component library in
`components/` that is published to npm, and an application server in `app/`
that consumes `components/`.

If you're actively developing the UI, you can start the Next.js server in dev
mode by running `make dev`. Note that this will symlink your local version of
the `components/` directory to pick up any changes you might have made. 

More details coming soon!

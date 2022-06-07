# Aqueduct UI

The Aqueduct UI is implemented in Next.js. Once you've installed the Aqueduct
package, you can start the UI by running `aqueduct ui`.

The code here is organized into two directories -- a component library in
`components/` that is published to npm, and an application server in `app/`
that consumes `components/`.

If you're actively developing the UI, you will need to copy the `.env.local` file 
to `/app` from `~/.aqueduct/ui/app`. If you are exposing the ui at a public IP, make
sure to update the `SERVER_ADDRESS` field accordingly (e.g. 
`SERVER_ADDRESS=3.142.105.48:8080`). Then, you can start the Next.js server in dev 
mode by running `make dev`. Note that this will symlink your local version of
the `components/` directory to pick up any changes you might have made. 

More details coming soon!

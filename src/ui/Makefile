.DEFAULT_GOAL := dev

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

# This command can be used to temporarily link a version of the component
# library that you've developed locally into the app UI.
link-dev:
	cd common && rm -rf node_modules && npm install && sudo npm link
	cd app && npm link @aqueducthq/common

dev: link-dev
	cd app && npm run start

dist:
	# Clear any previous distributions.
	cd app && rm -rf dist && rm -rf .parcel-cache

	# This is to ensure that the server address is empty in production mode. We
	# do this because we run the UI and the backend on the same Go server, so we
	# don't need to specify an address when making a request.
	cd app && echo "SERVER_ADDRESS=" > .env
	cd app && npm run build:dist


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
	cd app && npm run dev

build:
	cd app && npm run build

prod: build
	cd app && npm run start

run:
	cd app && npm run start

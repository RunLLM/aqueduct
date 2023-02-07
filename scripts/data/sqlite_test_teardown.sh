#!/usr/bin/env bash

# This script deletes the specified directory. It should be used to 
# remove the directory created to house the SQLite database.
# By default it assumes the directory is `~/tests`. However, you can
# change this behavior by passing in the directory to be deleted as
# a command-line argument.
# Does nothing if the directory does not exist.

# Example usage: `bash scripts/data/sqlite_test_teardown.sh ./sqlite-test`
directory=${1:-"${HOME}/tests"}

if [ ! -d ${directory} ]; then
  echo "Directory not found."
  exit 1
fi

rm -r "${directory}"
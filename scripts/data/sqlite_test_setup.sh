#!/usr/bin/env bash

# This script creates a SQLite database file called `test.db` in the 
# specified directory. 
# By default it assumes the directory is `~/tests`. However, you can
# change this behavior by passing in the directory to use as a
# command-line argument.
# It creates the directory if the directory does not exist.

# Example usage: `bash scripts/data/sqlite_test_setup.sh ./sqlite-test`

directory=${1:-"${HOME}/tests"}

if [ ! -d ${directory} ]; then
  mkdir -p ${directory};
fi

sqlite3 "${directory}/test.db" "VACUUM;"

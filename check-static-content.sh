#!/bin/bash

set -e

function assert_no_changes() {
  if [[ `git status --porcelain` ]]; then
    echo $1
    exit 1
  else
    echo $2
  fi
}

assert_no_changes "Dirty repo, commit changes before running this command"

make statik
assert_no_changes "Static content is out of date. Run \"make statik\" and commit the changes." \
  "Static content is up to date."

make compile
./gflows init --engine ytt
assert_no_changes "Default workflow files don't match generated files. Contents of static-content and .gflows need to match." \
  "Default workflow files are up to date."

#!/bin/bash

set -e

function assert_no_changes() {
  if [[ `git status --porcelain` ]]; then
    printf "ERROR: %s\n" "$1"
    exit 1
  else
    printf "OK: %s\n\n" "$2"
  fi
}

echo "Checking git repo status"
assert_no_changes "Dirty repo, commit changes before running this command" \
  "Git repo is clean, continuing..."

make statik
assert_no_changes "Static content is out of date. Run \"make statik\" and commit the changes." \
  "Static content is up to date."

# make compile
# ./gflows --config .gflows/examples/default-ytt/config.yml init \
#   --engine ytt --workflow-name ex-default-ytt-gflows --github-dir ../../.github
# ./gflows --config .gflows/examples/default-jsonnet/config.yml init \
#   --engine jsonnet --workflow-name ex-default-jsonnet-gflows --github-dir ../../.github
# assert_no_changes "Default workflow files don't match generated files. Contents of static/content and .gflows need to match." \
#   "Default workflow files are up to date."

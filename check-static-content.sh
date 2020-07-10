#!/bin/bash

make statik

if [[ `git status --porcelain` ]]; then
  echo "Static content is out of date. Run \"make statik\" and commit the changes."
  exit 1
else
  echo "Static content is up to date."
fi

make build

./jflows init

if [[ `git status --porcelain` ]]; then
  echo "Default workflow files don't match generated files. Contents of static-content and .jflows need to match."
  exit 1
else
  echo "Default workflow files are up to date."
fi

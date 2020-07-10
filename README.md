# GFlows - GitHub Workflow Templates

[![Build Status](https://github.com/jbrunton/gflows/workflows/ci-build/badge.svg?branch=master)](https://github.com/jbrunton/gflows/actions?query=branch%3Amaster+workflow%3Aci-build)
[![Maintainability](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/maintainability)](https://codeclimate.com/github/jbrunton/gflows/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/test_coverage)](https://codeclimate.com/github/jbrunton/gflows/test_coverage)

GFlows provides a templating mechanism for GitHub Workflows, using Jsonnet. It comprises a CLI tool that can:

* Import existing workflows into GFlow templates.
* Validate GitHub workflows are up to date with their source templates.
* Watch changes to the templates, so you can develop and refactor workflows with instant feedback on your changes.

## Installing

    go get github.com/jbrunton/gflows

## Getting Started

You'll probably want to run the `init` command the first time you use GFlows:

    gflows init
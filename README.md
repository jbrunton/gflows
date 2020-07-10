# GFlows - GitHub Workflow Templates

[![Build Status](https://github.com/jbrunton/gflows/workflows/build/badge.svg?branch=develop)](https://github.com/jbrunton/gflows/actions?query=branch%3Adevelop+workflow%3Abuild)
[![Maintainability](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/maintainability)](https://codeclimate.com/github/jbrunton/gflows/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/test_coverage)](https://codeclimate.com/github/jbrunton/gflows/test_coverage)

GFlows provides a templating mechanism for GitHub Workflows, using Jsonnet. It comprises a CLI tool that can:

* Import existing workflows into GFlow templates.
* Validate GitHub workflows are up to date with their source templates.
* Watch changes to the templates, so you can develop and refactor workflows with instant feedback on your changes.

## Installing

    go get github.com/jbrunton/gflows

## Getting Started

### Adding GFlows to a repository

► First, you'll probably want to run the `init` command to bootstrap GFlows:

    $ gflows init
         create .gflows/workflows/common/steps.libsonnet
         create .gflows/workflows/common/workflows.libsonnet
         create .gflows/workflows/config/git.libsonnet
         create .gflows/workflows/gflows.jsonnet
         create .gflows/config.yml

This generates:

* A workflow called `gflows` defined in `gflows.libsonnet`, which will run against PRs and your main branch to ensure your workflows are kept up to date with their source templates.
* Some common code factored out into libsonnet files in the `config/` and `common/` directories.

► At this point, you should update the `config/git.libsonnet` file to reference the correct name of your main branch.

► Finally, run the `update` command to create the `gflows` workflow:

    $ gflows update
         create .github/workflows/gflows.yml (from .gflows/workflows/gflows.jsonnet)

### Importing existing workflows

► If you want to import your existing workflows, you can use the `import` command:

    $ gflows import
    Found workflow: .github/workflows/my-workflow.yml
      Imported template: .gflows/workflows/my-workflow.jsonnet
    
    Important: imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.
      ► Run "gflows update" to do this now

► Because Jsonnet (very probably) renders yaml differently from your existing workflow, you'll need to run the `update` command to regenerate your workflows:

    $ gflows update
         update .github/workflows/my-workflow.yml (from .gflows/workflows/my-workflow.jsonnet)

At this point you can commit and push your changes. If you create a PR against your main branch you should see the `gflows` workflow checking your workflows are up to date.


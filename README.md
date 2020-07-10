# GFlows - GitHub Workflow Templates

[![Build Status](https://github.com/jbrunton/gflows/workflows/build/badge.svg?branch=develop)](https://github.com/jbrunton/gflows/actions?query=branch%3Adevelop+workflow%3Abuild)
[![Maintainability](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/maintainability)](https://codeclimate.com/github/jbrunton/gflows/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/test_coverage)](https://codeclimate.com/github/jbrunton/gflows/test_coverage)

GFlows provides a templating mechanism for GitHub Workflows, using [Jsonnet](https://jsonnet.org/). It comprises a CLI tool that can:

* Import existing workflows into GFlow templates.
* Validate GitHub workflows are up to date with their source templates.
* Watch changes to the templates, so you can develop and refactor workflows with instant feedback on your changes.

## Installing

    go get github.com/jbrunton/gflows

## Getting Started

### Adding GFlows to a repository

First, you'll probably want to run the `init` command to bootstrap GFlows:

    $ gflows init
         create .gflows/workflows/common/steps.libsonnet
         create .gflows/workflows/common/workflows.libsonnet
         create .gflows/workflows/config/git.libsonnet
         create .gflows/workflows/gflows.jsonnet
         create .gflows/config.yml

This generates:

* A workflow called `gflows` defined in `gflows.libsonnet`, which will run against PRs and your main branch to ensure your workflows are kept up to date with their source templates.
* Some common code factored out into libsonnet files in the `config/` and `common/` directories.

At this point, you should update the `config/git.libsonnet` file to reference the correct name of your main branch.

Finally, run the `update` command to create the `gflows` workflow:

    $ gflows update
         create .github/workflows/gflows.yml (from .gflows/workflows/gflows.jsonnet)

### Importing existing workflows

If you want to import your existing workflows, you can use the `import` command:

    $ gflows import
    Found workflow: .github/workflows/my-workflow.yml
      Imported template: .gflows/workflows/my-workflow.jsonnet
    
    Important: imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.
      ► Run "gflows update" to do this now

Because Jsonnet (very probably) renders yaml differently from your existing workflow, you'll need to run the `update` command to regenerate your workflows:

    $ gflows update
         update .github/workflows/my-workflow.yml (from .gflows/workflows/my-workflow.jsonnet)

At this point you can commit and push your changes. If you create a PR against your main branch you should see the `gflows` workflow checking your workflows are up to date.

## Validating your workflows

You can validate and verify your workflows with the `check` command:

    $ gflows check
    Checking gflows ... OK
    Checking my-workflow ... OK
    Workflows up to date

By default this command will check, for each workflow:

* That the Jsonnet source is valid.
* That the content of the generated workflow file in .github/workflows is up to date.
* That the workflow is validated by the [github-workflow schema](https://json.schemastore.org/github-workflow) from [schemastore.org](https://www.schemastore.org/json/). (Note that this schema is comprehensive but may fail for occasional edge cases. You can disable schema validation on a per workflow basis if need be.)

## Refactoring Workflows

One of the joys of Jsonnet is it gives you a whole host of options (including objects, functions and library files) for refactoring complex workflows.

To make the process of refactoring easier, you can run the `check` command with `--watch` and `--diff` flags. While refactoring, you should see no changes to the generated workflow, so any changes indicate an error in the refactor: the diff output will quickly show you what it is.

```
    $ gflows check --watch --show-diff
    Checking gflows ... UP TO DATE
    Checking my-workflow ... UP TO DATE
```

If you [install pygments](https://pygments.org/docs/cmdline/) then the diff will include syntax highlighting. For example:

![Example output from check command](https://raw.githubusercontent.com/jbrunton/gflows/develop/workflow-checks.png)

## Using remote templates

If you want to extract templates into a separate repository then the recommended approach is to use [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler).

Any additional library paths should be added to the `jsonnet.jpath` list in the config file. These paths may be relative, so, for example, if you install dependencies into `.jflows/vendor`, then your config may look like this:

```
jsonnet:
  jpath:
  - vendor
```

## Configuration

The config file (`.gflows/config.yml`) can be edited to configure validation and jsonnet options. The options look like this:

```yaml
# The .github directory. Set this if you put `gflows` in a different directory than the default.
# Default: .github
githubDir: .github

# Default options for generating workflows
defaults:
  # The checks to conduct when running `gflows check`
  checks:
    schema:
      # Whether or not to validate with a JSON schema.
      # Default: true
      enabled: true
      
      # The schema to use.
      # Default: https://json.schemastore.org/github-workflow
      uri: https://example.com/my-schema

    content:
      # Whether or not to validate that the workflow in .github is up to date
      # Default: true
      enabled: true

# Overrides for specific workflows
workflows:
  # For example, this overrides the schema options for my-workflow
  my-workflow:
    checks:
      schema:
        enabled: false

# Jsonnet options
jsonnet:
  # Additional paths to search for libraries. Useful if you use jsonnet-bundler.
  # Default: <empty list>
  jpath:
  - vendor
  - my-library
```

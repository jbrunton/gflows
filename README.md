# GFlows - GitHub Workflow Templates

[![Build Status](https://github.com/jbrunton/gflows/workflows/build/badge.svg?branch=develop)](https://github.com/jbrunton/gflows/actions?query=branch%3Adevelop+workflow%3Abuild)
[![Maintainability](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/maintainability)](https://codeclimate.com/github/jbrunton/gflows/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/test_coverage)](https://codeclimate.com/github/jbrunton/gflows/test_coverage)

GFlows is a CLI tool that makes templating GitHub Workflows easy, using either [Jsonnet](https://jsonnet.org/) or [ytt (Yaml Templating Tool)](https://get-ytt.io/). It can:

* Import existing workflows to help you quickly get started.
* Validate GitHub workflows are up to date with their source templates and conform to a valid schema.
* Watch changes to the templates, so you can develop and refactor workflows with fast feedback on your changes.

Note: this project is very new, so I expect there is room for improvement (especially around error handling). But I've used it comfortably in my own projects, and the risk of adoption is low since it mostly just builds on top of existing tooling (primarily Jsonnet). If you have any feedback I'd love to hear it!

## Contents

* [Installing](#installing)
* [Getting Started](#getting-started)
    * [Adding GFlows to a repository](#adding-gflows-to-a-repository)
    * [Importing existing workflows](#importing-existing-workflows)
* [Validating Workflows](#validating-workflows)
* [Refactoring Workflows](#refactoring-workflows)
* [Configuration](#configuration)
* [Development Tips](#development-tips)
* [Examples](#examples)


## Installing

    go get github.com/jbrunton/gflows

## Getting Started

### Adding GFlows to a repository

First, you'll probably want to run the `init` command to bootstrap GFlows. To use ytt as the templating engine:

    $ gflows init --engine ytt
         create .gflows/workflows/common/steps.lib.yml
         create .gflows/workflows/common/workflows.lib.yml
         create .gflows/workflows/config/git.yml
         create .gflows/workflows/gflows/gflows.yml
         create .gflows/config.yml

(You can also use jsonnet with `gflows init --engine jsonnet`.)

This generates:

* A workflow called `gflows` defined in `gflows.yml`, which will run against PRs and your main branch to ensure your workflows are kept up to date with their source templates.
* Some common code factored out into library files in the `config/` and `common/` directories.
* A `config.yml` file to customize build and validation options (see [Configuration](#configuration) for more details).

At this point, you should update the `config/git.yml` file to reference the correct name of your main branch.

Finally, run the `update` command to create the `gflows` workflow:

    $ gflows update
         create .github/workflows/gflows.yml (from .gflows/workflows/gflows)

### Importing existing workflows

If you want to import your existing workflows, you can use the `import` command:

    $ gflows import
    Found workflow: .github/workflows/my-workflow.yml
      Imported template: .gflows/workflows/my-workflow/my-workflow.yml
    
    Important: imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.
      ► Run "gflows update" to do this now

Note that the yaml generated will be a little different from your existing workflow so you'll need to run the `update` command to regenerate your workflows:

    $ gflows update
         update .github/workflows/my-workflow.yml (from .gflows/workflows/my-workflow/my-workflow.yml)

At this point you can commit and push your changes. If you create a PR against your main branch you should see the `gflows` workflow checking your workflows are up to date.

## Validating Workflows

You can validate and verify your workflows with the `check` command:

    $ gflows check
    Checking gflows ... OK
    Checking my-workflow ... OK
    Workflows up to date

By default this command will check, for each workflow:

* That the source templates (jsonnet, ytt) are valid.
* That the content of the generated workflow file in .github/workflows is up to date.
* That the workflow is validated by the [github-workflow schema](https://json.schemastore.org/github-workflow) from [schemastore.org](https://www.schemastore.org/json/). (Note that this schema is comprehensive but may fail for occasional edge cases. You can disable schema validation on a per workflow basis if need be.)

If it fails any of the validation checks, you'll see clear errors describing the problem:

![Example output from check command](https://raw.githubusercontent.com/jbrunton/gflows/develop/workflow-checks.png)

## Refactoring Workflows

One of the joys of a real templating language (whether ytt or jsonnet) is it gives you a whole host of options for refactoring complex workflows.

To make the process of refactoring easier, you can run the `watch` command (which is just an alias for `check --watch --show-diffs`). While refactoring, you should see no changes to the generated workflow, so any changes indicate an error in the refactor, and the diff output should quickly show you what it is.

```
    $ gflows watch
    2020/07/10 18:43:56 Watching workflow templates
      Watching .gflows/workflows/my-workflow.jsonnet
      Watching .github/workflows/my-workflow.yml
    Checking my-workflow ... UP TO DATE
```

If you [install bat](https://github.com/sharkdp/bat) then the diff will include syntax highlighting. For example:

![Example output from check command](https://raw.githubusercontent.com/jbrunton/gflows/develop/workflow-diff.png)

## Using jsonnet-bundler

If you use jsonnet as the templating engine and want to extract templates into a separate repository then the recommended approach is to use [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler).

Any additional library paths should be added to the `templates.libs` list in the config file. These paths may be relative, so, for example, if you install dependencies into `.jflows/vendor`, then your config may look like this:

```
templates:
  libs:
  - vendor
```

## Configuration

The config file (`.gflows/config.yml`) can be edited to configure validation and jsonnet options. The options look like this:

```yaml
# The .github directory. Set this if you put `gflows` in a different directory than the default.
# Default: .github
githubDir: .github

workflows:
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

templates:
  # The templating engine to use (either ytt or jsonnet)
  engine: ytt

  # Additional paths to search for libraries.
  #
  # For ytt, these paths are passed as files (i.e. the -f flag) to ytt. Note that it's important to
  # add all ytt library files here or they may be treated as workflows.
  #
  # For jsonnet, libs are treated as jpaths. For example, if you use jsonnet-bundler to add
  # library files to .gflows/vendor, add them like this:
  #   libs:
  #   - vendor
  libs:
  - workflows/common
  - workflows/lib
```

## Development Tips

* You can get Jsonnet syntax highlighting and autocomplete in VS Code from the [Jsonnet Language Support](https://marketplace.visualstudio.com/items?itemName=liamdawson.jsonnet-language) extension.
* You can get workflow schema validation directly in VS Code if you add the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension and add the below to your settings.json:

```json
    "yaml.schemas": {
        "https://json.schemastore.org/github-workflow": ["/.github/workflows/*.yml"]
    }
```

## Examples

* [This PR](https://github.com/jbrunton/bechdel-lists/pull/190/files) from another project I was working on shows how I was able to break up two very awkward workflow files into multiple smaller ones (and also reduce the loc). It was a large refactor, but only took a few minutes with the feedback from `gflows watch`.


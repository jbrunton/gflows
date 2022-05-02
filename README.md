# GFlows - GitHub Workflow Templates

[![Build Status](https://github.com/jbrunton/gflows/workflows/build/badge.svg?branch=develop)](https://github.com/jbrunton/gflows/actions?query=branch%3Adevelop+workflow%3Abuild)
[![Maintainability](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/maintainability)](https://codeclimate.com/github/jbrunton/gflows/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/02363f0b2588376bbf98/test_coverage)](https://codeclimate.com/github/jbrunton/gflows/test_coverage)

GFlows is a CLI tool that makes templating GitHub Workflows easy, using either [Jsonnet](https://jsonnet.org/) or [ytt (Yaml Templating Tool)](https://get-ytt.io/). It can:

* Import existing workflows to help you quickly get started.
* Validate GitHub workflows are up to date with their source templates and conform to a valid schema.
* Share common config code and workflows with [GFlows Packages](https://github.com/jbrunton/gflows/wiki/GFlows-Packages).
* Watch changes to the templates, so you can develop and refactor workflows with fast feedback on your changes.

![Example refactor](https://raw.githubusercontent.com/jbrunton/gflows/develop/refactor.gif)

Note: this project is very new so I expect there is room for improvement, but I've used it comfortably in my own projects and the risk of adoption is low since it mostly just builds on top of existing tooling (primarily Jsonnet and ytt). If you have any feedback I'd love to hear it!

## Installing

Either download from [Releases](https://github.com/jbrunton/gflows/releases) or install with Go:

    go install github.com/jbrunton/gflows@latest
    
You can also install in GitHub workflows using the [setup-gflows](https://github.com/jbrunton/setup-gflows) action:

```yaml
steps:
- uses: jbrunton/setup-gflows@v1
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
- run: gflows check
```

## Getting Started

See [Getting Started](https://github.com/jbrunton/gflows/wiki/Getting-Started).

## Docs

See [the wiki](https://github.com/jbrunton/gflows/wiki) for detailed documentation.


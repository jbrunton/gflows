# Templates

GFlows supports [Jsonnet](https://jsonnet.org/) and [ytt (Yaml Templating Tool)](https://get-ytt.io/). In both cases it uses them in slightly specific ways in order to generate distinct workflows.

## Jsonnet

When using Jsonnet, GFlows will treat any file matching the following glob as a workflow template:

    .gflows/workflows/**/*.jsonnet

The workflow name will be derived from the jsonnet file name. For example, a template `my-workflow.libsonnet` will be used to generate a workflow `my-workflow.yml`.

As with Jsonnet, library (.libsonnet) files can be added anywhere to the file system and referenced with relative paths.

### Library files

To add library files or directories to the JPATH, add a `templates.libs` field to `.gflows/config.yml`:

```yaml
templates:
  libs:
    - path/to/my/lib
```

This is equivalent to `jsonnet -J path/to/my/lib`.

### Using jsonnet-bundler

If you use jsonnet as the templating engine and want to extract templates into a separate repository then the recommended approach is to use [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler).

Any additional library paths should be added to the `templates.libs` list in the config file. These paths may be relative, so, for example, if you install dependencies into `.jflows/vendor`, then your config may look like this:

```yaml
templates:
  libs:
  - vendor
```

## Ytt

When using Ytt, GFlows will treat any directory in `.gflows/workflows` which itself contains yaml files as a workflow template, named after the directory. For example, given these files:

    .gflows/workflows/my-workflow/config.yml
    .gflows/workglows/my-worfklow/workflow.yml

GFlows will treat `.gflows/workflows/my-workflow` as the template for a workflow `my-workflow.yml`. The file generated will be equivalent to running ytt manually like this:

    ytt -f .gflows/workflows/my-workflow

### Library files

If you want to add library files (i.e. files which are are used by multiple workflows but which shouldn't be treated as workflow templates themselves), you can do that with the `templates.libs` property in `.gflows/config.yml`:

```yaml
templates:
  libs:
  - my-lib
```

For example, given these files and the above config:

    .gflows/workflows/my-lib/common.yml
    .gflows/workflows/my-workflow/config.yml
    .gflows/workglows/my-worfklow/workflow.yml

GFlows will treat `.gflows/workflows/my-workflow` as the template for a workflow `my-workflow.yml`, and the file generated will be equivalent to running ytt manually like this:

    ytt -f .gflows/workflows/my-workflow -f .gflows/workflows/my-lib

## Development Tips

* You can get Jsonnet syntax highlighting and autocomplete in VS Code from the [Jsonnet Language Support](https://marketplace.visualstudio.com/items?itemName=liamdawson.jsonnet-language) extension.
* You can get workflow schema validation directly in VS Code if you add the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension and add the below to your settings.json:

```json
    "yaml.schemas": {
        "https://json.schemastore.org/github-workflow": ["/.github/workflows/*.yml"]
    }
```

## Examples

* [This PR](https://github.com/jbrunton/bechdel-lists/pull/190/files) from another project I was working on shows how I was able to break up two very awkward workflow files into multiple smaller Jsonnet files (and also reduce the loc). It was a large refactor, but was pretty quick and easy given the fast feedback using `gflows watch`.

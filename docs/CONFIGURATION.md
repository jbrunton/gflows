# Configuration

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

# Refactoring Workflows

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

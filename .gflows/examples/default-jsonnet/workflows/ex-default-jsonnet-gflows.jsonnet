local git_config = import 'git.libsonnet';
local steps = import 'steps.libsonnet';
local workflows = import 'workflows.libsonnet';

local check_workflows_job = {
  'name': 'check-workflows [ex-default-jsonnet-gflows]',
  'runs-on': 'ubuntu-latest',
  steps: [
    steps.checkout,
    steps.setup_gflows,
    steps.check_workflows
  ]
};

local workflow = {
  name: 'gflows',
  on: workflows.triggers.pull_request_defaults,
  jobs: {
    check_workflows: check_workflows_job
  },
};

std.manifestYamlDoc(workflow)

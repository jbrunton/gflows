local git_config = import 'config/git.libsonnet';
local steps = import 'common/steps.libsonnet';
local workflows = import 'common/workflows.libsonnet';

local check_workflows_job = {
  'name': 'check-workflows',
  'runs-on': 'ubuntu-latest',
  steps: [
    steps.checkout,
    steps.setup_go,
    steps.named('install jflows', 'go get github.com/jbrunton/jflows/cmd/jflows'),
    steps.named('validate workflows', 'jflows check')
  ]
};

local workflow = {
  name: 'jflows-workflows',
  on: workflows.triggers.pull_request_defaults,
  jobs: {
    check_workflows: check_workflows_job
  },
};

std.manifestYamlDoc(workflow)

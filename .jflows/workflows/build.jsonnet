local git_config = import 'config/git.libsonnet';
local steps = import 'common/steps.libsonnet';
local workflows = import 'common/workflows.libsonnet';

local test_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
      steps.checkout,
      steps.setup_go,
      steps.run('go test ./...')
  ]
};

local build_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
    steps.checkout,
    steps.setup_go,
    steps.run('go get github.com/rakyll/statik'),
    steps.run('go build'),
    steps.run('./check-static-content.sh')
  ],
};

local workflow = {
  name: 'build',
  on: workflows.triggers.pull_request_defaults,
  jobs: {
    test: test_job,
    build: build_job
  },
};

std.manifestYamlDoc(workflow)

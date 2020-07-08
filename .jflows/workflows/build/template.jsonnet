local git_config = import '../config/git.libsonnet';
local steps = import '../common/steps.libsonnet';

local test_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
      steps.checkout,
      steps.setup_go,
      { name: 'build',
        run: 'go build' },
      { name: 'test',
        run: 'go test ./...' },
  ]
};

local workflow = {
  name: 'build',
  on: {
    pull_request: {
      branches: [git_config.main_branch]
    },
    push: {
      branches: [git_config.main_branch]
    }
  },
  jobs: {
    test: test_job
  },
};

std.manifestYamlDoc(workflow)

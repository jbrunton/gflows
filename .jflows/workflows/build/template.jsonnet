local git_config = import '../config/git.libsonnet';

local test_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
      { uses: 'actions/checkout@v2' },
      { uses: 'actions/setup-go@v2',
        with: { 'go-version': '^1.14.4' } },
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

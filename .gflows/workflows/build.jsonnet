local git_config = import 'config/git.libsonnet';
local steps = import 'common/steps.libsonnet';
local workflows = import 'common/workflows.libsonnet';

local test_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
      steps.checkout,
      steps.setup_go,
      {
        name: 'configure environment for pull request',
        'if': "github.event_name == 'pull_request'",
        env: {
          HEAD_SHA: '${{ github.event.pull_request.head.sha }}'
        },
        run: |||
          echo "::set-env name=GIT_BRANCH::$GITHUB_HEAD_REF"
          echo "::set-env name=GIT_COMMIT_SHA::$HEAD_SHA"
        |||
      },
      {
        name: 'configure environment for push',
        'if': "github.event_name == 'push'",
        env: {
          HEAD_SHA: '${{ github.event.pull_request.head.sha }}'
        },
        run: |||
          echo "::set-env name=GIT_BRANCH::${GITHUB_REF#refs/heads/}"
          echo "::set-env name=GIT_COMMIT_SHA::$GITHUB_SHA"
        |||
      },
      {
        name: 'prepare test reporter',
        run: |||
          curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > ./cc-test-reporter
          chmod +x ./cc-test-reporter
          ./cc-test-reporter before-build
        |||
      },
      steps.named('unit test', 'go test -coverprofile c.out $(go list ./... | grep -v /e2e)'),
      {
        name: 'upload coverage',
        env: {
          CC_TEST_REPORTER_ID: '${{ secrets.CC_TEST_REPORTER_ID }}'
        },
        run: './cc-test-reporter after-build --prefix github.com/jbrunton/gflows'
      },
      steps.named('e2e test', 'go test ./e2e')
  ]
};

local build_job = {
  'runs-on': 'ubuntu-latest',
  'steps': [
    steps.checkout,
    steps.setup_go,
    steps.run('go get github.com/rakyll/statik'),
    steps.run('make build'),
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

local git_config = import 'git.libsonnet';

{
  triggers: {
    pull_request_defaults: {
      pull_request: {
        branches: [git_config.main_branch]
      },
      push: {
        branches: [git_config.main_branch]
      }
    },
  },
}

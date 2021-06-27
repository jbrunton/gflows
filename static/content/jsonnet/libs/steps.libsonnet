{
  local steps = self,

  run(command):: {
    run: command
  },
  
  named(name, command):: steps.run(command) {
    name: name
  },

  uses(action):: {
    uses: action
  },

  checkout: steps.uses('actions/checkout@v2'),

  setup_gflows: steps.uses('jbrunton/setup-gflows@v1') {
    with: {
      token: "${{ secrets.GITHUB_TOKEN }}",
    }
  },

  check_workflows: steps.named('validate workflows', 'gflows check') {
    env: {
      GFLOWS_CONFIG: '$CONFIG_PATH'
    },
  },
}

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
  
  setup_go: steps.uses('actions/setup-go@v2') {
    with: {
      'go-version': '^1.14.4'
    }
  }
}

{
  checkout: {
    uses: 'actions/checkout@v2'
  },
  
  setup_go: {
    uses: 'actions/setup-go@v2',
    with: {
      'go-version': '^1.14.4'
    }
  },
  
  run:: function(command)
    { run: command },
  
  named:: function(name, command)
    { name: name, run: command },
}

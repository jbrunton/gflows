package engine

import (
	"fmt"

	"github.com/jbrunton/gflows/workflows"
)

const invalidJsonnetTemplate = `
local workflow = {
  on: {
    push: {
      branches: [ "develop" ],
    },
  }
};
std.manifestYamlDoc(workflow)
`

const invalidJsonnetWorkflow = `# File generated by gflows, do not modify
# Source: .gflows/workflows/test.jsonnet
"on":
  "push":
    "branches":
    - "develop"
`

const exampleJsonnetTemplate = `
local workflow = {
  on: {
    push: {
      branches: [ "develop" ],
    },
  },
	jobs: {
		test: {
			"runs-on": "ubuntu-latest",
			steps: [
			  { run: "echo Hello, World!" }
      ]
    }
  }
};
std.manifestYamlDoc(workflow)
`

func exampleWorkflow(sourceFileName string) string {
	return fmt.Sprintf(`# File generated by gflows, do not modify
# Source: .gflows/workflows/%s
"jobs":
  "test":
    "runs-on": "ubuntu-latest"
    "steps":
    - "run": "echo Hello, World!"
"on":
  "push":
    "branches":
    - "develop"
`, sourceFileName)
}

func newTestWorkflowDefinition(name string, content string) *workflows.WorkflowDefinition {
	json, _ := workflows.YamlToJson(content)
	return &workflows.WorkflowDefinition{
		Name:        name,
		Source:      fmt.Sprintf(".gflows/workflows/%s.jsonnet", name),
		Destination: fmt.Sprintf(".github/workflows/%s.yml", name),
		Content:     content,
		JSON:        json,
	}
}
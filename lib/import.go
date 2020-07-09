package lib

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jbrunton/jflows/jsonnet"
	"github.com/jbrunton/jflows/styles"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type GitWorkflow struct {
	path       string
	definition *WorkflowDefinition
}

func getWorkflows(fs *afero.Afero, context *JFlowsContext) []GitWorkflow {
	files := []string{}
	files, err := afero.Glob(fs, filepath.Join(context.GitHubDir, "workflows/*.yml"))
	if err != nil {
		panic(err)
	}

	definitions, err := GetWorkflowDefinitions(fs, context)
	if err != nil {
		panic(err) // TODO: improve handling
	}

	var workflows []GitWorkflow

	for _, file := range files {
		workflow := GitWorkflow{path: file}
		for _, definition := range definitions {
			if definition.Destination == file {
				workflow.definition = definition
				break
			}
		}
		workflows = append(workflows, workflow)
	}

	return workflows
}

func ImportWorkflows(fs *afero.Afero, context *JFlowsContext) {
	imported := 0
	workflows := getWorkflows(fs, context)
	for _, workflow := range workflows {
		fmt.Println("Found workflow:", workflow.path)
		if workflow.definition == nil {
			workflowContent, err := fs.ReadFile(workflow.path)
			if err != nil {
				panic(err)
			}
			var yamlData map[interface{}]interface{}
			err = yaml.Unmarshal(workflowContent, &yamlData)
			if err != nil {
				panic(err)
			}

			jsonData, err := convertToStringKeysRecursive(yamlData, "")
			if err != nil {
				panic(err)
			}

			json, err := jsonnet.Marshal(jsonData)
			if err != nil {
				panic(err)
			}

			templateContent := fmt.Sprintf("local workflow = %s;\n\nstd.manifestYamlDoc(workflow)\n", string(json))

			_, filename := filepath.Split(workflow.path)
			templateName := strings.TrimSuffix(filename, filepath.Ext(filename))
			templatePath := filepath.Join(context.WorkflowsDir, templateName, "template.jsonnet")
			safelyWriteFile(fs, templatePath, templateContent)
			fmt.Println("  Imported template:", templatePath)
			imported++
		} else {
			fmt.Println("  Exists:", workflow.definition.Source)
		}
	}

	if imported > 0 {
		fmt.Println()
		fmt.Println(styles.StyleWarning("Important:"), "imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.")
		fmt.Println("  ► Run \"jflows workflow update\" to do this now")
	}
}

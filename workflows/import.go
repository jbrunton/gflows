package workflows

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/jsonnet"
	"github.com/jbrunton/gflows/styles"

	"gopkg.in/yaml.v2"
)

func (manager *WorkflowManager) ImportWorkflows() {
	imported := 0
	workflows := manager.GetWorkflows()
	for _, workflow := range workflows {
		fmt.Println("Found workflow:", workflow.path)
		if workflow.definition == nil {
			workflowContent, err := manager.fs.ReadFile(workflow.path)
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
			templatePath := filepath.Join(manager.context.WorkflowsDir, templateName+".jsonnet")
			manager.contentWriter.SafelyWriteFile(templatePath, templateContent)
			fmt.Println("  Imported template:", templatePath)
			imported++
		} else {
			fmt.Println("  Exists:", workflow.definition.Source)
		}
	}

	if imported > 0 {
		fmt.Println()
		fmt.Println(styles.StyleWarning("Important:"), "imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.")
		fmt.Println("  ► Run \"gflows update\" to do this now")
	}
}

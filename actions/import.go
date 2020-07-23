package actions

import (
	"fmt"
)

func (manager *WorkflowManager) ImportWorkflows() error {
	imported := 0
	workflows := manager.GetWorkflows()
	for _, workflow := range workflows {
		manager.logger.Println("Found workflow:", workflow.Path)
		if workflow.Definition == nil {
			templatePath, err := manager.ImportWorkflow(&workflow)
			if err != nil {
				return err
			}
			manager.logger.Println("  Imported template:", templatePath)
			imported++
		} else {
			manager.logger.Println("  Exists:", workflow.Definition.Source)
		}
	}

	if imported > 0 {
		fmt.Println()
		fmt.Println(manager.styles.StyleWarning("Important:"), "imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.")
		fmt.Println("  ► Run \"gflows update\" to do this now")
	}
	return nil
}

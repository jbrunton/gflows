package action

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
		manager.logger.Println()
		manager.logger.Println(manager.styles.StyleWarning("Important:"), "imported workflow templates may generate yaml which is ordered differerently from the source. You will need to update the workflows before validation passes.")
		manager.logger.Println("  ► Run \"gflows update\" to do this now")
	}
	return nil
}

package workflows

import (
	"fmt"

	"github.com/jbrunton/gflows/config"
)

type TemplateManager struct {
	context          *config.GFlowsContext
	defaultEngine    string
	engines          map[string]TemplateEngine
	sourcesCache     map[string]*[]string
	templatesCache   map[string]*[]string
	definitionsCache map[string]*[]*WorkflowDefinition
}

func (manager *TemplateManager) GetWorkflowSources() []string {
	return manager.getWorkflowSourcesForEngine(manager.defaultEngine)
}

func (manager *TemplateManager) GetWorkflowTemplates() []string {
	return manager.getWorkflowTemplatesForEngine(manager.defaultEngine)
}

func (manager *TemplateManager) GetWorkflowDefinitions() ([]*WorkflowDefinition, error) {
	return manager.getWorkflowDefinitionsForEngine(manager.defaultEngine)
}

func (manager *TemplateManager) getWorkflowSourcesForEngine(engine string) []string {
	if manager.sourcesCache[engine] == nil {
		sources := manager.engines[engine].GetWorkflowSources()
		manager.sourcesCache[engine] = &sources
	}
	return *manager.sourcesCache[engine]
}

func (manager *TemplateManager) getWorkflowTemplatesForEngine(engine string) []string {
	if manager.templatesCache[engine] == nil {
		templates := manager.engines[engine].GetWorkflowTemplates()
		manager.templatesCache[engine] = &templates
	}
	return *manager.sourcesCache[engine]
}

func (manager *TemplateManager) getWorkflowDefinitionsForEngine(engine string) ([]*WorkflowDefinition, error) {
	if manager.definitionsCache[engine] == nil {
		definitions, err := manager.engines[engine].GetWorkflowDefinitions()
		if err != nil {
			return []*WorkflowDefinition{}, err
		}
		fmt.Println("engine:", engine)
		fmt.Printf("definitions:%+v\n", definitions)
		manager.definitionsCache[engine] = &definitions
	}
	return *manager.definitionsCache[engine], nil
}

func (manager *TemplateManager) getWorkflowDefinition(workflowName string, engine string) (*WorkflowDefinition, error) {
	definitions, err := manager.getWorkflowDefinitionsForEngine(engine)
	if err != nil {
		return nil, err
	}
	for _, definition := range definitions {
		if definition.Name == workflowName {
			return definition, nil
		}
	}
	return nil, nil
}

func NewTemplateManager(context *config.GFlowsContext, jsonnetEngine *JsonnetTemplateEngine, yttEngine *YttTemplateEngine) *TemplateManager {
	return &TemplateManager{
		context: context,
		engines: map[string]TemplateEngine{
			"jsonnet": jsonnetEngine,
			"ytt":     yttEngine,
		},
		defaultEngine:    context.Config.Templates.Defaults.Engine,
		sourcesCache:     make(map[string]*[]string),
		templatesCache:   make(map[string]*[]string),
		definitionsCache: make(map[string]*[]*WorkflowDefinition),
	}
}

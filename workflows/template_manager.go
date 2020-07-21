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

type TemplateOverride struct {
	workflowName   string
	templateConfig *config.GFlowsTemplateConfig
}

func (manager *TemplateManager) GetWorkflowSources() []string {
	return manager.getWorkflowSourcesForEngine(manager.defaultEngine)
}

func (manager *TemplateManager) GetWorkflowTemplates() []string {
	templates := manager.getWorkflowTemplatesForEngine(manager.defaultEngine)
	overrides := manager.getTemplateEngineOverrides()
	for _, override := range overrides {
		definition, err := manager.getWorkflowDefinition(override.workflowName, override.templateConfig.Engine)
		if err != nil {
			panic(err)
		}
		templates = append(templates, definition.Source)
	}
	return templates
}

func (manager *TemplateManager) GetWorkflowDefinitions() ([]*WorkflowDefinition, error) {
	return manager.getWorkflowDefinitionsForEngine(manager.defaultEngine)
}

func (manager *TemplateManager) getWorkflowSourcesForEngine(engine string) []string {
	manager.validateEngine(engine)
	if manager.sourcesCache[engine] == nil {
		sources := manager.engines[engine].GetWorkflowSources()
		manager.sourcesCache[engine] = &sources
	}
	return *manager.sourcesCache[engine]
}

func (manager *TemplateManager) getWorkflowTemplatesForEngine(engine string) []string {
	manager.validateEngine(engine)
	if manager.templatesCache[engine] == nil {
		templates := manager.engines[engine].GetWorkflowTemplates()
		manager.templatesCache[engine] = &templates
	}
	return *manager.templatesCache[engine]
}

func (manager *TemplateManager) getWorkflowDefinitionsForEngine(engine string) ([]*WorkflowDefinition, error) {
	manager.validateEngine(engine)
	if manager.definitionsCache[engine] == nil {
		definitions, err := manager.engines[engine].GetWorkflowDefinitions()
		if err != nil {
			return []*WorkflowDefinition{}, err
		}
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

func (manager *TemplateManager) getTemplateEngineOverrides() []TemplateOverride {
	var overrides []TemplateOverride
	for workflowName, override := range manager.context.Config.Templates.Overrides {
		if override.Engine != "" {
			overrides = append(overrides, TemplateOverride{workflowName: workflowName, templateConfig: override})
		}
	}
	return overrides
}

func (manager *TemplateManager) validateEngine(engine string) {
	for candidateEngine, _ := range manager.engines {
		if candidateEngine == engine {
			return
		}
	}

	panic(fmt.Errorf("Unexpected engine: %q", engine))
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

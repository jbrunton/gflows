package config

// GFlowsPackage - represents a package which may contain workflows or library files. Implemented
// by GFlowsLib for library packages, and GFlowsContext for the given context.
type GFlowsPackage interface {
	WorkflowsPath() string
	LibPath() string
	// GetPathInfo(localPath string)
}

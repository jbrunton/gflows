package pkg

type GFlowsPackage interface {
	WorkflowsDir() string
	LibsDir() string
	GetPathInfo(localPath string) (*PathInfo, error)
}

module github.com/jbrunton/gflows

go 1.14

require (
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-git/go-git v4.7.0+incompatible
	github.com/go-git/go-git/v5 v5.1.0
	github.com/google/go-jsonnet v0.16.0
	github.com/inancgumus/screen v0.0.0-20190314163918-06e984b86ed3
	github.com/k14s/ytt v0.28.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.4
	github.com/rakyll/statik v0.1.7
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.4.0
	github.com/xeipuuv/gojsonschema v1.2.0
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible
)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200522161834-8a7b2030a110

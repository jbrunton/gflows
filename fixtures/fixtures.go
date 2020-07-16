package fixtures

import (
	"archive/zip"
	"bytes"
	"net/http"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/di"
	"github.com/jbrunton/gflows/fs"
	"github.com/jbrunton/gflows/logs"
	statikFs "github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
)

type File struct {
	Name string
	Body string
}

func CreateTestFileSystem(files []File, assetNamespace string) http.FileSystem {
	out := new(bytes.Buffer)
	writer := zip.NewWriter(out)
	for _, file := range files {
		f, err := writer.Create(file.Name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}
	err := writer.Close()
	if err != nil {
		panic(err)
	}
	asset := out.String()
	statikFs.RegisterWithNamespace(assetNamespace, asset)
	sourceFs, err := statikFs.NewWithNamespace(assetNamespace)
	if err != nil {
		panic(err)
	}
	return sourceFs
}

func NewTestContext(cmd *cobra.Command, configString string) (*di.Container, *bytes.Buffer) {
	out := new(bytes.Buffer)
	fs := fs.CreateMemFs()
	fs.WriteFile(".gflows/config.yml", []byte(configString), 0644)
	context, _ := config.GetContext(fs, cmd)
	return di.BuildContainer(fs, logs.NewLogger(out), context), out
}

func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("config", "", "")
	return cmd
}

// func NewTestContainer(context *config.GFlowsContext) (*di.Container, *bytes.Buffer) {
// 	out := new(bytes.Buffer)
// 	return di.BuildContainer(fs.CreateMemFs(), logs.NewLogger(out), context), out
// }

package fixtures

import (
	"archive/zip"
	"bytes"
	"net/http"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
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

func NewTestContext(configString string) (*adapters.Container, *config.GFlowsContext, *bytes.Buffer) {
	fs := adapters.CreateMemFs()
	out := new(bytes.Buffer)
	container := adapters.NewContainer(fs, adapters.NewLogger(out))

	configPath := ".gflows/config.yml"
	fs.WriteFile(configPath, []byte(configString), 0644)
	context, _ := config.NewContext(fs, configPath)

	return container, context, out
}

func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("config", "", "")
	return cmd
}

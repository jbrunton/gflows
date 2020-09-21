/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jbrunton/gflows/env"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCommand(buildContainer)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(aurora.Red(err.Error()).Bold())
		os.Exit(1)
	}
}

// Version - the build version
var Version = "development"

func newVersionCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}
			container.Logger().Printfln("gflows version %s", Version)
			return nil
		},
	}
	return cmd
}

// NewRootCommand creates a new root command
func NewRootCommand(containerFunc ContainerBuilderFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gflows",
		Short:         "Generate GitHub workflows from jsonnet templates",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			env.CleanUpLibs()
		},
	}
	cmd.PersistentFlags().StringP("config", "c", "", "Location of config file")
	cmd.PersistentFlags().Bool("disable-colors", false, "Disable colors in output")
	cmd.PersistentFlags().BoolP("debug", "d", false, "Print debug information")

	cmd.AddCommand(newListWorkflowsCmd(containerFunc))
	cmd.AddCommand(newUpdateWorkflowsCmd(containerFunc))
	cmd.AddCommand(newCheckWorkflowsCmd(containerFunc))
	cmd.AddCommand(newWatchWorkflowsCmd(containerFunc))
	cmd.AddCommand(newImportWorkflowsCmd(containerFunc))
	cmd.AddCommand(newInitCmd(containerFunc))
	cmd.AddCommand(newVersionCmd(containerFunc))

	return cmd
}

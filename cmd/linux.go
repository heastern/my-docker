package cmd

import (
	"linux/container"
	"linux/namespace"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "app",
		Short: "测试 Linux 隔离机制",
		Long:  "测试 Linux 隔离机制",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(namespace.PIDCommand())
	rootCmd.AddCommand(namespace.CGroupsCommand())
	rootCmd.AddCommand(namespace.MountCommand())
	rootCmd.AddCommand(container.ContainerCommand()...)
}

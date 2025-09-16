package cmd

import (
	"fmt"
	"pnas/internal/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "harborark",
	Short: "HarborArk is a NAS open source system with gin",
	Long:  "HarborArk系统命令行工具，用于管理HarborArk系统",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  "显示HarborArk系统命令行工具的版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HarborArk系统命令行工具版本：" + config.AppConfig.Server.Version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

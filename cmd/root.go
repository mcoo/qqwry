/*
Copyright © 2023 enjoy <i@mcenjoy.cn>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qqwry",
	Short: "集合对纯真数据库的自动更新及查询功能",
	Long:  `集合对纯真数据库的自动更新及查询功能`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("path", "p", "qqwry.dat", "qqwry.dat 文件路径")
}

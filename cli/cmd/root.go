package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{Use: "rune", Short: "Remote runtime for autonomous agents"}

func Execute() { _ = rootCmd.Execute() }

func init() { rootCmd.AddCommand(runCmd) }

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildVersionCmd creates the version command
func buildVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Display the version number of 0perator.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("0perator %s\n", Version)
		},
	}
}

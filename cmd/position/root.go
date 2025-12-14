// ABOUTME: Root Cobra command and global flags
// ABOUTME: Sets up CLI structure and database connection

package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "position",
	Short: "Simple location tracking for items",
	Long: `Position tracks items (people, things) and their locations over time.

Examples:
  position add harper 41.8781 -87.6298 --label chicago
  position current harper
  position timeline harper
  position list`,
}

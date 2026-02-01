// ABOUTME: Root Cobra command and global flags
// ABOUTME: Sets up CLI structure and SQLite database connection

package main

import (
	"fmt"

	"github.com/harper/position/internal/storage"
	"github.com/spf13/cobra"
)

var db *storage.SQLiteDB

var rootCmd = &cobra.Command{
	Use:   "position",
	Short: "Simple location tracking for items",
	Long: `
██████╗  ██████╗ ███████╗██╗████████╗██╗ ██████╗ ███╗   ██╗
██╔══██╗██╔═══██╗██╔════╝██║╚══██╔══╝██║██╔═══██╗████╗  ██║
██████╔╝██║   ██║███████╗██║   ██║   ██║██║   ██║██╔██╗ ██║
██╔═══╝ ██║   ██║╚════██║██║   ██║   ██║██║   ██║██║╚██╗██║
██║     ╚██████╔╝███████║██║   ██║   ██║╚██████╔╝██║ ╚████║
╚═╝      ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝

         Track items and their locations over time

Examples:
  position add harper --lat 41.8781 --lng -87.6298 --label chicago
  position current harper
  position timeline harper
  position list`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		db, err = storage.NewSQLiteDB(storage.DefaultDBPath())
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
		}
		return nil
	},
}

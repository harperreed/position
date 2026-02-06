// ABOUTME: Root Cobra command and global flags
// ABOUTME: Sets up CLI structure and storage backend connection via config

package main

import (
	"fmt"

	"github.com/harper/position/internal/config"
	"github.com/harper/position/internal/storage"
	"github.com/spf13/cobra"
)

var db storage.Repository

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
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		db, err = cfg.OpenStorage()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
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

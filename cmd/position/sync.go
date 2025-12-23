// ABOUTME: Sync subcommand for Charm cloud sync
// ABOUTME: Provides status, link, unlink, and wipe commands

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/charm/client"
	"github.com/charmbracelet/charm/kv"
	"github.com/fatih/color"
	"github.com/harper/position/internal/charm"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manage cloud sync for position data",
	Long: `Sync your position data with Charm Cloud using SSH key authentication.

Commands:
  status  - Show sync status and user info
  link    - Link this device to your Charm account
  unlink  - Unlink this device from your account
  repair  - Repair corrupted database (checkpoint WAL, check integrity, vacuum)
  reset   - Reset local database from cloud (discards local changes)
  wipe    - Permanently delete all data (local and cloud)

Data syncs automatically on every write operation.

Examples:
  position sync status
  position sync link
  position sync repair --force
  position sync reset
  position sync wipe`,
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	Long:  `Display current sync configuration, user ID, and connection status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := charm.DefaultConfig()

		fmt.Printf("Charm Host: %s\n", cfg.CharmHost)
		fmt.Printf("Database:   %s\n", charm.DBName)

		// Get user info from charm client
		cc, err := client.NewClientWithDefaults()
		if err != nil {
			color.Yellow("\nStatus: Not connected")
			fmt.Println("Run 'position sync link' to connect your account.")
			return nil
		}

		user, err := cc.ID()
		if err != nil {
			color.Yellow("\nStatus: Not linked")
			fmt.Println("Run 'position sync link' to connect your account.")
			return nil
		}

		fmt.Printf("\nUser ID: %s\n", user)
		color.Green("Status: Connected")
		fmt.Println("\nData syncs automatically on every write.")

		return nil
	},
}

var syncLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link this device to your Charm account",
	Long: `Link this device to your Charm Cloud account.

This opens the Charm linking flow which authenticates using SSH keys.
If you don't have an account, one will be created automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting Charm link process...")
		fmt.Println("This will open an interactive linking flow.")
		fmt.Println()

		// Run charm link command
		linkCmd := exec.Command("charm", "link")
		linkCmd.Stdin = os.Stdin
		linkCmd.Stdout = os.Stdout
		linkCmd.Stderr = os.Stderr

		if err := linkCmd.Run(); err != nil {
			// charm might not be installed
			return fmt.Errorf("failed to run 'charm link': %w\nMake sure the charm CLI is installed: go install github.com/charmbracelet/charm@latest", err)
		}

		color.Green("\n✓ Device linked successfully")
		fmt.Println("Position data will now sync automatically.")

		return nil
	},
}

var syncUnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink this device from your Charm account",
	Long: `Unlink this device from your Charm Cloud account.

This will stop syncing data but won't delete local data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Unlinking device from Charm...")

		// Run charm unlink command
		unlinkCmd := exec.Command("charm", "unlink")
		unlinkCmd.Stdin = os.Stdin
		unlinkCmd.Stdout = os.Stdout
		unlinkCmd.Stderr = os.Stderr

		if err := unlinkCmd.Run(); err != nil {
			return fmt.Errorf("failed to run 'charm unlink': %w", err)
		}

		color.Green("\n✓ Device unlinked")
		fmt.Println("Local data is preserved. Sync is disabled.")

		return nil
	},
}

var (
	repairForce bool
)

var syncRepairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair corrupted database",
	Long: `Attempt to repair a corrupted position database.

Steps performed:
  1. Checkpoint WAL (merge pending writes)
  2. Remove stale SHM file
  3. Run integrity check
  4. Vacuum database

If --force is specified and integrity check fails:
  5. Attempt REINDEX recovery
  6. Reset from cloud as last resort

This is useful when:
- Database is corrupted or locked
- WAL files are out of sync
- You see "database disk image is malformed" errors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Repairing position database...")
		fmt.Println()

		result, err := kv.Repair(charm.DBName, repairForce)
		if err != nil && !repairForce {
			color.Red("✗ Repair failed: %v", err)
			fmt.Println("\nRun with --force to attempt recovery:")
			fmt.Println("  position sync repair --force")
			return err
		} else if err != nil {
			color.Red("✗ Repair failed even with --force: %v", err)
			return err
		}

		// Display results
		fmt.Println("Repair results:")
		if result.WalCheckpointed {
			color.Green("  ✓ WAL checkpointed")
		}
		if result.ShmRemoved {
			color.Green("  ✓ SHM file removed")
		}
		if result.IntegrityOK {
			color.Green("  ✓ Integrity check passed")
		} else {
			color.Red("  ✗ Integrity check failed")
		}
		if result.Vacuumed {
			color.Green("  ✓ Database vacuumed")
		}
		if result.RecoveryAttempted {
			color.Yellow("  ⚠ Recovery attempted (REINDEX)")
		}
		if result.ResetFromCloud {
			color.Yellow("  ⚠ Reset from cloud")
		}
		if result.Error != nil {
			color.Yellow("  ⚠ Warning: %v", result.Error)
		}

		fmt.Println()
		color.Green("✓ Repair completed")
		return nil
	},
}

var syncResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset local database from cloud",
	Long: `Delete local database and pull fresh data from Charm Cloud.

This discards any unsynced local changes.

This is useful when:
- You want to pull a clean copy from the cloud
- Local database is corrupted beyond repair
- You need to force sync from cloud

WARNING: Any local changes not yet synced to cloud will be lost.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will DELETE your local database and pull fresh data from the cloud.")
		color.Yellow("WARNING: Any unsynced local changes will be lost.")
		fmt.Print("\nContinue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.ToLower(strings.TrimSpace(confirmation))

		if confirmation != "y" && confirmation != "yes" {
			fmt.Println("Aborted.")
			return nil
		}

		fmt.Println("\nResetting database...")

		if err := kv.Reset(charm.DBName); err != nil {
			return fmt.Errorf("failed to reset: %w", err)
		}

		color.Green("✓ Database reset from cloud")
		fmt.Println("Your data has been refreshed from the cloud.")

		return nil
	},
}

var syncWipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Permanently delete all data (local and cloud)",
	Long: `Permanently delete ALL position data, both local and on Charm Cloud.

This is DESTRUCTIVE and CANNOT be undone.

This is useful when:
- You want to completely remove all position data
- You're decommissioning your account
- You need to start completely fresh

WARNING: This deletes data from ALL linked devices, not just this one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will PERMANENTLY DELETE all position data.")
		color.Red("WARNING: This deletes data from ALL linked devices and cloud backups.")
		color.Red("WARNING: This action CANNOT be undone.")
		fmt.Print("\nType 'wipe' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)

		if confirmation != "wipe" {
			fmt.Println("Aborted.")
			return nil
		}

		fmt.Println("\nWiping all data...")

		result, err := kv.Wipe(charm.DBName)
		if err != nil {
			return fmt.Errorf("failed to wipe: %w", err)
		}

		// Display results
		fmt.Println()
		if result.CloudBackupsDeleted > 0 {
			color.Green("✓ Deleted %d cloud backup(s)", result.CloudBackupsDeleted)
		}
		if result.LocalFilesDeleted > 0 {
			color.Green("✓ Deleted %d local file(s)", result.LocalFilesDeleted)
		}
		if result.Error != nil {
			color.Yellow("⚠ Warning: %v", result.Error)
		}

		fmt.Println()
		color.Green("✓ All data wiped")
		fmt.Println("Run 'position add' to start tracking again.")

		return nil
	},
}

func init() {
	// Add --force flag to repair command
	syncRepairCmd.Flags().BoolVarP(&repairForce, "force", "f", false, "Force recovery even if integrity check fails")

	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncLinkCmd)
	syncCmd.AddCommand(syncUnlinkCmd)
	syncCmd.AddCommand(syncRepairCmd)
	syncCmd.AddCommand(syncResetCmd)
	syncCmd.AddCommand(syncWipeCmd)

	rootCmd.AddCommand(syncCmd)
}

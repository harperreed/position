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
  wipe    - Clear all local data and start fresh

Data syncs automatically on every write operation.

Examples:
  position sync status
  position sync link
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

var syncWipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Wipe all local data and start fresh",
	Long: `Clear all local position data and reset the KV store.

This is useful when:
- You want to start fresh
- Local data became corrupted
- You're changing accounts

This does NOT delete data from other linked devices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will DELETE all local position data.")
		fmt.Println("Data on other linked devices will NOT be affected.")
		fmt.Print("\nType 'wipe' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)

		if confirmation != "wipe" {
			fmt.Println("Aborted.")
			return nil
		}

		fmt.Println("\nWiping local data...")

		if err := charmClient.Reset(); err != nil {
			return fmt.Errorf("failed to reset: %w", err)
		}

		color.Green("✓ Local data wiped")
		fmt.Println("Run 'position add' to start tracking again.")

		return nil
	},
}

func init() {
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncLinkCmd)
	syncCmd.AddCommand(syncUnlinkCmd)
	syncCmd.AddCommand(syncWipeCmd)

	rootCmd.AddCommand(syncCmd)
}

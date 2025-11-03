package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/xerudro/DASHBOARD-v2/internal/vault"
)

var (
	dbDSN     string
	masterKey string
	v         *vault.Vault
)

// parseUserID parses a user ID string to UUID
func parseUserID(userIDStr string) (uuid.UUID, error) {
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID is required")
	}
	return uuid.Parse(userIDStr)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "vaultctl",
		Short: "VIP Hosting Panel Secrets Vault CLI",
		Long:  "Command-line interface for managing secrets in the VIP Hosting Panel vault",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip vault initialization for version command
			if cmd.Name() == "version" {
				return nil
			}

			// Connect to database
			db, err := sqlx.Connect("postgres", dbDSN)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}

			// Create vault instance
			config := vault.DefaultVaultConfig()
			v, err = vault.NewVault(db, config)
			if err != nil {
				return fmt.Errorf("failed to create vault: %w", err)
			}

			// Unlock vault if master key provided
			if masterKey != "" {
				if err := v.Unlock(masterKey); err != nil {
					return fmt.Errorf("failed to unlock vault: %w", err)
				}
			} else {
				// Try environment variable
				if err := v.UnlockFromEnv(); err != nil {
					fmt.Println("Warning: Vault is locked. Use --master-key or set VAULT_MASTER_KEY")
				}
			}

			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&dbDSN, "db-dsn", os.Getenv("DATABASE_URL"), "Database connection string")
	rootCmd.PersistentFlags().StringVar(&masterKey, "master-key", "", "Vault master key (or set VAULT_MASTER_KEY)")

	// Add subcommands
	rootCmd.AddCommand(
		unlockCmd(),
		lockCmd(),
		statusCmd(),
		createCmd(),
		getCmd(),
		updateCmd(),
		deleteCmd(),
		listCmd(),
		versionsCmd(),
		rotateCmd(),
		auditCmd(),
		cleanupCmd(),
		generateCmd(),
		versionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func unlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Unlock the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			if masterKey == "" {
				return fmt.Errorf("master key is required")
			}

			if err := v.Unlock(masterKey); err != nil {
				return err
			}

			fmt.Println("✅ Vault unlocked successfully")
			return nil
		},
	}
}

func lockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			v.Lock()
			fmt.Println("🔒 Vault locked")
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show vault status",
		RunE: func(cmd *cobra.Command, args []string) error {
			health := v.Health()
			data, _ := json.MarshalIndent(health, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func createCmd() *cobra.Command {
	var (
		path        string
		value       string
		description string
		expiresIn   string
		userID      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			var expiresInDuration *time.Duration
			if expiresIn != "" {
				duration, err := time.ParseDuration(expiresIn)
				if err != nil {
					return fmt.Errorf("invalid expiration duration: %w", err)
				}
				expiresInDuration = &duration
			}

			// Parse user ID
			parsedUserID, err := parseUserID(userID)
			if err != nil {
				return err
			}

			err = v.CreateSecret(context.Background(), path, value, description, parsedUserID, expiresInDuration)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Secret created at path: %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.Flags().StringVar(&value, "value", "", "Secret value (required)")
	cmd.Flags().StringVar(&description, "description", "", "Secret description")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "", "Expiration duration (e.g., 24h, 7d, 30d)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID (UUID) creating the secret")

	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("value")

	return cmd
}

func getCmd() *cobra.Command {
	var (
		path   string
		userID string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a secret value",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedUserID, err := parseUserID(userID)
			if err != nil {
				return err
			}

			value, err := v.GetSecret(context.Background(), path, parsedUserID, "cli")
			if err != nil {
				return err
			}

			fmt.Printf("Path: %s\n", path)
			fmt.Printf("Value: %s\n", value)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID (UUID) retrieving the secret")

	cmd.MarkFlagRequired("path")

	return cmd
}

func updateCmd() *cobra.Command {
	var (
		path   string
		value  string
		userID string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a secret value",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedUserID, err := parseUserID(userID)
			if err != nil {
				return err
			}

			err = v.UpdateSecret(context.Background(), path, value, parsedUserID)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Secret updated at path: %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.Flags().StringVar(&value, "value", "", "New secret value (required)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID (UUID) updating the secret")

	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("value")

	return cmd
}

func deleteCmd() *cobra.Command {
	var (
		path   string
		userID string
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedUserID, err := parseUserID(userID)
			if err != nil {
				return err
			}

			err = v.DeleteSecret(context.Background(), path, parsedUserID)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Secret deleted at path: %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID (UUID) deleting the secret")

	cmd.MarkFlagRequired("path")

	return cmd
}

func listCmd() *cobra.Command {
	var prefix string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := v.ListSecrets(context.Background(), prefix)
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets found")
				return nil
			}

			fmt.Printf("Found %d secret(s):\n\n", len(secrets))
			for _, secret := range secrets {
				fmt.Printf("Path: %s\n", secret.Path)
				fmt.Printf("  Description: %s\n", secret.Description)
				fmt.Printf("  Version: %d\n", secret.Version)
				fmt.Printf("  Created: %s\n", secret.CreatedAt.Format(time.RFC3339))
				fmt.Printf("  Updated: %s\n", secret.UpdatedAt.Format(time.RFC3339))
				if secret.ExpiresAt != nil {
					fmt.Printf("  Expires: %s\n", secret.ExpiresAt.Format(time.RFC3339))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&prefix, "prefix", "", "Path prefix to filter")

	return cmd
}

func versionsCmd() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List secret versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			versions, err := v.ListSecretVersions(context.Background(), path)
			if err != nil {
				return err
			}

			if len(versions) == 0 {
				fmt.Println("No versions found")
				return nil
			}

			fmt.Printf("Found %d version(s):\n\n", len(versions))
			for _, version := range versions {
				fmt.Printf("Version: %d\n", version.Version)
				fmt.Printf("  Updated: %s\n", version.UpdatedAt.Format(time.RFC3339))
				fmt.Printf("  Updated By: User ID %d\n", version.UpdatedBy)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.MarkFlagRequired("path")

	return cmd
}

func rotateCmd() *cobra.Command {
	var (
		path         string
		newMasterKey string
		userID       string
		rotateAll    bool
	)

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate secret(s) with new master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedUserID, err := parseUserID(userID)
			if err != nil {
				return err
			}

			if rotateAll {
				err := v.RotateAllSecrets(context.Background(), newMasterKey, parsedUserID)
				if err != nil {
					return err
				}
				fmt.Println("✅ All secrets rotated successfully")
			} else {
				err := v.RotateSecret(context.Background(), path, newMasterKey, parsedUserID)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Secret rotated at path: %s\n", path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required unless --all)")
	cmd.Flags().StringVar(&newMasterKey, "new-master-key", "", "New master key (required)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID (UUID) performing rotation")
	cmd.Flags().BoolVar(&rotateAll, "all", false, "Rotate all secrets")

	cmd.MarkFlagRequired("new-master-key")

	return cmd
}

func auditCmd() *cobra.Command {
	var (
		path  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "View audit logs for a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			logs, err := v.GetAuditLogs(context.Background(), path, limit)
			if err != nil {
				return err
			}

			if len(logs) == 0 {
				fmt.Println("No audit logs found")
				return nil
			}

			fmt.Printf("Found %d audit log(s):\n\n", len(logs))
			for _, log := range logs {
				status := "✅"
				if !log.Success {
					status = "❌"
				}
				fmt.Printf("%s %s - User ID %d - %s - %s\n",
					status,
					log.Timestamp.Format(time.RFC3339),
					log.UserID,
					log.Action,
					log.IPAddress,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of logs to retrieve")

	cmd.MarkFlagRequired("path")

	return cmd
}

func cleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup expired secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			count, err := v.CleanupExpiredSecrets(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("✅ Cleaned up %d expired secret(s)\n", count)
			return nil
		},
	}
}

func generateCmd() *cobra.Command {
	var length int

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a secure random token",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := v.GenerateSecureToken(length)
			if err != nil {
				return err
			}

			fmt.Printf("Generated token (%d bytes):\n%s\n", length, token)
			return nil
		},
	}

	cmd.Flags().IntVar(&length, "length", 32, "Token length in bytes")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show vaultctl version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("vaultctl version 1.0.0")
			fmt.Println("VIP Hosting Panel Secrets Vault CLI")
		},
	}
}

package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	snapshotID     string
	repoURL        string
	awsAccessKey   string
	awsSecretKey   string
	resticPassword string
	s3Endpoint     string
	bucketName     string
	bucketPath     string
	s3PathStyle    bool
	outputPath     string
	includeGlobal  bool
	excludeConfig  bool
	dryRun         bool
	listSnapshots  bool
	snapshotsLimit int
	snapshotsPage  int
)

// ResticSnapshot represents the JSON output from restic snapshots command
type ResticSnapshot struct {
	Time           time.Time `json:"time"`
	Parent         string    `json:"parent,omitempty"`
	Tree           string    `json:"tree"`
	Paths          []string  `json:"paths"`
	Hostname       string    `json:"hostname"`
	Username       string    `json:"username"`
	UID            int       `json:"uid"`
	GID            int       `json:"gid"`
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	ProgramVersion string    `json:"program_version"`
	Summary        struct {
		TotalFilesProcessed int   `json:"total_files_processed"`
		TotalBytesProcessed int64 `json:"total_bytes_processed"`
	} `json:"summary,omitempty"`
}

// NewRestoreCmd returns a cobra command for the backup restore operation
func NewRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a backup of ChainLaunch",
		Long: `Restore a ChainLaunch instance from a backup.
This command allows restoring from a specific restic snapshot to a designated location.
`,
		RunE: runRestore,
	}

	cmd.Flags().StringVar(&snapshotID, "snapshot-id", "latest", "ID of the restic snapshot to restore (defaults to latest snapshot)")
	cmd.Flags().StringVar(&repoURL, "repo-url", "", "Restic repository URL (e.g. s3:https://s3.amazonaws.com/bucket-name/path)")
	cmd.Flags().StringVar(&awsAccessKey, "aws-access-key", "", "AWS Access Key ID for S3 access")
	cmd.Flags().StringVar(&awsSecretKey, "aws-secret-key", "", "AWS Secret Access Key for S3 access")
	cmd.Flags().StringVar(&resticPassword, "restic-password", "", "Password for the restic repository")
	cmd.Flags().StringVar(&s3Endpoint, "s3-endpoint", "", "S3 endpoint URL")
	cmd.Flags().StringVar(&bucketName, "bucket-name", "", "S3 bucket name")
	cmd.Flags().StringVar(&bucketPath, "bucket-path", "", "Path within the S3 bucket")
	cmd.Flags().BoolVar(&s3PathStyle, "s3-path-style", false, "Use S3 path style addressing")
	cmd.Flags().StringVar(&outputPath, "output", "", "Path where to restore the backup (default: .chainlaunch in home directory)")
	cmd.Flags().BoolVar(&includeGlobal, "include-global", false, "Include global configuration in the restore")
	cmd.Flags().BoolVar(&excludeConfig, "exclude-config", false, "Exclude configuration files from restore")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without performing the actual restore")
	cmd.Flags().BoolVar(&listSnapshots, "list-snapshots", false, "List available snapshots instead of performing a restore")
	cmd.Flags().IntVar(&snapshotsLimit, "limit", 10, "Number of snapshots to list per page (default: 10)")
	cmd.Flags().IntVar(&snapshotsPage, "page", 1, "Page number for snapshot listing (default: 1)")

	return cmd
}

func runRestore(cmd *cobra.Command, cmdArgs []string) error {
	ctx := context.Background()

	// Setup logger
	log := logger.NewDefault()
	log.Info("Starting backup restore operation")

	// Set default output path if not specified
	if outputPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, ".chainlaunch")
	}

	// Validate required flags for S3 access
	if repoURL == "" && (s3Endpoint == "" || bucketName == "") {
		return fmt.Errorf("either --repo-url or both --s3-endpoint and --bucket-name must be specified")
	}

	// Construct repository URL if not provided directly
	if repoURL == "" {
		endpointURL := s3Endpoint
		if !strings.HasPrefix(endpointURL, "http://") && !strings.HasPrefix(endpointURL, "https://") {
			endpointURL = "https://" + endpointURL
		}

		path := strings.TrimPrefix(bucketPath, "/")
		repoURL = fmt.Sprintf("s3:%s/%s/%s",
			strings.TrimSuffix(endpointURL, "/"),
			bucketName,
			path)
	}

	log.Info("Using restic repository", "url", repoURL)

	// Prepare environment variables for restic
	env := prepareResticEnvironment()

	// Verify repository
	if err := verifyResticRepo(env); err != nil {
		return fmt.Errorf("failed to verify restic repository: %w", err)
	}

	// If list-snapshots flag is set, list available snapshots and exit
	if listSnapshots {
		return listAvailableSnapshots(env, log)
	}

	// Prepare filters for restore
	var filters []string

	if !includeGlobal {
		// Exclude global config by default
		filters = append(filters, "--exclude=*/global-config.json")
	}

	if excludeConfig {
		// Exclude configuration files if requested
		filters = append(filters, "--exclude=*/config/*")
	}

	// Create restore target directory if it doesn't exist
	if !dryRun {
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outputPath, err)
		}
	}

	// If snapshot ID is set to "latest", get the actual latest snapshot ID
	if snapshotID == "latest" {
		var err error
		snapshotID, err = getLatestSnapshotID(env)
		if err != nil {
			return fmt.Errorf("failed to get latest snapshot: %w", err)
		}
		log.Info("Found latest snapshot", "id", snapshotID)
	}

	// Build restore command
	args := []string{"restore", snapshotID, "--target", outputPath}

	// Add filters
	args = append(args, filters...)

	if dryRun {
		// In dry run mode, just print what files would be restored
		args = append(args, "--dry-run")
	}

	// Execute restore command
	cmd2 := exec.CommandContext(ctx, "restic", args...)
	cmd2.Env = append(os.Environ(), env...)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	log.Info("Starting restore operation", "outputPath", outputPath, "dryRun", dryRun)

	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	if dryRun {
		log.Info("Dry run completed. No files were actually restored.")
	} else {
		log.Info("Restore completed successfully", "outputPath", outputPath, "timestamp", time.Now().Format(time.RFC3339))
	}

	return nil
}

// listAvailableSnapshots lists available snapshots with pagination
func listAvailableSnapshots(env []string, log *logger.Logger) error {
	// Get all snapshots in JSON format
	cmd := exec.Command("restic", "snapshots", "--json")
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get snapshots: %w", err)
	}

	// Parse JSON output
	var snapshots []ResticSnapshot
	if err := json.Unmarshal(output, &snapshots); err != nil {
		return fmt.Errorf("failed to parse snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		log.Info("No snapshots found in the repository")
		return nil
	}

	// Sort snapshots by time (most recent first)
	// Sort happens automatically as restic already returns them in descending time order

	// Calculate pagination
	totalSnapshots := len(snapshots)
	totalPages := (totalSnapshots + snapshotsLimit - 1) / snapshotsLimit // Ceiling division

	// Validate page number
	if snapshotsPage < 1 {
		snapshotsPage = 1
	} else if snapshotsPage > totalPages {
		snapshotsPage = totalPages
	}

	// Calculate start and end indices for the current page
	startIndex := (snapshotsPage - 1) * snapshotsLimit
	endIndex := startIndex + snapshotsLimit
	if endIndex > totalSnapshots {
		endIndex = totalSnapshots
	}

	// Display pagination info
	log.Info(fmt.Sprintf("Showing snapshots %d-%d of %d (Page %d of %d)",
		startIndex+1, endIndex, totalSnapshots, snapshotsPage, totalPages))

	fmt.Println("\nAvailable Snapshots:")
	fmt.Println("===================")

	// Display snapshots for the current page
	for i := startIndex; i < endIndex; i++ {
		snapshot := snapshots[i]
		sizeStr := "unknown"
		if snapshot.Summary.TotalBytesProcessed > 0 {
			sizeStr = formatByteSize(snapshot.Summary.TotalBytesProcessed)
		}

		fmt.Printf("%d. ID: %s (Short: %s)\n", i+1, snapshot.ID, snapshot.ShortID)
		fmt.Printf("   Created: %s\n", snapshot.Time.Format(time.RFC3339))
		fmt.Printf("   Host: %s, User: %s\n", snapshot.Hostname, snapshot.Username)
		fmt.Printf("   Size: %s\n", sizeStr)
		fmt.Printf("   Paths: %s\n", strings.Join(snapshot.Paths, ", "))
		fmt.Println()
	}

	// Display pagination help
	if totalPages > 1 {
		fmt.Println("\nPagination:")
		fmt.Printf("Use --page=N to view different pages (1-%d)\n", totalPages)
		fmt.Printf("Use --limit=N to change the number of snapshots per page (current: %d)\n", snapshotsLimit)
	}

	// Display usage help for restore
	fmt.Println("\nTo restore a specific snapshot:")
	fmt.Println("chainlaunch backup restore --snapshot-id=<ID> [other options]")

	return nil
}

// formatByteSize formats byte size in a human-readable format
func formatByteSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// prepareResticEnvironment sets up the environment variables required for restic to access S3
func prepareResticEnvironment() []string {
	// Set up restic environment variables for S3
	env := []string{
		fmt.Sprintf("RESTIC_REPOSITORY=%s", repoURL),
		fmt.Sprintf("RESTIC_PASSWORD=%s", resticPassword),
	}

	// Add AWS credentials if provided
	if awsAccessKey != "" {
		env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", awsAccessKey))
	}
	if awsSecretKey != "" {
		env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", awsSecretKey))
	}

	// Add S3 endpoint if provided
	if s3Endpoint != "" {
		endpoint := strings.TrimPrefix(strings.TrimPrefix(s3Endpoint, "https://"), "http://")
		env = append(env, fmt.Sprintf("AWS_ENDPOINT=%s", endpoint))
	}

	// Configure path style if requested
	if s3PathStyle {
		env = append(env, "AWS_S3_FORCE_PATH_STYLE=true")
	}

	return env
}

// verifyResticRepo verifies that the restic repository exists and is accessible
func verifyResticRepo(env []string) error {
	cmd := exec.Command("restic", "check")
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = nil // Discard output

	return cmd.Run()
}

// getLatestSnapshotID gets the ID of the latest snapshot in the repository
func getLatestSnapshotID(env []string) (string, error) {
	cmd := exec.Command("restic", "snapshots", "latest", "--json")
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get snapshots: %w", err)
	}

	// Parse the snapshot ID from the JSON output
	// This is a simplified approach - the output contains a JSON array
	if len(output) == 0 {
		return "", fmt.Errorf("no snapshots found in the repository")
	}

	// Extract ID from JSON
	idStart := strings.Index(string(output), `"id":"`) + 6
	if idStart < 6 {
		return "", fmt.Errorf("couldn't parse snapshot ID from output")
	}

	idEnd := strings.Index(string(output)[idStart:], `"`) + idStart
	if idEnd <= idStart {
		return "", fmt.Errorf("couldn't parse snapshot ID from output")
	}

	return string(output)[idStart:idEnd], nil
}

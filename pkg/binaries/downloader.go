package binaries

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultVersion = "3.0.0"
	// Base URL for Hyperledger Fabric binary releases
	githubReleaseURL = "https://github.com/hyperledger/fabric/releases/download"
)

// BinaryType represents the type of binary (peer or orderer)
type BinaryType string

const (
	PeerBinary    BinaryType = "peer"
	OrdererBinary BinaryType = "orderer"
)

// BinaryDownloader handles downloading and managing Fabric binaries
type BinaryDownloader struct {
	homeDir string
}

// NewBinaryDownloader creates a new BinaryDownloader instance
func NewBinaryDownloader(homeDir string) (*BinaryDownloader, error) {
	binDir := filepath.Join(homeDir, ".chainlaunch", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create binary directory: %w", err)
	}
	return &BinaryDownloader{homeDir: homeDir}, nil
}

// GetBinaryPath returns the path to the binary, downloading it if necessary
func (d *BinaryDownloader) GetBinaryPath(binaryType BinaryType, version string) (string, error) {
	if version == "" {
		version = DefaultVersion
	}

	binDir := filepath.Join(d.homeDir, ".chainlaunch", "bin")
	binaryName := string(binaryType)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	versionDir := filepath.Join(binDir, version)
	binaryPath := filepath.Join(versionDir, "bin", binaryName)

	// Check if binary already exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// Create version directory
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create version directory: %w", err)
	}

	// Download and extract binaries
	if err := d.downloadAndExtractBinaries(version, versionDir); err != nil {
		return "", fmt.Errorf("failed to download and extract binaries: %w", err)
	}

	// Verify binary exists after extraction
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("binary not found after extraction: %w", err)
	}

	return binaryPath, nil
}

// downloadAndExtractBinaries downloads and extracts the Fabric binaries
func (d *BinaryDownloader) downloadAndExtractBinaries(version, destDir string) error {
	// Construct download URL
	arch := runtime.GOARCH
	runtimeOs := runtime.GOOS
	filename := fmt.Sprintf("hyperledger-fabric-%s-%s-%s.tar.gz", runtimeOs, arch, version)
	url := fmt.Sprintf("%s/v%s/%s", githubReleaseURL, version, filename)

	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "fabric-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download archive: HTTP %d", resp.StatusCode)
	}

	// Copy download to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save archive: %w", err)
	}

	// Rewind temp file for reading
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind temp file: %w", err)
	}

	// Open gzip reader
	gzr, err := gzip.NewReader(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Skip if not a file
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Check for directory traversal
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("invalid file path in tar: %s", header.Name)
		}

		// Get the target path
		targetPath := filepath.Join(destDir, header.Name)
		cleanTargetPath := filepath.Clean(targetPath)

		// Ensure the target path is within the destination directory
		if !strings.HasPrefix(cleanTargetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in tar: %s", header.Name)
		}

		// Create directory structure
		if err := os.MkdirAll(filepath.Dir(cleanTargetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory structure: %w", err)
		}

		// Create file
		f, err := os.OpenFile(cleanTargetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Copy contents
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return fmt.Errorf("failed to write file: %w", err)
		}
		f.Close()

		// Make binary executable if in bin directory
		if strings.HasPrefix(header.Name, "bin/") {
			if err := os.Chmod(cleanTargetPath, 0755); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}
		}
	}

	return nil
}

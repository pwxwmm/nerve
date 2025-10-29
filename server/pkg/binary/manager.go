// Package binary provides agent binary distribution and management functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package binary

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// AgentBinaryManager manages agent binary distribution
type AgentBinaryManager struct {
	binaryPath    string
	versions      map[string]*BinaryVersion
	currentVersion string
}

// BinaryVersion represents a versioned agent binary
type BinaryVersion struct {
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	Arch        string    `json:"arch"`
	Path        string    `json:"path"`
	Checksum    string    `json:"checksum"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewAgentBinaryManager creates a new binary manager
func NewAgentBinaryManager(binaryPath string) *AgentBinaryManager {
	return &AgentBinaryManager{
		binaryPath:   binaryPath,
		versions:     make(map[string]*BinaryVersion),
		currentVersion: "latest",
	}
}

// SetupBinaryRoutes sets up binary distribution routes
func (bm *AgentBinaryManager) SetupBinaryRoutes(router *gin.Engine) {
	binaries := router.Group("/api/binaries")
	{
		binaries.GET("/list", bm.listBinaries)
		binaries.POST("/upload", bm.uploadBinary)
		binaries.GET("/download/:version/:platform/:arch", bm.downloadBinary)
		binaries.DELETE("/:version", bm.deleteBinary)
	}

	// Install script endpoint
	router.GET("/install.sh", bm.serveInstallScript)
}

// listBinaries lists available agent binaries
func (bm *AgentBinaryManager) listBinaries(c *gin.Context) {
	versions := make([]*BinaryVersion, 0, len(bm.versions))
	for _, version := range bm.versions {
		versions = append(versions, version)
	}

	c.JSON(http.StatusOK, gin.H{
		"binaries": versions,
		"current":  bm.currentVersion,
	})
}

// uploadBinary handles binary upload
func (bm *AgentBinaryManager) uploadBinary(c *gin.Context) {
	file, err := c.FormFile("binary")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	version := c.PostForm("version")
	platform := c.PostForm("platform")
	arch := c.PostForm("arch")

	if version == "" || platform == "" || arch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version, platform, and arch are required"})
		return
	}

	// Save uploaded file
	dst := filepath.Join(bm.binaryPath, version, platform, arch, file.Filename)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create version record
	binaryVersion := &BinaryVersion{
		Version:   version,
		Platform:  platform,
		Arch:      arch,
		Path:      dst,
		Size:      file.Size,
		CreatedAt: time.Now(),
	}

	bm.versions[version] = binaryVersion

	c.JSON(http.StatusOK, gin.H{
		"message": "Binary uploaded successfully",
		"version": binaryVersion,
	})
}

// downloadBinary handles binary download
func (bm *AgentBinaryManager) downloadBinary(c *gin.Context) {
	version := c.Param("version")
	platform := c.Param("platform")
	arch := c.Param("arch")

	if version == "latest" {
		version = bm.currentVersion
	}

	versionKey := filepath.Join(version, platform, arch)
	binary, exists := bm.versions[versionKey]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "binary not found"})
		return
	}

	c.File(binary.Path)
}

// deleteBinary deletes a binary version
func (bm *AgentBinaryManager) deleteBinary(c *gin.Context) {
	version := c.Param("version")

	binary, exists := bm.versions[version]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "binary not found"})
		return
	}

	if err := os.Remove(binary.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	delete(bm.versions, version)

	c.JSON(http.StatusOK, gin.H{
		"message": "Binary deleted successfully",
	})
}

// serveInstallScript serves the installation script
func (bm *AgentBinaryManager) serveInstallScript(c *gin.Context) {
	token := c.Query("token")
	serverURL := c.Query("server")
	platform := c.Query("platform")
	arch := c.Query("arch")

	if token == "" || serverURL == "" {
		c.String(http.StatusBadRequest, "token and server parameters are required")
		return
	}

	if platform == "" {
		platform = "linux"
	}
	if arch == "" {
		arch = "amd64"
	}

	script := bm.generateInstallScript(token, serverURL, platform, arch)
	c.Header("Content-Type", "text/x-shellscript")
	c.String(http.StatusOK, script)
}

// generateInstallScript generates the installation script
func (bm *AgentBinaryManager) generateInstallScript(token, serverURL, platform, arch string) string {
	return `#!/bin/bash

set -e

# Configuration
TOKEN="` + token + `"
SERVER_URL="` + serverURL + `"
PLATFORM="` + platform + `"
ARCH="` + arch + `"

echo "Nerve Agent Installation Script"
echo "==============================="

# Detect platform if not specified
if [ -z "$PLATFORM" ]; then
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux" ;;
        Darwin*)    PLATFORM="darwin" ;;
        *)          echo "Unsupported platform" ; exit 1 ;;
    esac
fi

# Detect arch if not specified
if [ -z "$ARCH" ]; then
    case "$(uname -m)" in
        x86_64)     ARCH="amd64" ;;
        arm64)      ARCH="arm64" ;;
        *)          echo "Unsupported architecture" ; exit 1 ;;
    esac
fi

echo "Platform: $PLATFORM-$ARCH"
echo "Server: $SERVER_URL"

# Download agent binary
BINARY_URL="$SERVER_URL/api/binaries/download/latest/$PLATFORM/$ARCH"
AGENT_PATH="/usr/local/bin/nerve-agent"

echo "Downloading agent binary..."
curl -fsSL "$BINARY_URL" -o "$AGENT_PATH"
chmod +x "$AGENT_PATH"

# Create systemd service
cat > /etc/systemd/system/nerve-agent.service <<EOF
[Unit]
Description=Nerve Agent
After=network.target

[Service]
ExecStart=$AGENT_PATH --server=$SERVER_URL --token=$TOKEN --debug
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable nerve-agent
systemctl start nerve-agent

echo "Nerve Agent installed and started successfully!"
echo "Status: $(systemctl is-active nerve-agent)"
`
}


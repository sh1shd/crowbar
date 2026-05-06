// Package config provides configuration management utilities for the 3x-ui panel,
// including version information, logging levels, database paths, and environment variable handling.
package config

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

// LogLevel represents the logging level for the application.
type LogLevel string

// Logging level constants
const (
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Notice  LogLevel = "notice"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

// GetVersion returns the version string of the 3x-ui application.
func GetVersion() string {
	return strings.TrimSpace(version)
}

// GetName returns the name of the 3x-ui application.
func GetName() string {
	return strings.TrimSpace(name)
}

// GetLogLevel returns the current logging level based on environment variables or defaults to Info.
func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("XUI_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

// IsDebug returns true if debug mode is enabled via the XUI_DEBUG environment variable.
func IsDebug() bool {
	return os.Getenv("XUI_DEBUG") == "true"
}

// GetBinFolderPath returns the path to the binary folder, defaulting to "bin" if not set via XUI_BIN_FOLDER.
func GetBinFolderPath() string {
	binFolderPath := os.Getenv("XUI_BIN_FOLDER")
	if binFolderPath == "" {
		binFolderPath = "bin"
	}
	return binFolderPath
}

// GetPanelListenAddress returns the listen address for the panel server, defaulting to an empty string if not set via XUI_PANEL_LISTEN_ADDR.
func GetPanelListenAddress() string {
	addr := os.Getenv("XUI_PANEL_LISTEN_ADDR")
	if addr == "" {
		addr = ""
	}
	return addr
}

// GetPanelDomain returns the domain for the panel server, defaulting to an empty string if not set via XUI_PANEL_DOMAIN.
func GetPanelDomain() string {
	domain := os.Getenv("XUI_PANEL_DOMAIN")
	if domain == "" {
		domain = ""
	}
	return domain
}

// GetPanelListenPort returns the listen port for the panel server, defaulting to 2053 if not set via XUI_PANEL_LISTEN_PORT.
func GetPanelListenPort() int {
	portStr := os.Getenv("XUI_PANEL_LISTEN_PORT")
	if portStr == "" {
		return 2053
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

// GetPanelPath returns the base path for the panel server, defaulting to "/" if not set via XUI_PANEL_PATH.
func GetPanelPath() string {
	path := os.Getenv("XUI_PANEL_PATH")
	if path == "" {
		path = "/"
	}
	return path
}

// GetPanelCerificateFile returns the path to the SSL certificate file for the panel server, defaulting to an empty string if not set via XUI_PANEL_CERT_FILE.
func GetPanelCerificateFile() string {
	certFile := os.Getenv("XUI_PANEL_CERT_FILE")
	if certFile == "" {
		certFile = ""
	}
	return certFile
}

// GetPanelCertificateKey returns the path to the SSL private key file for the panel server, defaulting to an empty string if not set via XUI_PANEL_CERT_KEY.
func GetPanelCertificateKey() string {
	keyFile := os.Getenv("XUI_PANEL_CERT_KEY")
	if keyFile == "" {
		keyFile = ""
	}
	return keyFile
}

// GetPanelSessionAge returns the session maximum age in minutes for the panel server, defaulting to 360 (6 hours) if not set via XUI_PANEL_SESSION_AGE.
func GetPanelSessionAge() int {
	sessionAgeStr := os.Getenv("XUI_PANEL_SESSION_AGE")
	if sessionAgeStr == "" {
		return 360 // default to 24 hours
	}
	var sessionAge int
	fmt.Sscanf(sessionAgeStr, "%d", &sessionAge)
	return sessionAge
}

// GetSubscriptionListenAddress returns the listen address for subscription server, defaulting to an empty string if not set via XUI_SUB_LISTEN_ADDR.
func GetSubscriptionListenAddress() string {
	addr := os.Getenv("XUI_SUB_LISTEN_ADDR")
	if addr == "" {
		addr = ""
	}
	return addr
}

// GetSubscriptionDomain returns the domain for the subscription server, defaulting to an empty string if not set via XUI_SUB_DOMAIN.
func GetSubscriptionDomain() string {
	domain := os.Getenv("XUI_SUB_DOMAIN")
	if domain == "" {
		domain = ""
	}
	return domain
}

// GetSubscriptionListenPort returns the listen port for the subscription server, defaulting to 2096 if not set via XUI_SUB_LISTEN_PORT.
func GetSubscriptionListenPort() int {
	portStr := os.Getenv("XUI_SUB_LISTEN_PORT")
	if portStr == "" {
		return 2096
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

// GetSubscriptionPath returns the base path for subscription URLs, defaulting to "/sub/" if not set via XUI_SUB_PATH.
func GetSubscriptionPath() string {
	path := os.Getenv("XUI_SUB_PATH")
	if path == "" {
		path = "/sub/"
	}
	return path
}

// GetSubscriptionCertificateFile returns the path to the SSL certificate file for the subscription server, defaulting to an empty string if not set via XUI_SUB_CERT_FILE.
func GetSubscriptionCertificateFile() string {
	certFile := os.Getenv("XUI_SUB_CERT_FILE")
	if certFile == "" {
		certFile = ""
	}
	return certFile
}

// GetSubscriptionCertificateKey returns the path to the SSL private key file for the subscription server, defaulting to an empty string if not set via XUI_SUB_KEY_FILE.
func GetSubscriptionCertificateKey() string {
	keyFile := os.Getenv("XUI_SUB_KEY_FILE")
	if keyFile == "" {
		keyFile = ""
	}
	return keyFile
}

// getBaseDir returns the base directory for the application.
func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exeDir := filepath.Dir(exePath)
	exeDirLower := strings.ToLower(filepath.ToSlash(exeDir))
	if strings.Contains(exeDirLower, "/appdata/local/temp/") || strings.Contains(exeDirLower, "/go-build") {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return exeDir
}

// GetDBFolderPath returns the path to the database folder based on environment variables or platform defaults.
func GetDBFolderPath() string {
	dbFolderPath := os.Getenv("XUI_DB_FOLDER")
	if dbFolderPath != "" {
		return dbFolderPath
	}
	return "/etc/x-ui"
}

// GetDBPath returns the full path to the database file.
func GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", GetDBFolderPath(), GetName())
}

// GetLogFolder returns the path to the log folder based on environment variables or platform defaults.
func GetLogFolder() string {
	logFolderPath := os.Getenv("XUI_LOG_FOLDER")
	if logFolderPath != "" {
		return logFolderPath
	}
	return "/var/log/x-ui"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func init() {
	if os.Getenv("XUI_DB_FOLDER") != "" {
		return
	}
	oldDBFolder := "/etc/x-ui"
	oldDBPath := fmt.Sprintf("%s/%s.db", oldDBFolder, GetName())
	newDBFolder := GetDBFolderPath()
	newDBPath := fmt.Sprintf("%s/%s.db", newDBFolder, GetName())
	_, err := os.Stat(newDBPath)
	if err == nil {
		return // new exists
	}
	_, err = os.Stat(oldDBPath)
	if os.IsNotExist(err) {
		return // old does not exist
	}
	_ = copyFile(oldDBPath, newDBPath) // ignore error
}
package core

// LanguageProvider defines the interface that all language providers must implement
type LanguageProvider interface {
	Name() string
	DetectInstalled() ([]Installation, error)
	GetGlobalCacheUsage() (*DiskUsage, error)
	GetEnvVars() map[string]string

	// Phase 2: Cleaning support
	GetCleanableItems() ([]CleanableItem, error)
	Clean(items []CleanableItem) (*CleanResult, error)
}

// Installation represents a detected installation of a language/runtime
type Installation struct {
	Version     string
	Source      InstallSource
	BinaryPath  string
	ManagerPath string
}

// InstallSource represents where the language was installed from
type InstallSource string

const (
	SourceVersionManager InstallSource = "Version Manager"
	SourceHomebrew       InstallSource = "Homebrew"
	SourceSystem         InstallSource = "System"
	SourceManual         InstallSource = "Manual"
	SourceUnknown        InstallSource = "Unknown"
)

// DiskUsage represents disk space usage information
type DiskUsage struct {
	Items []DiskUsageItem
	Total int64
}

// DiskUsageItem represents a single disk usage entry
type DiskUsageItem struct {
	Path        string
	Description string
	Size        int64
}

// Status represents the health status of an installation
type Status int

const (
	StatusGood    Status = iota // 🟢 Version Manager
	StatusWarning               // 🟡 Homebrew
	StatusBad                   // 🔴 System/Conflict
)

// GetStatusIcon returns the emoji icon for a status
func (s Status) GetStatusIcon() string {
	switch s {
	case StatusGood:
		return "🟢"
	case StatusWarning:
		return "🟡"
	case StatusBad:
		return "🔴"
	default:
		return "⚪"
	}
}

// DetermineStatus determines the status based on install source
func DetermineStatus(source InstallSource) Status {
	switch source {
	case SourceVersionManager:
		return StatusGood
	case SourceHomebrew:
		return StatusWarning
	case SourceSystem, SourceUnknown:
		return StatusBad
	default:
		return StatusBad
	}
}

// CleanableItem represents an item that can be cleaned
type CleanableItem struct {
	Path        string
	Description string
	Size        int64
	Command     string // Optional: command to run instead of rm -rf
	Safe        bool   // Whether it's safe to delete without extra confirmation
}

// CleanResult represents the result of a cleaning operation
type CleanResult struct {
	ItemsCleaned   int
	SpaceReclaimed int64
	Errors         []error
}

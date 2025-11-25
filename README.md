# D-Hell CLI

<div align="center">

**Map, Measure, and Master your Dev Environment**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey)](https://www.apple.com/macos/)

*A comprehensive CLI tool to discover, classify, and audit development environment dependencies across multiple programming languages.*

</div>

---

## 🎯 Overview

D-Hell CLI helps you understand what's installed in your development environment and how much disk space it's consuming. It automatically detects programming languages, identifies their installation sources, and calculates disk usage for SDKs, caches, and package managers.

### The Problem

As developers, we install multiple programming languages, version managers, package managers, and their dependencies. Over time, these accumulate and consume significant disk space:

- 🗂️ **Module caches** (Go's `pkg/mod`, npm's `_cacache`, Maven's `.m2`)
- 📦 **Package manager stores** (pnpm's hardlink store, Yarn cache)
- 🔧 **Build caches** (Go build cache, Gradle cache)
- 🎯 **Multiple SDK versions** (via nvm, goenv, sdkman)

D-Hell CLI gives you visibility into all of this.

### Key Features

✅ **Multi-Language Support** - Go, Node.js, Java (Python, Rust coming in Phase 2)  
### ✅ **Automatic Source Detection** - Identifies Homebrew, Version Managers (nvm, goenv, sdkman), System installations  
✅ **Disk Usage Analysis** - Calculates space used by SDKs, caches, and package managers  
✅ **Beautiful Terminal UI** - Color-coded status indicators and formatted tables  
✅ **Environment Variable Inspection** - Shows relevant env vars (GOPATH, JAVA_HOME, etc.)  
✅ **Cache Cleaning** - Safe cache cleaning with dry-run and interactive confirmation (Phase 2)  
🚧 **Project Scanning** - Find wasted space in old projects (Phase 3)

---

## 📦 Installation

### Via Go Install (Recommended)

```bash
go install github.com/nguyentantai21042004/dependency-hell-cli@latest
```

### From Source

```bash
git clone https://github.com/nguyentantai21042004/dependency-hell-cli.git
cd dependency-hell-cli
go build -o dhell
sudo mv dhell /usr/local/bin/
```

### Verify Installation

```bash
dhell --version
```

---

## 🚀 Quick Start

### Scan All Languages

```bash
dhell scan
```

**Output:**
```
                                     
  Dependency Hell Analyzer (v0.1.0)  
                                     
OS: darwin 15.7.1 (ARM64)

 STATUS   LANGUAGE     VERSION         SOURCE             DISK USAGE                                  
────────────────────────────────────────────────────────────────────────────────────────────────────
 🟢      Golang        1.25.4          Version Manager   Total: 2.2 GB                                
                                                           ↳ SDK: 203 MB                              
                                                           ↳ Build Cache: 893 MB                      
                                                           ↳ Module Cache: 1.1 GB                     
────────────────────────────────────────────────────────────────────────────────────────────────────
 🟢      Node.js       v18.20.8        Version Manager   Total: 5.2 GB                                
                                                           ↳ NVM Versions: 241 MB                     
                                                           ↳ NPM Cache: 4.9 GB                        
────────────────────────────────────────────────────────────────────────────────────────────────────
```

### Scan Specific Languages

```bash
# Scan only Go
dhell scan --lang go

# Scan Go and Node.js
dhell scan --lang go,node
```

### Verbose Output

```bash
dhell scan --verbose
```

---

## 📊 Supported Languages

| Language | Detection Method | Version Managers | Cache Locations |
|----------|-----------------|------------------|-----------------|
| **Go** | `go version` | goenv, Homebrew, Manual | `GOMODCACHE`, `GOCACHE`, `GOROOT` |
| **Node.js** | `node --version` | nvm, Volta, Homebrew | npm cache, yarn cache, **pnpm store** |
| **Java** | `java -version` | SDKMAN!, Homebrew, Manual | Maven `.m2`, Gradle cache |

---

## 🎨 Status Indicators

D-Hell CLI uses color-coded status indicators to show the health of your installations:

- 🟢 **Green (Good)** - Managed by a version manager (nvm, goenv, sdkman)
  - ✅ Easy to switch versions
  - ✅ Isolated from system
  - ✅ Best practice

- 🟡 **Yellow (Warning)** - Installed via Homebrew
  - ⚠️ Harder to manage multiple versions
  - ⚠️ Global installation
  - ℹ️ Acceptable for single-version use

- 🔴 **Red (Bad)** - System installation or unknown source
  - ❌ Difficult to update
  - ❌ May conflict with other tools
  - ❌ Consider migrating to version manager

---

## 🏗️ How It Works

### Architecture

D-Hell CLI uses a **Plugin/Provider Pattern**:

```
┌─────────────────────────────────────────┐
│          CLI Command (Cobra)            │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │  Scan Controller  │
        └─────────┬─────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼───┐    ┌───▼───┐    ┌───▼───┐
│  Go   │    │ Node  │    │ Java  │
│Provider│    │Provider│    │Provider│
└───────┘    └───────┘    └───────┘
```

Each provider implements the `LanguageProvider` interface:

```go
type LanguageProvider interface {
    Name() string
    DetectInstalled() ([]Installation, error)
    GetGlobalCacheUsage() (*DiskUsage, error)
    GetEnvVars() map[string]string
}
```

### Detection Strategy

1. **Find Executable** - Use `which` to locate binary in PATH
2. **Resolve Symlinks** - Follow symlinks to actual installation
3. **Classify Source** - Analyze path to determine installation source:
   - Contains `.goenv`, `.nvm`, `.sdkman` → Version Manager
   - Contains `/opt/homebrew`, `/usr/local/Cellar` → Homebrew
   - System paths → System installation
4. **Scan Caches** - Calculate disk usage for known cache locations
5. **Extract Env Vars** - Collect relevant environment variables

### Disk Usage Calculation

D-Hell CLI walks directory trees to calculate actual disk space:

- **Concurrent Scanning** - Uses goroutines for fast parallel scanning
- **Symlink Aware** - Handles symlinks correctly
- **Error Tolerant** - Continues on permission errors

---

## 🛠️ Commands

### `dhell scan`

Scan installed languages and their disk usage.

**Flags:**
- `--lang, -l` - Filter languages (comma-separated: `go,node,java`)
- `--verbose, -v` - Verbose output

**Examples:**
```bash
dhell scan                    # Scan all
dhell scan --lang go          # Go only
dhell scan --lang go,node -v  # Go and Node with verbose output
```

### `dhell clean`

Clean caches for a specific language.

**Arguments:**
- `<language>` - Language to clean (go, node, java, all)

**Flags:**
- `--dry-run` - Preview what would be deleted without actually deleting
- `--force` - Skip confirmation prompts (use with caution)
- `--verbose, -v` - Show detailed progress

**Examples:**
```bash
dhell clean go                   # Clean Go caches (with confirmation)
dhell clean node --dry-run       # Preview Node.js cleaning
dhell clean java --force         # Clean Java without confirmation
dhell clean all                  # Clean all languages
```

**Safety:**
- Interactive confirmation by default
- Shows size of items to be deleted
- Dry-run mode for safe preview
- Caches will be rebuilt on next use

### `dhell --version`

Show version information.

### `dhell --help`

Show help message.

---

## 💡 Use Cases

### 1. Disk Space Audit

Find out what's consuming disk space in your dev environment:

```bash
dhell scan
```

Look for large caches (Go module cache, pnpm store, Maven repository).

### 2. Environment Debugging

Check if languages are installed correctly:

```bash
dhell scan --verbose
```

Verify binary paths and environment variables.

### 3. Migration Planning

Identify system installations that should be migrated to version managers:

```bash
dhell scan
```

Look for 🔴 red status indicators.

### 4. Team Onboarding

New team members can quickly see what's installed:

```bash
dhell scan
```

Helps ensure consistent development environments.

---

## 🔧 Development

### Project Structure

```
dependency-hell-cli/
├── main.go                    # Entry point
├── cmd/                       # Cobra commands
│   ├── root.go               # Root command
│   └── scan.go               # Scan command
├── internal/
│   ├── core/                 # Core interfaces & types
│   │   └── types.go
│   ├── scanner/              # Filesystem utilities
│   │   ├── disk.go          # Disk usage calculator
│   │   └── path.go          # Path utilities
│   ├── providers/            # Language providers
│   │   ├── golang.go        # Go provider
│   │   ├── nodejs.go        # Node.js provider
│   │   └── java.go          # Java provider
│   └── output/               # Output formatting
│       ├── table.go         # Table renderer
│       └── styles.go        # Lipgloss styles
```

### Adding a New Language Provider

1. Create a new file in `internal/providers/`:

```go
package providers

import "github.com/nguyentantai21042004/dependency-hell-cli/internal/core"

type PythonProvider struct{}

func NewPythonProvider() *PythonProvider {
    return &PythonProvider{}
}

func (p *PythonProvider) Name() string {
    return "Python"
}

func (p *PythonProvider) DetectInstalled() ([]core.Installation, error) {
    // Implementation
}

func (p *PythonProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
    // Implementation
}

func (p *PythonProvider) GetEnvVars() map[string]string {
    // Implementation
}
```

2. Register the provider in `cmd/scan.go`:

```go
allProviders := []core.LanguageProvider{
    providers.NewGoProvider(),
    providers.NewNodeProvider(),
    providers.NewJavaProvider(),
    providers.NewPythonProvider(), // Add here
}
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o dhell
```

---

## ❓ FAQ

### Q: Is it safe to clean caches?

**A:** Yes, but understand what you're cleaning:
- **Go module cache** - Safe to clean, will re-download on next build
- **npm/yarn cache** - Safe to clean, may slow down next install
- **pnpm store** - Safe to clean, but uses hardlinks to save space
- **Maven/Gradle cache** - Safe to clean, will re-download dependencies

Cache cleaning feature coming in Phase 2!

### Q: Why is my pnpm store so large?

**A:** pnpm uses a content-addressable store with hardlinks. The actual disk usage is shared across projects, but D-Hell CLI shows the total size. This is expected behavior.

### Q: Can I use this on Linux/Windows?

**A:** Phase 1 (MVP) is optimized for macOS. Linux support is planned for Phase 2. Windows support requires additional path handling.

### Q: Why does it show "Unknown" source?

**A:** D-Hell CLI uses heuristics to detect installation sources. If the binary path doesn't match known patterns (Homebrew, version managers), it's marked as "Unknown". This usually means manual installation.

### Q: How accurate is the disk usage calculation?

**A:** Very accurate. D-Hell CLI walks directory trees and sums file sizes. However:
- Symlinks are handled correctly
- Hardlinks (pnpm) may show inflated sizes
- Permission errors are skipped

---

## 🗺️ Roadmap

### ✅ Phase 1: MVP (Complete)
- [x] Core scanning for Go, Node.js, Java
- [x] Source detection (Homebrew, Version Managers, System)
- [x] Disk usage analysis
- [x] Beautiful terminal UI

### ✅ Phase 2: Deep Clean (Complete)
- [x] `dhell clean <lang>` command
- [x] Safe cache cleaning (Go, npm, Maven, Gradle)
- [x] Interactive mode with confirmations
- [x] Dry-run mode for safe preview

### 🔮 Phase 3: Project Scanner (Future)
- [ ] `dhell projects` command
- [ ] Scan `~/Projects` for wasted space
- [ ] Find old `node_modules`, `target`, `venv` directories
- [ ] Calculate total reclaimable space
- [ ] Python and Rust support

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Guidelines

1. Follow Go best practices
2. Add tests for new providers
3. Update README for new features
4. Keep the Plugin/Provider pattern

---

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [gopsutil](https://github.com/shirou/gopsutil) - System information
- [go-humanize](https://github.com/dustin/go-humanize) - Human-readable sizes

---

<div align="center">

**Made with ❤️ for developers drowning in dependency hell**

[Report Bug](https://github.com/nguyentantai21042004/dependency-hell-cli/issues) · [Request Feature](https://github.com/nguyentantai21042004/dependency-hell-cli/issues)

</div>

# TECHNICAL PROPOSAL: D-HELL CLI

**Slogan:** Map, Measure, and Master your Dev Environment.

## 1\. Mục tiêu dự án

Xây dựng một công cụ CLI (Command Line Interface) tập trung vào việc:

1.  **Discovery (Khám phá):** Tự động phát hiện các ngôn ngữ/runtime đã cài đặt (Go, Node, Python, Java, Rust, PHP...).
2.  **Classification (Phân loại):** Xác định chính xác nguồn gốc cài đặt (System Pre-installed, Homebrew, Version Manager như nvm/goenv/sdkman, hay cài thủ công từ binary).
3.  **Audit (Kiểm toán):** Đo lường dung lượng ổ cứng thực tế mà hệ sinh thái đó chiếm dụng (bao gồm Binary, Global Packages, Caches, Registries).
4.  **Config Check:** Hiển thị các biến môi trường quan trọng (`PATH`, `GOPATH`, `JAVA_HOME`...) để debug lỗi path.

-----

## 2\. Kiến trúc Kỹ thuật (Architecture)

Sử dụng kiến trúc **Plugin/Provider Pattern**. Mỗi ngôn ngữ sẽ là một "Provider" tuân theo một Interface chung.

### 2.1 Tech Stack

  * **Ngôn ngữ:** Golang (1.21+).
  * **CLI Library:** `spf13/cobra` (Command structure).
  * **UI/Output:** `lipgloss` hoặc `pterm` (Để render bảng biểu, màu sắc đẹp mắt trên terminal).
  * **System Info:** `shirou/gopsutil` (Lấy thông tin hệ thống).

### 2.2 Core Logic (Interface Design)

Mỗi ngôn ngữ (ví dụ: `GoProvider`, `NodeProvider`) sẽ phải implement interface sau:

```go
type LanguageProvider interface {
    Name() string                  // e.g., "Golang"
    DetectInstalled() []Installation // Trả về list các version tìm thấy
    GetGlobalCacheUsage() DiskUsage  // Scan các thư mục cache (e.g., ~/.npm, ~/go/pkg)
    GetEnvVars() map[string]string   // Lấy env vars liên quan (GOPATH...)
}

type Installation struct {
    Version      string
    Source       string // "Homebrew", "Version Manager", "System", "Unknown"
    BinaryPath   string
    ManagerPath  string // e.g., ~/.nvm/versions/node/v18...
}
```

-----

## 3\. Chiến lược phát hiện & Scan (Implementation Detail)

Đây là phần quan trọng nhất ("Logic nghiệp vụ"). Tool sẽ quét theo các quy tắc heuristic sau:

### 3.1. Golang

  * **Detection:** Quét `go version`.
  * **Phân loại nguồn:**
      * Nếu path chứa `.goenv`: -\> **goenv**.
      * Nếu path chứa `/opt/homebrew` hoặc `/usr/local/Cellar`: -\> **Homebrew**.
      * Nếu path là `/usr/local/go`: -\> **Manual Install**.
  * **Dung lượng cần quét:**
      * SDKs: `~/.goenv/versions` hoặc `$(go env GOROOT)`.
      * Build Cache: `$(go env GOCACHE)` (Thường là `~/Library/Caches/go-build`).
      * Module Cache: `$(go env GOPATH)/pkg/mod` (**Thủ phạm ngốn dung lượng số 1**).

### 3.2. Node.js Ecosystem (JS, TS, Node, NPM, PNPM, Yarn)

  * **Detection:** Quét `node`, `npm`, `pnpm`, `yarn`.
  * **Phân loại nguồn:**
      * Path chứa `.nvm`: -\> **NVM**.
      * Path chứa `.voltap`: -\> **Volta**.
      * Path `/opt/homebrew`: -\> **Homebrew**.
  * **Dung lượng cần quét:**
      * NVM Versions: `~/.nvm/versions`.
      * NPM Global Cache: `~/.npm/_cacache`.
      * Yarn Cache: `~/Library/Caches/Yarn` (hoặc `~/.yarn`).
      * **PNPM Store:** `~/.local/share/pnpm/store` (Cái này thường rất lớn vì chứa hardlink của tất cả project).

### 3.3. Python

  * **Detection:** Quét `python3`, `pip`.
  * **Phân loại nguồn:**
      * Path chứa `.pyenv`: -\> **pyenv**.
      * Path chứa `anaconda` / `miniconda`: -\> **Conda**.
      * Path `/usr/bin/python3`: -\> **System (macOS default - Do not touch)**.
  * **Dung lượng cần quét:**
      * Pyenv Versions: `~/.pyenv/versions`.
      * Pip Cache: `~/Library/Caches/pip`.
      * Virtualenvs (nếu gom tập trung): `~/.virtualenvs`.

### 3.4. Rust

  * **Detection:** Quét `rustc`, `cargo`.
  * **Phân loại nguồn:**
      * Path chứa `.cargo/bin`: -\> **Rustup** (Chuẩn mực).
      * Khác: Homebrew.
  * **Dung lượng cần quét:**
      * Toolchains: `~/.rustup/toolchains`.
      * Registry & Git Checkouts: `~/.cargo/registry` và `~/.cargo/git` (**Rất nặng**).
      * Target (Build artifacts): Thường nằm trong project, nhưng cần cảnh báo user.

### 3.5. Java

  * **Detection:** Quét `java`, check biến `JAVA_HOME`.
  * **Phân loại:**
      * Path chứa `.sdkman`: -\> **SDKMAN\!**.
      * Path `/Library/Java/...`: -\> **Manual/Installer**.
  * **Dung lượng cần quét:**
      * SDKs: `~/.sdkman/candidates/java`.
      * **Maven Repo:** `~/.m2/repository` (Nơi chứa các thư viện `.jar` đã tải về).
      * Gradle Cache: `~/.gradle/caches`.

-----

## 4\. Thiết kế giao diện CLI (UX)

Khi user gõ lệnh `dhell scan`, output sẽ có dạng bảng như sau:

```text
Dependency Hell Analyzer (v0.1.0)
OS: macOS Sequoia (ARM64)

STATUS | LANGUAGE | VERSION     | SOURCE    | BINARY PATH                  | DISK USAGE (Est.)
-------|----------|-------------|-----------|------------------------------|------------------
🟢     | Golang   | 1.21.3      | goenv     | ~/.goenv/shims/go            | 1.2 GB (SDK)
       |          |             |           |                              | 5.4 GB (Mod Cache)
-------|----------|-------------|-----------|------------------------------|------------------
🟡     | Node.js  | v18.17.0    | Homebrew  | /opt/homebrew/bin/node       | 350 MB
       |          |             |           |                              | 12.0 GB (pnpm store)
-------|----------|-------------|-----------|------------------------------|------------------
🔴     | Python   | 3.9.6       | System    | /usr/bin/python3             | N/A (Protected)
🟢     | Python   | 3.11.0      | pyenv     | ~/.pyenv/shims/python        | 800 MB
-------|----------|-------------|-----------|------------------------------|------------------
🟢     | Java     | 17.0.9-tem  | SDKMAN    | ~/.sdkman/.../current/java   | 300 MB
       |          |             |           |                              | 2.1 GB (.m2 repo)
```

**Chú thích màu sắc:**

  * 🟢 **Xanh:** Quản lý tốt (Dùng Version Manager).
  * 🟡 **Vàng:** Cài qua Homebrew (Chấp nhận được nhưng khó switch version).
  * 🔴 **Đỏ:** Cài thẳng vào System hoặc xung đột phiên bản / Đường dẫn lạ.

-----

## 5\. Roadmap phát triển

1.  **Phase 1 (MVP):**

      * Dựng khung CLI bằng Golang.
      * Implement module `Scanner` cơ bản (Scan path, size).
      * Implement detection cho: Golang, Node, Java (3 cái quan trọng nhất của bạn).
      * Output ra bảng đơn giản.

2.  **Phase 2 (Deep Clean):**

      * Thêm tính năng `dhell clean <lang>`.
      * Ví dụ: `dhell clean go` -\> Tự động chạy `go clean -modcache`.
      * Ví dụ: `dhell clean npm` -\> `npm cache clean --force`.

3.  **Phase 3 (Project Scanner):**

      * Quét toàn bộ thư mục `~/Projects`.
      * Phát hiện `node_modules`, `target` (Rust), `venv` (Python) nằm rải rác trong các dự án cũ và tính tổng dung lượng lãng phí.

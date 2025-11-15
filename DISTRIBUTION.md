# Distribution Guide

This document explains how to distribute the dictation CLI application.

## Distribution Methods

### 1. GitHub Releases (Recommended)

The easiest and most common way to distribute a Go CLI is via GitHub Releases with pre-built binaries.

#### Using GoReleaser (Automated)

1. **Install GoReleaser:**
   ```bash
   brew install goreleaser  # macOS
   # or
   go install github.com/goreleaser/goreleaser@latest
   ```

2. **Configure `.goreleaser.yml`:**
   - Update `owner` and `name` in the file
   - Customize as needed

3. **Set up GitHub authentication:**
   GoReleaser needs a GitHub token to create releases. You have two options:
   
   **Option A: Personal Access Token (for local runs)**
   ```bash
   # Create a token at: https://github.com/settings/tokens
   # Needs 'repo' scope (full control of private repositories)
   export GITHUB_TOKEN=ghp_your_token_here
   ```
   
   **Option B: GitHub Actions (recommended for CI/CD)**
   - No manual token needed
   - Uses `GITHUB_TOKEN` automatically provided by GitHub Actions
   - More secure and automated

4. **Create a release:**
   ```bash
   # Test locally first (creates files but doesn't upload)
   goreleaser release --snapshot
   
   # Create a real release (requires git tag and GITHUB_TOKEN)
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   goreleaser release
   ```
   
   **What GoReleaser does:**
   - Builds binaries for all configured platforms (Linux, macOS, Windows)
   - Creates archives (tar.gz, zip) with binaries and assets
   - Generates checksums (SHA256) for verification
   - Creates a GitHub Release (or updates existing)
   - Uploads all artifacts to the GitHub Release
   - Optionally creates/updates Homebrew formulas, etc.

4. **Users download:**
   - Go to GitHub Releases page
   - Download binary for their OS/architecture
   - Extract and run

#### Manual Build

If you prefer manual builds:

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o dictation-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o dictation-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o dictation-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o dictation-windows-amd64.exe

# Create archives
tar -czf dictation-linux-amd64.tar.gz dictation-linux-amd64 config.yaml active.*.toml
tar -czf dictation-darwin-amd64.tar.gz dictation-darwin-amd64 config.yaml active.*.toml
zip dictation-windows-amd64.zip dictation-windows-amd64.exe config.yaml active.*.toml
```

### 2. Homebrew (macOS/Linux)

For macOS users, Homebrew is the most convenient installation method.

#### Create a Homebrew Formula

1. **Create a tap repository** (or use existing):
   ```bash
   brew tap-new yourusername/homebrew-dictation
   ```

2. **Create formula file:**
   ```ruby
   # Formula: dictation.rb
   class Dictation < Formula
     desc "German dictation practice CLI application"
     homepage "https://github.com/yourusername/dictation"
     url "https://github.com/yourusername/dictation/releases/download/v1.0.0/dictation_darwin_arm64.tar.gz"
     sha256 "..." # Calculate with: shasum -a 256 <file>
     version "1.0.0"
   
     def install
       bin.install "dictation"
       # Install translation files
       (prefix/"etc/dictation").install "active.en.toml", "active.de.toml"
       (prefix/"etc/dictation").install "config.yaml"
     end
   
     test do
       system "#{bin}/dictation", "--version"
     end
   end
   ```

3. **Users install:**
   ```bash
   brew install yourusername/dictation/dictation
   ```

### 3. Go Install (For Developers)

If the repository is public on GitHub:

```bash
go install github.com/yourusername/dictation@latest
```

**Pros:**
- Simple for developers
- Always gets latest version

**Cons:**
- Requires Go installed
- Not suitable for end users

### 4. Package Managers (Linux)

#### Debian/Ubuntu (apt)

1. Create a `.deb` package using `dpkg-deb`
2. Host in a PPA or repository
3. Users install: `apt install dictation`

#### Arch Linux (AUR)

Create a PKGBUILD file for the AUR.

#### Fedora/RHEL (rpm)

Create an RPM package.

### 5. Docker

For containerized distribution:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dictation

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dictation .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/active.*.toml .
ENTRYPOINT ["./dictation"]
```

Users run:
```bash
docker run -v $(pwd)/config.yaml:/root/config.yaml yourusername/dictation
```

## Best Practices

1. **Versioning:**
   - Use semantic versioning (v1.0.0)
   - Tag releases in git
   - Include version in binary: `go build -ldflags "-X main.Version=v1.0.0"`

2. **Include Assets:**
   - Translation files (active.*.toml)
   - Example config.yaml
   - README.md

3. **Checksums:**
   - Always provide checksums (SHA256) for verification
   - GoReleaser does this automatically

4. **Cross-platform:**
   - Build for common platforms:
     - Linux (amd64, arm64)
     - macOS (amd64, arm64)
     - Windows (amd64)

5. **Documentation:**
   - Installation instructions
   - Usage examples
   - Troubleshooting guide

## Recommended Approach

For this CLI application, I recommend:

1. **Primary:** GitHub Releases with GoReleaser
   - Automated builds
   - Multiple platforms
   - Easy for users

2. **Secondary:** Homebrew (if targeting macOS users)
   - Most convenient for macOS
   - Professional distribution method

3. **Optional:** Go install for developers
   - Quick setup for contributors

## Example Workflow

### Automated (Recommended - Using GitHub Actions)

1. **Set up GitHub Actions workflow** (already configured in `.github/workflows/release.yml`)

2. **Create and push a tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions automatically:**
   - Detects the tag push
   - Runs GoReleaser
   - Builds all binaries
   - Creates GitHub Release
   - Uploads all artifacts

**No manual authentication needed!** GitHub Actions provides `GITHUB_TOKEN` automatically.

### Manual (Local)

1. **Get GitHub token:**
   - Go to https://github.com/settings/tokens
   - Generate new token (classic) with `repo` scope
   - Copy the token

2. **Set environment variable:**
   ```bash
   export GITHUB_TOKEN=ghp_your_token_here
   ```

3. **Create release:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   goreleaser release
   ```

4. **GoReleaser will:**
   - Build binaries for all platforms
   - Create archives
   - Generate checksums
   - Create GitHub Release
   - Upload everything automatically

## Resources

- [GoReleaser Documentation](https://goreleaser.com)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Go Release Process](https://go.dev/doc/modules/release-workflow)

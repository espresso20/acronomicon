# 📖 THE ACRONOMICON — The Bible of Cloud Tech Acronyms

**Acronomicon** is a fast, zero-dependency, full-color cross-platform terminal tool (macOS Terminal, Linux, and Windows CMD/PowerShell) containing **400+ curated Cloud & DevOps acronyms** spanning **A to Z**.

Every entry includes:
- **Canonical Acronym & Aliases** (e.g., `K8s` / `KUBERNETES`, `VPC` / `VNET`, `eBPF`, `mTLS`, `CIDR`, `IRSA`)
- **Cloud Provider / Ecosystem Badge** (`Cloud-Agnostic`, `AWS`, `GCP`, `Azure`, `CNCF`)
- **Category Tag** (`Networking`, `Security & Identity`, `Containers & K8s`, `Storage & Databases`, `Compute`, `DevOps & IaC`, `Observability & SRE`, `Serverless & Messaging`, `Architecture & Cloud Models`, `Compliance & Governance`)
- **Full Name** (what the letters stand for)
- **Meaning (TL;DR)** (crisp 1-sentence summary)
- **What It Does & Why It Matters** (deep explanation of how it works under the hood)
- **In Practice / Real-World Example** (concrete engineering usage)
- **Related Acronyms** (cross-referenced entries to explore next)

---

## 🚀 Quick Start

### 1. Interactive REPL Mode (Default)
Run with no arguments to launch the interactive prompt with ASCII banner:
```bash
./acronomicon
```
Inside `acronomicon>`:
- Type any acronym (`vpc`, `k8s`, `ebpf`, `iam`, `mtls`, `s3`, `raft`) to view its full-color card
- Type any single letter (`a`–`z`) to list all acronyms starting with that letter
- Type `?<query>` or `/search <query>` to search across all definitions
- Type `/cat` or `/cat <name>` to browse by category
- Type `/random` (or `/r`) for a random *Acronomicon Verse of the Day*
- Type `/quiz` for an interactive 5-question multiple-choice quiz
- Type `/stats` for database statistics and A–Z distribution charts
- Type `q` or `quit` to exit

### 2. Direct CLI Commands
```bash
# Look up any acronym directly (case-insensitive, supports aliases & fuzzy suggestions)
./acronomicon k8s
./acronomicon ebpf
./acronomicon vpc

# Search across acronyms, full names, categories, and descriptions
./acronomicon search zero trust
./acronomicon -s network

# List the full A-Z index or filter by a single letter
./acronomicon list
./acronomicon list k

# Browse categories
./acronomicon category
./acronomicon category security

# Random "Verse of the Day"
./acronomicon random

# Interactive multiple-choice quiz
./acronomicon quiz

# Database statistics & A-Z bar charts
./acronomicon stats

# Disable ANSI colors (or set NO_COLOR=1)
./acronomicon --no-color vpc
```

---

## 🛠️ Building from Source (Pure Go stdlib, Zero Dependencies)

```bash
# Run unit tests
go test -v ./...

# Build native binary
go build -o acronomicon .

# Cross-compile for macOS (Apple Silicon & Intel) and Windows (x64 & ARM64)
mkdir -p dist
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/acronomicon-darwin-arm64 .
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/acronomicon-darwin-amd64 .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/acronomicon-windows-amd64.exe .
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/acronomicon-windows-arm64.exe .
```

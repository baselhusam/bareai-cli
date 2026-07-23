# Installation

Copy a block for your platform, run it, then verify with `bareai version`.

**Latest release:** [GitHub Releases](https://github.com/baselhusam/bareai-cli/releases/latest)

---

## macOS / Linux (recommended)

No `sudo` required. Installs to `~/.local/bin` and adds it to your shell `PATH`.

```bash
curl -fsSL https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.sh | bash
```

If `bareai` is not found yet, reload your shell config (the installer prints the right file):

```bash
source ~/.zshrc    # zsh (default on macOS)
# source ~/.bashrc # bash
bareai version
```

Pin a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.sh | bash -s -- --version v0.1.0
```

System-wide install to `/usr/local/bin` (may prompt for your password):

```bash
curl -fsSL https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.sh | bash -s -- --system
```

---

## Homebrew (macOS / Linux)

Third-party tap — Homebrew asks you to trust it once before the first install.

```bash
brew tap baselhusam/tap
brew trust baselhusam/tap
brew install bareai
bareai version
```

Upgrade later:

```bash
brew update && brew upgrade bareai
```

---

## Windows

**winget:**

```powershell
winget install baselhusam.bareai
bareai version
```

**PowerShell install script** (downloads from GitHub Releases):

```powershell
irm https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.ps1 | iex
```

Add install dir to your user `PATH` automatically:

```powershell
irm https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.ps1 | iex; .\scripts\install.ps1 -AddToPath
```

---

## Debian / Ubuntu (APT via Cloudsmith)

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/baselhusam/bareai/setup.deb.sh' | sudo bash
sudo apt update && sudo apt install bareai
bareai version
man bareai-doctor
```

---

## Manual download

1. Open [GitHub Releases](https://github.com/baselhusam/bareai-cli/releases/latest)
2. Download the archive for your OS/arch (and `checksums.txt`)
3. Verify SHA256, extract, move `bareai` onto your `PATH`

**macOS (Apple Silicon):**

```bash
curl -LO https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/bareai_0.1.0_darwin_arm64.tar.gz
curl -LO https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing
tar xzf bareai_0.1.0_darwin_arm64.tar.gz
mkdir -p ~/.local/bin
mv bareai ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
bareai version
```

**Linux (amd64):** use `bareai_0.1.0_linux_amd64.tar.gz` instead.

**Windows:** download `bareai_0.1.0_windows_amd64.zip` or `..._arm64.zip`, extract `bareai.exe`, add its folder to `PATH`.

---

## Shell completions

```bash
bareai completion bash >> ~/.bashrc
bareai completion zsh >> ~/.zshrc
bareai completion fish > ~/.config/fish/completions/bareai.fish
bareai completion powershell >> $PROFILE
```

Reload your shell, then type `bareai <Tab>`.

---

## Build from source

**Requirements:** Go 1.25+

```bash
git clone https://github.com/baselhusam/bareai-cli.git
cd bareai-cli
make build
./bareai version
```

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `bareai: command not found` after install script | Run `source ~/.zshrc` (or `~/.bashrc`), or open a new terminal |
| Install script asks for a password | You are not on the latest script — re-run the `curl \| bash` one-liner from above (defaults to `~/.local/bin`, no sudo) |
| `Refusing to load formula … untrusted tap` | Run `brew trust baselhusam/tap`, then `brew install bareai` |
| `Repository not found` on `brew tap` | The tap repo is missing — use the install script instead, or check [homebrew-tap](https://github.com/baselhusam/homebrew-tap) |
| `winget` package not found | Package may still be pending in winget-pkgs — use the PowerShell install script or a GitHub Release zip |
| APT `bareai` not found | Cloudsmith repo not configured — run the `curl … setup.deb.sh` line first |

---

Release process for maintainers: [RELEASE.md](RELEASE.md).

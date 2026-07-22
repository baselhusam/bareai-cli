# Installation

## macOS / Linux (install script)

```bash
curl -fsSL https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.sh | bash
```

Pin a version: `VERSION=v0.1.0 curl -fsSL ... | bash`

## Homebrew

```bash
brew tap baselhusam/tap
brew install bareai
```

## Windows

**winget:**

```powershell
winget install baselhusam.bareai
```

**PowerShell install script:**

```powershell
irm https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.ps1 | iex
```

Add to PATH: `.\scripts\install.ps1 -AddToPath`

## Debian / Ubuntu (Cloudsmith APT)

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/baselhusam/bareai/cfg/setup/deb.sh' | sudo bash
sudo apt update && sudo apt install bareai
man bareai-doctor
```

## Manual download

Download archives and `checksums.txt` from [GitHub Releases](https://github.com/baselhusam/bareai-cli/releases). Verify SHA256 before installing.

## Shell completions

```bash
bareai completion bash >> ~/.bashrc
bareai completion zsh >> ~/.zshrc
bareai completion fish > ~/.config/fish/completions/bareai.fish
bareai completion powershell >> $PROFILE
```

## Build from source

**Requirements:** Go 1.25+

```bash
git clone https://github.com/baselhusam/bareai-cli.git
cd bareai-cli
make build
./bareai version
```

Release process for maintainers: [RELEASE.md](RELEASE.md).

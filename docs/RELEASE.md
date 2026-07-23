# Release runbook

How to cut a `bareai` release and publish packages.

## Prerequisites

### GitHub secrets (`bareai-cli` repository)

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Provided by Actions; publishes GitHub Releases |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT with `repo` scope for `baselhusam/homebrew-tap` |
| `CLOUDSMITH_API_KEY` | Cloudsmith API key for `.deb` upload |
| `WINGET_TOKEN` | PAT with fork/PR access for `microsoft/winget-pkgs` |

### External services

1. **Homebrew tap:** `https://github.com/baselhusam/homebrew-tap` — see [`packaging/homebrew-tap/README.md`](../packaging/homebrew-tap/README.md)
2. **Cloudsmith APT:** repository `baselhusam/bareai` (deb, public)
3. **winget:** automated via `winget-releaser` in [`.github/workflows/release.yml`](../.github/workflows/release.yml)

## Cloudsmith setup (one-time)

1. Create a Cloudsmith account and repository: **bareai** (format: deb)
2. Set repository visibility to **public**
3. Generate an API key with upload permissions
4. Add `CLOUDSMITH_API_KEY` to GitHub secrets
5. Copy the setup script URL for README/docs:
   ```bash
   curl -1sLf 'https://dl.cloudsmith.io/public/baselhusam/bareai/cfg/setup/deb.sh' | sudo bash
   sudo apt update && sudo apt install bareai
   ```

The Release workflow uploads `.deb` artifacts to Cloudsmith after GoReleaser finishes (see [`.github/workflows/release.yml`](../.github/workflows/release.yml)).

## Local dry-run

```bash
make man
make goreleaser-check
make release-snapshot
ls dist/
```

Man pages are regenerated automatically in the GoReleaser `before` hook; run `make man` locally before cutting a release to verify.

Snapshot builds use version `0.0.0-next` (or similar) and do not publish.

## Cut a release

1. Ensure `main` is green in CI
2. Update changelog / version notes if needed
3. Tag and push:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. GitHub Actions **Release** workflow runs GoReleaser
5. Verify artifacts on GitHub Releases:
   - 6 archives (`linux/darwin/windows` × `amd64/arm64`)
   - `checksums.txt`
   - `.deb` packages
6. Verify downstream:
   - Homebrew formula updated in `homebrew-tap`
   - Cloudsmith shows new `.deb`
   - winget PR opened (or submit manually from [`packaging/winget/`](../packaging/winget/))

## Manual winget fallback

If `winget-releaser` fails:

1. Copy templates from `packaging/winget/baselhusam.bareai/`
2. Update version, URLs, and SHA256 from `checksums.txt`
3. Open PR to `microsoft/winget-pkgs` under `manifests/b/baselhusam/bareai/<version>/`

## Install script smoke test

After a release is published:

```bash
VERSION=v0.1.0 ./scripts/install.sh
```

Windows:

```powershell
.\scripts\install.ps1 -Version v0.1.0 -AddToPath
```

## Troubleshooting

| Issue | Check |
|-------|-------|
| GoReleaser `already_exists` on re-run | Use **Actions → Release → Run workflow** with **Cloudsmith only** + tag `v0.1.0`, or tag a new release from `main` (has `replace_existing_artifacts`) |
| Homebrew formula not updated | `HOMEBREW_TAP_GITHUB_TOKEN` scope; tap repo exists |
| Cloudsmith upload failed | `CLOUDSMITH_API_KEY`; repo name `baselhusam/bareai`; try **Run workflow → Cloudsmith only** |
| winget PR failed | `WINGET_TOKEN`; fork of `winget-pkgs` |
| Checksum verify fails in install script | Asset names match `bareai_<ver>_<os>_<arch>.tar.gz` template |

# homebrew-tap

Separate Homebrew tap repository for `bareai`.

GoReleaser publishes `Formula/bareai.rb` here on each release when `HOMEBREW_TAP_GITHUB_TOKEN` is configured.

## One-time setup (maintainer)

1. Create a **public** GitHub repository named exactly `homebrew-tap` under your user/org:
   - URL: `https://github.com/baselhusam/homebrew-tap`
   - Homebrew maps `brew tap baselhusam/tap` → `github.com/baselhusam/homebrew-tap`

2. Push this folder as the initial contents:

   ```bash
   cd packaging/homebrew-tap
   git init
   git add README.md Formula/bareai.rb
   git commit -m "Add bareai v0.1.0 formula"
   git branch -M main
   git remote add origin https://github.com/baselhusam/homebrew-tap.git
   git push -u origin main
   ```

3. On `bareai-cli`, add GitHub Actions secret `HOMEBREW_TAP_GITHUB_TOKEN`:
   - Fine-grained PAT with **Contents: read/write** on `homebrew-tap`, or
   - Classic PAT with `repo` scope

4. Future releases: tag `bareai-cli` → GoReleaser updates `Formula/bareai.rb` automatically.

## User install

```bash
brew tap baselhusam/tap
brew trust baselhusam/tap
brew install bareai
bareai version
```

## Formula maintenance

Do not edit `Formula/bareai.rb` manually after GoReleaser is wired up — it is updated from [`.goreleaser.yaml`](../../.goreleaser.yaml) on each tag push.

The copy in this folder is a bootstrap for v0.1.0 only.

# homebrew-tap

Separate Homebrew tap repository for `bareai`.

GoReleaser publishes `Formula/bareai.rb` here on each release when `HOMEBREW_TAP_GITHUB_TOKEN` is configured.

## Setup (one-time)

1. Create GitHub repository: `https://github.com/baselhusam/homebrew-tap`
2. Add an empty commit or this README
3. Create GitHub secret `HOMEBREW_TAP_GITHUB_TOKEN` on `bareai-cli` with a PAT that has `repo` scope on `homebrew-tap`

## User install

```bash
brew tap baselhusam/tap
brew install bareai
bareai version
```

## Formula maintenance

Do not edit `Formula/bareai.rb` manually — GoReleaser updates it from [`.goreleaser.yaml`](../../.goreleaser.yaml) on tag push.

# typed: false
# frozen_string_literal: true

# This file is published by GoReleaser on each release.
# Bootstrap copy for v0.1.0 — push to https://github.com/baselhusam/homebrew-tap

class Bareai < Formula
  desc "CLI and TUI for inspecting bare-metal AI infrastructure"
  homepage "https://github.com/baselhusam/bareai-cli"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/bareai_0.1.0_darwin_amd64.tar.gz"
      sha256 "e5a451b1bd9ed1328adca0d4083913ee3b243f842ee2848c9fb887995b0ecae2"
    end
    on_arm do
      url "https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/bareai_0.1.0_darwin_arm64.tar.gz"
      sha256 "50e918a8bd11685b7bf087f0a65fc9ec9d4441e705882d2e60285621fa449b7d"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/bareai_0.1.0_linux_amd64.tar.gz"
      sha256 "ac67162d34d693bc956b2d67ca90d1e2c102c3635a5376e7b45787c21e24c3d0"
    end
    on_arm do
      url "https://github.com/baselhusam/bareai-cli/releases/download/v0.1.0/bareai_0.1.0_linux_arm64.tar.gz"
      sha256 "2547119d52b43bf916ad07a2c86bbe2b240f0025e55e3c8e15567e05fd441dbf"
    end
  end

  def install
    bin.install "bareai"
  end

  test do
    system "#{bin}/bareai", "version"
  end
end

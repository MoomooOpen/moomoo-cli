# moomoo CLI Installation Guide

This document covers the standalone `moomoo` CLI installation, release, update, and uninstall flow.

## GitHub Release Source

The default GitHub Release source is built in:

```text
https://github.com/MoomooOpen/moomoo-cli
```

The CLI resolves release metadata from:

```text
{release_base}/releases/latest
{release_base}/releases/download/{version}/{asset}
```

Set `MOOMOO_CLI_RELEASE_BASE` only when testing a fork or private release source.

## macOS / Linux

Recommended installation:

```bash
curl -fsSL https://raw.githubusercontent.com/MoomooOpen/moomoo-cli/main/scripts/install.sh | sh
```

To verify the installer from a local checkout:

```bash
sh scripts/install.sh
```

Default install path:

```text
~/.moomoo/bin/moomoo
```

The installer adds `~/.moomoo/bin` to the current user's shell profile when needed. After the first installation, restart the terminal or run the `source` command printed by the installer, then verify:

```bash
moomoo version
```

To install into a custom directory:

```bash
INSTALL_DIR=/usr/local/bin sh scripts/install.sh
```

Linux downloads the musl build by default. To install the glibc build:

```bash
MOOMOO_CLI_LIBC=glibc sh scripts/install.sh
```

To install a specific version, for example a prerelease or temporary test tag:

```bash
MOOMOO_CLI_VERSION=v0.1.0-test.1 sh scripts/install.sh
```

## Windows

Recommended installation:

```powershell
iwr -UseBasicParsing https://raw.githubusercontent.com/MoomooOpen/moomoo-cli/main/scripts/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

To verify a specific version from a local checkout:

```powershell
$env:MOOMOO_CLI_VERSION="v0.1.0-test.1"
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

Default install path:

```text
%LOCALAPPDATA%\MoomooCLI\bin\moomoo.exe
```

## npm

npm package name:

```text
@moomoo/moomoo-cli
```

Install from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

After the package is published to the npm registry:

```bash
npm install -g @moomoo/moomoo-cli
```

The npm wrapper version must match the Git tag. For example, npm version `0.1.0` downloads GitHub Release `v0.1.0`.

To test the local npm wrapper against a temporary release:

```bash
MOOMOO_CLI_VERSION=v0.1.0-test.1 npm install -g .
```

## Release Flow

Releases are published with the manual GitHub Actions `Release` workflow. The workflow is intended for stable releases.

The release version can be provided manually, for example `v0.1.0`. If omitted, the workflow reads the root `package.json` version and prefixes it with `v`, so npm version `0.1.0` publishes tag `v0.1.0`.

The workflow runs:

```text
go test ./...
scripts/build-release.sh <version>
git tag -a <version> <current-commit>
git push origin <version>
gh release create <version> dist/* --verify-tag
```

If the tag already exists, the workflow verifies that it points to the current commit. If it points elsewhere, publishing stops so source tags and binary assets cannot drift.

To verify release downloads:

```bash
sh scripts/install.sh
```

## Verification

```bash
moomoo version
moomoo update --check
moomoo config init
moomoo auth login
moomoo auth status
```

## Update

```bash
moomoo update
moomoo update --check
moomoo update --force
moomoo update --release-notes
```

If installed from GitHub through npm, reinstall from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

After npm registry publication:

```bash
npm update -g @moomoo/moomoo-cli
```

## Uninstall

macOS / Linux:

```bash
rm -f ~/.moomoo/bin/moomoo
```

Windows:

```powershell
Remove-Item "$env:LOCALAPPDATA\MoomooCLI\bin\moomoo.exe" -Force
```

npm:

```bash
npm uninstall -g @moomoo/moomoo-cli
```

To remove local configuration and tokens as well:

```bash
rm -rf ~/.moomoo
```

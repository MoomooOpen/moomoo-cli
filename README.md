# moomoo-cli

moomoo CLI is a command-line tool for moomoo OpenAPI workflows, including OAuth login, market data queries, trading operations, crypto trading, configuration, and self-update.

## Install

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/MoomooOpen/moomoo-cli/main/scripts/install.sh | sh
```

Restart the terminal after installation, then verify:

```bash
moomoo version
```

Windows PowerShell:

```powershell
iwr -UseBasicParsing https://raw.githubusercontent.com/MoomooOpen/moomoo-cli/main/scripts/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

npm:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

After the package is published to the npm registry, this will also be available:

```bash
npm install -g @futunn/moomoo-cli
```

## Quick Start

Initialize configuration and sign in:

```bash
moomoo config init
```

Check authentication and API connectivity:

```bash
moomoo auth status
moomoo check
```

Run a quote query:

```bash
moomoo quote AAPL.US
```

## Common Commands

```bash
moomoo account list --format json
moomoo account funds --account <account-id> --currency USD
moomoo position --account <account-id> --format json
moomoo order list --account <account-id> --market US
moomoo order history --account <account-id> --market US --symbol US.AAPL
moomoo deal history --account <account-id> --market US --start <microseconds> --end <microseconds>
moomoo crypto account list --format json
moomoo crypto balance --account <account-id> --format json
moomoo crypto order history --account <account-id> --symbol BTCUSD --format json
```

Trading commands such as `order buy`, `order sell`, `order modify`, and `order cancel` operate on real accounts and require local confirmation before submission.

## Update

Check whether a newer stable GitHub Release is available:

```bash
moomoo update --check
```

Apply the latest stable release:

```bash
moomoo update
```

If installed from GitHub through npm, reinstall from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

After the package is published to the npm registry, use:

```bash
npm update -g @futunn/moomoo-cli
```

## Configuration

List configuration:

```bash
moomoo config list
```

Common keys:

```text
api_base
client_id
redirect_uri
default_account
auto_update
```

## Release Assets

Official releases are published from GitHub Actions. The release tag, GitHub Release, binary version, and asset names must use the same version, for example `v0.1.0`.

Each release includes platform archives and one checksum file:

```text
moomoo_v0.1.0_darwin_amd64.tar.gz
moomoo_v0.1.0_darwin_arm64.tar.gz
moomoo_v0.1.0_linux_amd64.tar.gz
moomoo_v0.1.0_linux_arm64.tar.gz
moomoo_v0.1.0_linux_musl_amd64.tar.gz
moomoo_v0.1.0_linux_musl_arm64.tar.gz
moomoo_v0.1.0_windows_amd64.zip
moomoo_checksums.txt
```

## Development

```bash
go test ./...
scripts/build-release.sh v0.1.0
```

## License

MIT

# moomoo-cli

moomoo CLI is a command-line tool for moomoo OpenAPI workflows, including OAuth login, market data queries, trading operations, crypto trading, configuration, and self-update.

## Install

The shell installers are the recommended path because they do not require Node.js.

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
npm install -g github:MoomooOpen/moomoo-cli --install-links=true
```

The npm wrapper requires Node.js 18 or newer and npm 9 or newer. `--install-links=true` is recommended for GitHub installs so npm copies the packed package instead of linking to a temporary git checkout.

The scoped npm registry package is not published yet. After it is published, this will also be available:

```bash
npm install -g @moomoo/moomoo-cli
```

If npm reports `spawn sh ENOENT`, configure the lifecycle script shell:

```bash
npm config set script-shell /bin/sh
```

If npm reports `uv_cwd` or `getcwd`, run the install from an existing directory:

```bash
cd ~
npm install -g github:MoomooOpen/moomoo-cli --install-links=true
```

If you use nvm, make sure `node` and `npm` come from the same Node installation before installing:

```bash
which node
which npm
node -v
npm -v
npm prefix -g
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

Query market data:

```bash
moomoo quote HK.00700 US.AAPL
moomoo kline US.AAPL --ktype 2 --format json
```

List authorized trading accounts:

```bash
moomoo account list
```

## Common Commands

```bash
moomoo quote HK.00700 US.AAPL --format json
moomoo snapshot HK.00700 US.AAPL --format json
moomoo broker HK.00700 --view history --days-before 5 --format json
moomoo capital HK.00700 --view distribution --format json
moomoo option chain US.AAPL --format json
moomoo account list --format json
moomoo account funds --account <account-id> --currency USD
moomoo position --account <account-id> --format json
moomoo order list --account <account-id> --market US
moomoo order history --account <account-id> --market US --symbol US.AAPL
moomoo deal history --account <account-id> --market US --start <microseconds> --end <microseconds>
moomoo sim account list --format json
moomoo sim order buy US.AAPL 2 --market US --type limit --price 180.5 --format json
moomoo crypto account list --format json
moomoo crypto balance --account <account-id> --format json
moomoo crypto order history --account <account-id> --symbol BTCUSD --format json
```

Securities and crypto order writes default to dry-run. The first call validates and previews the request, returns a three-digit confirmation code, and does not submit. Re-run the same command with `--execute <code>` to place, modify, or cancel.

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
npm install -g github:MoomooOpen/moomoo-cli --install-links=true
```

After the package is published to the npm registry, use:

```bash
npm update -g @moomoo/moomoo-cli
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
client_id_api_base
redirect_uri
default_account
auto_update
```

## Telemetry

Each API call is reported to `https://ai-report.moomoo.com` so a failed or slow command can be traced back to your account without asking you to copy logs. The report is sent with the same OAuth access token as the call, which lets support match it to the session that just ran, see whether the CLI or an agent sent an invalid request, and tell you if your installed version is out of date.

A report includes the command name, HTTP method, API path, status code, duration, CLI version, operating system, architecture, agent name, and the sanitized request and response. Token, password, and secret fields are removed. Securities, account, and crypto trading payloads are masked. Quote and other calls keep the remaining request and response fields, including order IDs, so a specific request can be identified when you ask for help. There is no setting to turn reporting off.

## Command Reference

This section summarizes the supported authentication, market data, securities trading, and cryptocurrency trading commands. Use `--format json` when consuming command output from scripts.

### Authentication

| Command | Purpose | Notes |
| --- | --- | --- |
| `moomoo auth login` | Start OAuth login. | Reuses a valid local token, refreshes when possible, or opens browser authorization when login is required. |
| `moomoo auth login --no-browser` | Log in without opening a local browser. | Prints the authorization URL and waits for the callback URL to be pasted back into the terminal. |
| `moomoo auth status` | Show local login and permission status. | Includes token status, expiry, account metadata, and quote or trade permission status. |
| `moomoo auth logout` | Remove the stored OAuth token. | The next authenticated command will require login again. |
| `moomoo auth register` | Register a new OAuth client. | Normal login reuses the existing client; run only when client registration must be refreshed. |
| `moomoo check` | Verify token validity and API connectivity. | Checks the current API base and token state. |

### Market Data

Quote commands require quote permission. Securities symbols use `MARKET.CODE`, such as `HK.00700` and `US.AAPL`. Hong Kong numeric codes are normalized to five digits, so `HK.700` is sent as `HK.00700`. `moomoo search [keyword]` is the exception: it accepts a plain search keyword such as `AAPL` or `renewable energy`, and the CLI sends that keyword through the backend `symbol` query field without symbol-format validation.

| Command | Purpose | Key options |
| --- | --- | --- |
| `moomoo quote <symbol...>` | Get real-time quotes for one or more securities. | Add `--format json` for scripts. |
| `moomoo basicinfo <symbol...>` | Get static symbol information in batch. | Add `--format json` for complete fields. |
| `moomoo snapshot <symbol...>` | Get market snapshots in batch. | Add `--format json` for complete fields. |
| `moomoo kline <symbol>` | Get the latest K-line bars. | `--num`, `--ktype`, `--autype`, `--extended-time`, `--format json` |
| `moomoo history-kline <symbol>` | Get historical K-line bars by date range or cursor. | `--start`, `--end`, `--next-time`, `--num`, `--format json` |
| `moomoo ticker <symbol>` | Get latest tick-by-tick trades. | `--num`, repeatable `--period`, `--format json` |
| `moomoo rt-data <symbol>` | Get intraday time-share data. | `--section`, `--format json` |
| `moomoo order-book <symbol>` | Get real-time order book depth. | `--num`, `--format json`; `depth` is an alias. |
| `moomoo rank [type]` | Show market rankings. | `volume`, `turnover`, `gainers`, `losers`, `amplitude`; `--market`, `--count` |
| `moomoo option chain <symbol>` | Query an underlying's option chain. | `--start`, `--end`, `--index-option-type`, `--filter-standard` |
| `moomoo search [keyword]` | Search news, community content, or stock feeds by keyword. | `--type news/community/feed`, `--keyword`, `--size`, `--lang`, `--format json` |

Additional quote REST commands include aggregated views such as `broker`, `capital`, `company`, `corporate-actions`, `financials`, `option-data`, `plate`, `research`, `shareholders`, `short`, `watchlist`, and `valuation`, plus standalone screen/list commands such as `stock-screen`, `warrant-screen`, `option-screen`, `future-info`, `market-state`, and `trading-days`. See `.local_docs/moomoo-cli/docs/quote-cli-guide.md` for the full Quote command guide.

### Securities Accounts and Positions

| Command | Purpose | Key options |
| --- | --- | --- |
| `moomoo account list` | List authorized trading accounts and available markets. | Add `--format json` for scripts. |
| `moomoo account funds --account <account-id>` | Show cash, total assets, market value, available funds, and buying power. | `--currency`, `--format json` |
| `moomoo position --account <account-id>` | List positions with quantity, cost price, market value, and profit or loss. | `--symbol`, `--pl-ratio-min`, `--pl-ratio-max`, `--format json` |

### Securities Orders and Trades

Securities order writes default to dry-run. Preview first, then re-run with `--execute <confirmation_code>` to submit.

| Command | Purpose | Key options |
| --- | --- | --- |
| `moomoo order buy <symbol> <qty>` | Preview a securities buy order; submit with `--execute`. | `--account`, `--price`, `--type`, `--time-in-force`, `--lot-type`, `--session`, `--remark`, `--execute` |
| `moomoo order sell <symbol> <qty>` | Preview a securities sell order; submit with `--execute`. | Same core order options as `order buy`. |
| `moomoo order place <qty>` | Preview advanced securities orders; submit with `--execute`. | `--symbol`, `--side`, `--order-class`, `--multi-leg-info`; supports `BUY`, `SELL`, `SELL_SHORT`, `BUY_BACK`, conditional, and multi-leg orders. |
| `moomoo order modify <order-id>` | Preview quantity, price, or trigger price changes; submit with `--execute`. | `--account`, `--exchange`, `--qty`, `--price`, `--aux-price`, `--execute` |
| `moomoo order cancel <order-id>` | Preview cancel; submit with `--execute`. | `--account`, `--exchange`, `--execute` |
| `moomoo order list --account <account-id> --market <market>` | List open orders and recent related orders. | `--page-size`, `--format json` |
| `moomoo order history --account <account-id> --market <market>` | List historical orders. | `--symbol`, `--start`, `--end`, `--page-size`, `--format json` |
| `moomoo order detail <order-id> [order-id...]` | Show detailed order information for up to 100 order IDs. | `--account`, `--exchange`, `--format json` |
| `moomoo order max-qty <symbol>` | Query maximum cash buy, margin buy, position sell, short sell, and buy-back quantities. | `--account`, `--type`, `--price`, `--order-id`, `--format json` |
| `moomoo deal --account <account-id> --market <market>` | List today's trades. | `--page-size`, `--format json` |
| `moomoo deal history --account <account-id> --market <market>` | List historical trades. | `--symbol`, `--start`, `--end`, `--page-size`, `--format json` |

```bash
moomoo order buy US.AAPL 1 --type limit --price 180
moomoo order buy US.AAPL 1 --type limit --price 180 --execute 203
```

### Simulated Trading

Simulated trading commands use a dedicated simulated account ID. If `--account` is omitted, the CLI uses the first simulated account returned by the account-list endpoint. Simulated orders do not use root-level `sim-trade-*` commands; use the grouped `sim` commands below.

| Command | Purpose | Key options |
| --- | --- | --- |
| `moomoo sim account list` | List simulated trading accounts. | Add `--format json` for scripts. |
| `moomoo sim account funds` | Show simulated account funds and buying power. | `--account`, `--format json` |
| `moomoo sim position` | List simulated positions. | `--account`, `--market`, `--format json` |
| `moomoo sim order buy <symbol> <qty>` | Submit a simulated buy order. | `--account`, `--market`, `--type`, `--price`, `--session`, `--text` |
| `moomoo sim order sell <symbol> <qty>` | Submit a simulated sell order. | Same core order options as `sim order buy`. |
| `moomoo sim order place <qty>` | Submit a generic simulated order. | `--symbol`, `--side`, `--account`, `--market`, `--type`, `--price` |
| `moomoo sim order modify <order-id>` | Modify a simulated order. | `--account`, `--market`, required `--qty`, required `--price` |
| `moomoo sim order cancel <order-id>` | Cancel a simulated order. | `--account`, `--market`, `--format json` |
| `moomoo sim order list` | List open simulated orders. | `--account`, `--market`, `--format json` |
| `moomoo sim order history` | List historical simulated orders. | `--account`, `--market`, `--start`, `--end`, `--page-size`, `--next-key` |
| `moomoo sim order max-qty <symbol>` | Query simulated maximum tradable quantity. | `--account`, `--market`, `--type`, `--price`, `--order-id` |

See `.local_docs/moomoo-cli/docs/sim-trading-cli-guide.md` for the full simulated trading command guide.

### Cryptocurrency Trading

| Command | Purpose | Key options |
| --- | --- | --- |
| `moomoo crypto account list` | List authorized cryptocurrency trading accounts. | Add `--format json` for scripts. |
| `moomoo crypto balance --account <account-id>` | Show total crypto account assets, including cash balances and digital asset positions. | `--format json` |
| `moomoo crypto max-qty <symbol>` | Query maximum cash-buy quantity, sell quantity, and cash-buy amount for a cryptocurrency pair. | `--account`, `--type`, `--currency`, `--coin`, `--price`, `--format json` |
| `moomoo crypto order place <symbol>` | Preview a cryptocurrency order with explicit side; submit with `--execute`. | `--side`, `--account`, `--req-id`, `--type`, `--time-in-force`, `--price`, `--qty`, `--cash-qty`, `--conditional-info`, `--execute` |
| `moomoo crypto order buy <symbol>` | Preview a cryptocurrency buy order; submit with `--execute`. | Same core order options as `crypto order place`, without `--side`. |
| `moomoo crypto order sell <symbol>` | Preview a cryptocurrency sell order; submit with `--execute`. | Same core order options as `crypto order place`, without `--side`; sell does not support `--cash-qty`. |
| `moomoo crypto order modify <order-id>` | Preview a crypto order modification; submit with `--execute`. | `--account`, `--req-id`, `--price`, `--qty`, `--order-version`, `--conditional-info`, `--execute` |
| `moomoo crypto order cancel <order-id>` | Preview cancel; submit with `--execute`. | `--account`, `--req-id`, `--execute`, `--format json` |
| `moomoo crypto order list` | List open cryptocurrency orders. | `--account`, `--page-size`, `--page-flag`, `--format json` |
| `moomoo crypto order history` | List historical cryptocurrency orders. | `--account`, `--symbol`, `--start`, `--end`, `--order-status`, `--currency`, `--side`, `--type`, `--page-size`, `--page-flag`, `--format json` |
| `moomoo crypto order detail <order-id> [order-id...]` | Show details for up to 20 cryptocurrency order IDs. | `--account`, `--format json` |
| `moomoo crypto deal <order-id>` | List fills for a cryptocurrency order. | `--account`, `--page-size`, `--page-flag`, `--group-by-order`, `--format json` |
| `moomoo crypto deal history` | List historical cryptocurrency fills. | `--account`, `--symbol`, `--start`, `--end`, `--page-size`, `--page-flag`, `--group-by-order`, `--format json` |

### Common Trading Rules

Most trading commands accept `--account` and fall back to `default_account` when configured. Commands that list historical orders or fills accept microsecond timestamp filters.

Securities symbol arguments and `--symbol` filters use `MARKET.CODE`, such as `HK.00700` and `US.AAPL`. Hong Kong numeric codes are normalized to five digits, so `HK.700` is sent as `HK.00700`. Cryptocurrency commands use trading pairs such as `BTCUSD` and `ETHHKD`.

Order placement, modification, and cancellation on real securities and crypto accounts default to dry-run. Use `--execute <confirmation_code>` to submit. Simulated trading (`sim order`) is not gated.

## Release Assets

Official releases are published from GitHub Actions. The release tag, GitHub Release, binary version, and asset names must use the same version, for example `v0.1.0`.

Each release includes platform archives, a checksum file, and an ECDSA signature over the checksum file:

```text
moomoo_v0.1.0_darwin_amd64.tar.gz
moomoo_v0.1.0_darwin_arm64.tar.gz
moomoo_v0.1.0_linux_amd64.tar.gz
moomoo_v0.1.0_linux_arm64.tar.gz
moomoo_v0.1.0_linux_musl_amd64.tar.gz
moomoo_v0.1.0_linux_musl_arm64.tar.gz
moomoo_v0.1.0_windows_amd64.zip
moomoo_checksums.txt
moomoo_checksums.txt.sig
```

Installers and `moomoo update` verify `moomoo_checksums.txt.sig` with the embedded release public key before trusting checksums or replacing the binary. Archives also include a `VERSION` file that is checked without executing the downloaded binary.

## Development

```bash
go test ./...
scripts/build-release.sh v0.1.0
```

## License

MIT

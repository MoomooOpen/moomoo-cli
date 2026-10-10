# @moomoo/moomoo-cli

npm wrapper for the `moomoo` CLI.

```bash
npm install -g github:MoomooOpen/moomoo-cli
moomoo version
```

Direct npm registry installation is not available yet and is still in progress:

```bash
npm install -g @moomoo/moomoo-cli
```

Update a GitHub npm installation by reinstalling from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

Direct npm registry updates are not available yet and are still in progress:

```bash
npm update -g @moomoo/moomoo-cli
```

The default GitHub Release source is `https://github.com/MoomooOpen/moomoo-cli`.
Set `MOOMOO_CLI_RELEASE_BASE` only when testing a fork or private release source.

To test a temporary prerelease:

```bash
MOOMOO_CLI_VERSION=v0.1.0-test.1 npm install -g .
```

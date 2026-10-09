# @moomoo/moomoo-cli

npm wrapper for the `moomoo` CLI.

```bash
npm install -g github:MoomooOpen/moomoo-cli
moomoo version
```

After the package is published to the npm registry:

```bash
npm install -g @moomoo/moomoo-cli
```

Update a GitHub npm installation by reinstalling from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli
```

After npm registry publication:

```bash
npm update -g @moomoo/moomoo-cli
```

The default GitHub Release source is `https://github.com/MoomooOpen/moomoo-cli`.
Set `MOOMOO_CLI_RELEASE_BASE` only when testing a fork or private release source.

To test a temporary prerelease:

```bash
MOOMOO_CLI_VERSION=v0.1.0-test.1 npm install -g .
```

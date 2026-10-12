# @moomoo/moomoo-cli

npm wrapper for the `moomoo` CLI.

```bash
npm install -g github:MoomooOpen/moomoo-cli --install-links=true
moomoo version
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

Update a GitHub npm installation by reinstalling from GitHub:

```bash
npm install -g github:MoomooOpen/moomoo-cli --install-links=true
```

After the package is published to the npm registry, use:

```bash
npm update -g @moomoo/moomoo-cli
```

The default GitHub Release source is `https://github.com/MoomooOpen/moomoo-cli`.
Set `MOOMOO_CLI_RELEASE_BASE` only when testing a fork or private release source.

To test a temporary prerelease:

```bash
MOOMOO_CLI_VERSION=v0.1.0-test.1 npm install -g .
```

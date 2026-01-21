# Website

> **⚠️ DEPRECATED**: This docs-site has been moved to a dedicated repository.
>
> **New Location**: [aura-blockchain/docs](https://github.com/aura-blockchain/docs)
>
> **Live Documentation**: https://docs.aurablockchain.org
>
> This directory is preserved for historical reference only. All new documentation
> should be added to the [aura-blockchain/docs](https://github.com/aura-blockchain/docs) repository.

---

*Legacy README below for reference:*

This website is built using [Docusaurus](https://docusaurus.io/), a modern static website generator.

## Installation

```bash
yarn
```

## Local Development

```bash
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```bash
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

Using SSH:

```bash
USE_SSH=true yarn deploy
```

Not using SSH:

```bash
GIT_USER=<Your GitHub username> yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.

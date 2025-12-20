# Aura Documentation Site

This directory contains the Docusaurus-based documentation website for the Aura blockchain.

## Local Development

```bash
# Install dependencies
npm install

# Start development server
npm start

# This command starts a local development server and opens up a browser window.
# Most changes are reflected live without having to restart the server.
```

## Build

```bash
# Generate static content
npm run build

# This command generates static content into the `build` directory
# and can be served using any static contents hosting service.
```

## Deployment

The documentation site is automatically deployed to GitHub Pages when changes are pushed to the `main` branch in the `docs-site/` directory.

The deployment is handled by the `.github/workflows/deploy-docs.yml` workflow.

## Structure

- `docs/` - Documentation content in Markdown
  - `getting-started/` - Installation and quick start guides
  - `developers/` - Developer guides and API references
  - `validators/` - Validator setup and operations
- `src/` - React components and custom pages
- `static/` - Static assets (images, etc.)
- `docusaurus.config.ts` - Site configuration
- `sidebars.ts` - Sidebar navigation configuration

## Contributing

To add new documentation:

1. Create a new `.md` file in the appropriate `docs/` subdirectory
2. Add frontmatter with `sidebar_position` if needed
3. Update `sidebars.ts` if creating a new section
4. Test locally with `npm start`
5. Build and verify with `npm run build`
6. Commit and push changes

## More Information

See the [Docusaurus documentation](https://docusaurus.io/) for more details on customization and advanced features.

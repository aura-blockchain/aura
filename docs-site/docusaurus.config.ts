import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'Aura Documentation',
  tagline: 'Privacy-preserving identity and DeFi infrastructure',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://docs.aura.network',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'aura-blockchain', // Usually your GitHub org/user name.
  projectName: 'aura', // Usually your repo name.

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/aura-blockchain/aura/tree/main/docs-site/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/docusaurus-social-card.jpg',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Aura',
      logo: {
        alt: 'Aura Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          type: 'docSidebar',
          sidebarId: 'developersSidebar',
          position: 'left',
          label: 'Developers',
        },
        {
          type: 'docSidebar',
          sidebarId: 'validatorsSidebar',
          position: 'left',
          label: 'Validators',
        },
        {
          href: 'https://github.com/aura-blockchain/aura',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/intro',
            },
            {
              label: 'Developers',
              to: '/docs/developers/overview',
            },
            {
              label: 'Validators',
              to: '/docs/validators/setup',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Discord',
              href: 'https://discord.gg/aura',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/AuraNetwork',
            },
            {
              label: 'Forum',
              href: 'https://forum.aura.network',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/aura-blockchain/aura',
            },
            {
              label: 'Block Explorer',
              href: 'https://explorer.aura.network',
            },
            {
              label: 'Testnet Faucet',
              href: 'https://faucet.aura.network',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Aura Blockchain. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

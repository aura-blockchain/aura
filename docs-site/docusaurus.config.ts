import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'AURA Blockchain',
  tagline: 'Decentralized Credential Verification',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://aura-blockchain.github.io',
  baseUrl: '/aura/',

  organizationName: 'aura-blockchain',
  projectName: 'aura',

  onBrokenLinks: 'throw',

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
          editUrl: 'https://github.com/aura-blockchain/aura/tree/main/docs-site/',
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          editUrl: 'https://github.com/aura-blockchain/aura/tree/main/docs-site/',
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/aura-social-card.jpg',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'AURA',
      logo: {
        alt: 'AURA Blockchain Logo',
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
          to: '/docs/api',
          label: 'API Reference',
          position: 'left',
        },
        {to: '/blog', label: 'Blog', position: 'left'},
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
              label: 'API Reference',
              to: '/docs/api',
            },
            {
              label: 'Validator Guide',
              to: '/docs/validators',
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
              href: 'https://twitter.com/useyouraura',
            },
            {
              label: 'Telegram',
              href: 'https://t.me/aurachain',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'Blog',
              to: '/blog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/aura-blockchain/aura',
            },
            {
              label: 'Testnet Explorer',
              href: 'https://testnet-explorer.aurablockchain.org',
            },
            {
              label: 'Testnet Faucet',
              href: 'https://testnet-faucet.aurablockchain.org',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} AURA Blockchain. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'go', 'json', 'protobuf', 'toml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/installation',
        'getting-started/quick-start',
      ],
    },
    {
      type: 'category',
      label: 'Tutorials',
      items: [
        'tutorials/index',
        'tutorials/create-did',
        'tutorials/issue-credential',
        'tutorials/stake-tokens',
        'tutorials/governance',
        'tutorials/dex-trading',
      ],
    },
  ],

  developersSidebar: [
    'developers/overview',
    {
      type: 'category',
      label: 'Building on Aura',
      items: [
        'developers/module-development',
        'developers/sdk-integration',
      ],
    },
  ],

  modulesSidebar: [
    'modules/index',
    {
      type: 'category',
      label: 'Module Reference',
      items: [
        'modules/identity',
        'modules/privacy',
        'modules/defi',
        'modules/security',
        'modules/governance',
        'modules/infrastructure',
      ],
    },
  ],

  validatorsSidebar: [
    'validators/setup',
    {
      type: 'category',
      label: 'Operations',
      items: [
        'validators/monitoring',
        'validators/upgrades',
      ],
    },
  ],
};

export default sidebars;

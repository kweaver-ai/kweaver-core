import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'KWeaver',
  tagline: 'Open-Source Decision Intelligence AI Ecosystem',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://kweaver-ai.github.io',
  baseUrl: '/kweaver-core/',

  organizationName: 'kweaver-ai',
  projectName: 'kweaver-core',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: false,
        blog: {
          routeBasePath: '/',
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          editUrl: 'https://github.com/kweaver-ai/kweaver/tree/main/website/',
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
    image: 'img/logo-light.png',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'KWeaver',
      logo: {
        alt: 'KWeaver Logo',
        src: 'img/logo-light.png',
        srcDark: 'img/logo-dark.png',
      },
      items: [
        {
          href: 'https://github.com/kweaver-ai/kweaver',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub Issues',
              href: 'https://github.com/kweaver-ai/kweaver/issues',
            },
            {
              label: 'GitHub Discussions',
              href: 'https://github.com/kweaver-ai/kweaver/discussions',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/kweaver-ai/kweaver',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} KWeaver. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

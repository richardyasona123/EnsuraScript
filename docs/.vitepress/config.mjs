import { defineConfig } from 'vitepress'
import { readFileSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

const __dirname = dirname(fileURLToPath(import.meta.url))

// Load EnsuraScript grammar
const ensGrammar = JSON.parse(
  readFileSync(resolve(__dirname, '../../editors/shared/ensurascript.tmLanguage.json'), 'utf-8')
)

export default defineConfig({
  title: 'EnsuraScript',
  description: 'Programming by guarantees, not instructions',
  base: '/EnsuraScript/',
  ignoreDeadLinks: true,

  appearance: 'dark', // Default to dark mode with toggle
  lastUpdated: true,

  markdown: {
    languages: [ensGrammar],
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    },
    lineNumbers: true
  },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#2563eb' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:title', content: 'EnsuraScript' }],
    ['meta', { name: 'og:description', content: 'Programming by guarantees, not instructions' }]
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Learn', link: '/learn/' },
      { text: 'Reference', link: '/reference/syntax' },
      { text: 'Examples', link: '/examples/' },
      { text: 'GitHub', link: 'https://github.com/GustyCube/EnsuraScript' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is EnsuraScript?', link: '/guide/' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Core Concepts', link: '/guide/core-concepts' }
          ]
        },
        {
          text: 'Features',
          items: [
            { text: 'Resources', link: '/guide/resources' },
            { text: 'Guarantees', link: '/guide/guarantees' },
            { text: 'Handlers', link: '/guide/handlers' },
            { text: 'Policies', link: '/guide/policies' }
          ]
        }
      ],
      '/learn/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Overview', link: '/learn/' },
            { text: 'Installation', link: '/learn/installation' },
            { text: 'Your First Guarantee', link: '/learn/first-guarantee' }
          ]
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Understanding Resources', link: '/learn/resources' },
            { text: 'Writing Guarantees', link: '/learn/guarantees' },
            { text: 'Using Handlers', link: '/learn/handlers' }
          ]
        },
        {
          text: 'Advanced Features',
          items: [
            { text: 'Creating Policies', link: '/learn/policies' },
            { text: 'Guards & Conditions', link: '/learn/guards' },
            { text: 'Dependencies', link: '/learn/dependencies' },
            { text: 'Collections & Invariants', link: '/learn/collections' },
            { text: 'Violation Handling', link: '/learn/violations' }
          ]
        },
        {
          text: 'Deep Dives',
          items: [
            { text: 'Implication System', link: '/learn/implications' },
            { text: 'Execution Model', link: '/learn/execution' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Language Reference',
          items: [
            { text: 'Syntax', link: '/reference/syntax' },
            { text: 'Conditions', link: '/reference/conditions' },
            { text: 'Handlers', link: '/reference/handlers' },
            { text: 'CLI Commands', link: '/reference/cli' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Overview', link: '/examples/' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/GustyCube/EnsuraScript' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 GustyCube'
    },

    search: {
      provider: 'local'
    }
  }
})

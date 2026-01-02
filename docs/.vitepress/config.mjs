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

  markdown: {
    languages: [ensGrammar]
  },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }]
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Reference', link: '/reference/syntax' },
      { text: 'Examples', link: '/examples/' },
      { text: 'GitHub', link: 'https://github.com/ensurascript/ensura' }
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
      '/reference/': [
        {
          text: 'Language Reference',
          items: [
            { text: 'Syntax', link: '/reference/syntax' },
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
      { icon: 'github', link: 'https://github.com/ensurascript/ensura' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024 EnsuraScript'
    },

    search: {
      provider: 'local'
    }
  }
})

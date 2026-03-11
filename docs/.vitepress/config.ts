import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Deployer',
  description: 'Simple deployment platform for your applications',
  head: [
    ['meta', { name: 'theme-color', content: '#4f46e5' }],
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API Reference', link: '/api/' },
      { text: 'CLI Reference', link: '/cli/' },
      { text: 'Pricing', link: '/pricing' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is Deployer?', link: '/guide/what-is-deployer' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Quick Start', link: '/guide/quick-start' },
          ]
        },
        {
          text: 'Deploying',
          items: [
            { text: 'Deploy with ZIP', link: '/guide/deploy-zip' },
            { text: 'Deploy with Git', link: '/guide/deploy-git' },
            { text: 'Deploy with Dockerfile', link: '/guide/deploy-dockerfile' },
            { text: 'Environment Variables', link: '/guide/env-vars' },
            { text: 'Custom Domains', link: '/guide/custom-domains' },
          ]
        },
        {
          text: 'Databases',
          items: [
            { text: 'Overview', link: '/guide/databases' },
            { text: 'PostgreSQL', link: '/guide/db-postgres' },
            { text: 'MySQL', link: '/guide/db-mysql' },
            { text: 'MongoDB', link: '/guide/db-mongodb' },
            { text: 'Redis', link: '/guide/db-redis' },
            { text: 'Backups', link: '/guide/db-backups' },
          ]
        },
        {
          text: 'Frameworks',
          items: [
            { text: 'Node.js / Express', link: '/guide/framework-nodejs' },
            { text: 'Python / Django', link: '/guide/framework-python' },
            { text: 'Go', link: '/guide/framework-go' },
            { text: 'Next.js', link: '/guide/framework-nextjs' },
            { text: 'Laravel', link: '/guide/framework-laravel' },
            { text: 'Static Sites', link: '/guide/framework-static' },
          ]
        },
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Authentication', link: '/api/' },
            { text: 'Apps', link: '/api/apps' },
            { text: 'Deployments', link: '/api/deployments' },
            { text: 'Databases', link: '/api/databases' },
            { text: 'Billing', link: '/api/billing' },
          ]
        }
      ],
      '/cli/': [
        {
          text: 'CLI Reference',
          items: [
            { text: 'Installation', link: '/cli/' },
            { text: 'Commands', link: '/cli/commands' },
          ]
        }
      ]
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/deployer' }
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 Deployer'
    }
  }
})

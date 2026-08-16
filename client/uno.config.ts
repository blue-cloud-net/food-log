import { defineConfig, presetUno } from 'unocss'

export default defineConfig({
  presets: [presetUno()],
  theme: {
    colors: {
      primary: 'var(--fl-primary)',
      'primary-light': 'var(--fl-primary-light)',
      bg: 'var(--fl-bg)'
    }
  },
  shortcuts: {
    'fl-card': 'bg-white rounded-xl shadow-sm p-4',
    'page-container': 'max-w-960px mx-auto p-4',
    'page-header': 'flex items-center justify-between mb-4',
    'page-title': 'text-xl font-semibold m-0',
    'section-title': 'text-lg font-semibold mb-3',
    'text-secondary': 'text-[#909399]',
    'text-tertiary': 'text-[#b0b3b8]'
  },
  safelist: [
    'text-primary',
    'bg-primary',
    'bg-primary/90',
    'border-primary'
  ]
})

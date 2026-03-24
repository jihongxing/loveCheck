import { createI18n } from 'vue-i18n'
import zh from './zh.js'

const supportedLocales = ['zh', 'en', 'ja', 'ko', 'th', 'vi', 'es', 'hi']

// Lazy loaders for non-default languages — only fetched when needed
const lazyLoaders = {
  en: () => import('./en.js'),
  ja: () => import('./ja.js'),
  ko: () => import('./ko.js'),
  th: () => import('./th.js'),
  vi: () => import('./vi.js'),
  es: () => import('./es.js'),
  hi: () => import('./hi.js'),
}

function detectLocale() {
  const saved = localStorage.getItem('loveTrust_locale')
  if (saved && supportedLocales.includes(saved)) return saved

  const browserLangs = navigator.languages || [navigator.language || 'en']
  for (const tag of browserLangs) {
    const primary = tag.toLowerCase().split('-')[0]
    if (supportedLocales.includes(primary)) return primary
  }

  return 'zh'
}

const detectedLocale = detectLocale()

const i18n = createI18n({
  legacy: false,
  locale: detectedLocale,
  fallbackLocale: 'zh',
  messages: { zh },
})

// Load the detected locale if it's not zh
export async function loadLocaleMessages(locale) {
  if (locale === 'zh') return
  if (i18n.global.availableLocales.includes(locale) && locale !== 'zh') {
    // Already loaded
    return
  }
  const loader = lazyLoaders[locale]
  if (!loader) return
  const messages = await loader()
  i18n.global.setLocaleMessage(locale, messages.default)
}

// Pre-load the detected locale on startup
if (detectedLocale !== 'zh') {
  loadLocaleMessages(detectedLocale)
}

export default i18n

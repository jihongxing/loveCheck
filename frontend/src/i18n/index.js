import { createI18n } from 'vue-i18n'
import zh from './zh.js'
import en from './en.js'
import ja from './ja.js'
import ko from './ko.js'
import th from './th.js'
import vi from './vi.js'
import es from './es.js'
import hi from './hi.js'

const supportedLocales = ['zh', 'en', 'ja', 'ko', 'th', 'vi', 'es', 'hi']

function detectLocale() {
  const saved = localStorage.getItem('loveTrust_locale')
  if (saved && supportedLocales.includes(saved)) return saved

  const browserLangs = navigator.languages || [navigator.language || 'en']
  for (const tag of browserLangs) {
    const primary = tag.toLowerCase().split('-')[0]
    if (supportedLocales.includes(primary)) return primary
  }

  return 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: { zh, en, ja, ko, th, vi, es, hi },
})

export default i18n

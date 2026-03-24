import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import i18n from './i18n/index.js'

createApp(App).use(i18n).mount('#app')

if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').then((reg) => {
      // Check for updates every 30 minutes
      setInterval(() => reg.update(), 30 * 60 * 1000)
    }).catch(() => {})
  })
}

// Umami analytics: set VITE_UMAMI_ID and VITE_UMAMI_SRC in .env to activate
;(function initUmami() {
  const id = import.meta.env.VITE_UMAMI_ID
  const src = import.meta.env.VITE_UMAMI_SRC
  if (!id || !src) return
  const s = document.createElement('script')
  s.defer = true
  s.setAttribute('data-website-id', id)
  s.src = src
  document.head.appendChild(s)
})()


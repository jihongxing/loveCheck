<template>
  <div class="app-container">
    <!-- Header -->
    <header class="header">
      <div class="lang-toggle">
        <button
          v-for="lang in langs"
          :key="lang.code"
          :class="['lang-btn', { active: locale === lang.code }]"
          @click="switchLang(lang.code)"
        >{{ lang.label }}</button>
      </div>
      <h1 class="logo text-gradient">{{ $t('app.title') }}</h1>
      <p class="subtitle">{{ $t('app.subtitle') }}</p>
    </header>

    <!-- Navigation Tabs -->
    <div class="nav-tabs">
      <button
        :class="['tab-btn', { active: currentTab === 'search' }]"
        @click="currentTab = 'search'">
        {{ $t('app.nav.search') }}
      </button>
      <button
        :class="['tab-btn', { active: currentTab === 'report' }]"
        @click="currentTab = 'report'">
        {{ $t('app.nav.report') }}
      </button>
      <button
        :class="['tab-btn', { active: currentTab === 'appeal' }]"
        @click="currentTab = 'appeal'">
        {{ $t('app.nav.appeal') }}
      </button>
    </div>

    <!-- Main Content Area -->
    <main class="content-area">
      <Search v-if="currentTab === 'search'" />
      <Report v-if="currentTab === 'report'" />
      <Appeal v-if="currentTab === 'appeal'" />
      <Admin v-if="currentTab === 'admin'" />
    </main>

    <!-- Footer -->
    <footer class="glass-footer">
      <div class="legal-links">
        <a href="#" class="legal-link" @click.prevent="showAgreement = true">{{ $t('app.footer.agreement') }}</a>
        <span class="legal-sep">|</span>
        <a href="#" class="legal-link" @click.prevent="showPrivacy = true">{{ $t('app.footer.privacy') }}</a>
        <span class="legal-sep">|</span>
        <a href="#" class="legal-link" @click.prevent="showContact = true">{{ $t('app.footer.contact') }}</a>
      </div>
      <p class="disclaimer" style="white-space: pre-line;">{{ $t('app.footer.disclaimer') }}</p>
    </footer>

    <!-- Privacy Policy Modal -->
    <transition name="fade">
      <div v-if="showPrivacy" class="modal-overlay" @click="showPrivacy = false">
        <div class="modal-content glass-panel" @click.stop>
          <button class="close-btn" @click="showPrivacy = false">&times;</button>
          <h3 class="modal-title">{{ $t('app.privacy.title') }}</h3>
          <div class="modal-body scroll-pane">
            <h4>{{ $t('app.privacy.s1_title') }}</h4>
            <p style="white-space:pre-line">{{ $t('app.privacy.s1') }}</p>
            <h4>{{ $t('app.privacy.s2_title') }}</h4>
            <p style="white-space:pre-line">{{ $t('app.privacy.s2') }}</p>
            <h4>{{ $t('app.privacy.s3_title') }}</h4>
            <p style="white-space:pre-line">{{ $t('app.privacy.s3') }}</p>
            <h4>{{ $t('app.privacy.s4_title') }}</h4>
            <p style="white-space:pre-line">{{ $t('app.privacy.s4') }}</p>
            <h4>{{ $t('app.privacy.s5_title') }}</h4>
            <p style="white-space:pre-line">{{ $t('app.privacy.s5') }}</p>
            <p style="text-align:center;margin-top:2rem;color:var(--text-secondary);font-size:0.8rem;">
              {{ $t('app.privacy.version') }}
            </p>
          </div>
          <button class="btn-premium" style="margin-top:1.5rem;width:100%" @click="showPrivacy = false">
            {{ $t('app.privacy.close') }}
          </button>
        </div>
      </div>
    </transition>

    <!-- Legal Agreement Fulltext Modal -->
    <transition name="fade">
      <div v-if="showAgreement" class="modal-overlay" @click="showAgreement = false">
        <div class="modal-content glass-panel" @click.stop>
          <button class="close-btn" @click="showAgreement = false">×</button>

          <h3 class="modal-title">{{ $t('app.agreement.title') }}</h3>
          <div class="modal-body scroll-pane">
            <h4>{{ $t('app.agreement.preamble_title') }}</h4>
            <p>{{ $t('app.agreement.preamble') }}</p>
            <p><strong style="color:#facc15">{{ $t('app.agreement.preamble_warn') }}</strong></p>

            <h4>{{ $t('app.agreement.art1_title') }}</h4>
            <p style="white-space: pre-line;">{{ $t('app.agreement.art1') }}</p>

            <h4>{{ $t('app.agreement.art2_title') }}</h4>
            <p style="white-space: pre-line;">{{ $t('app.agreement.art2') }}</p>

            <h4>{{ $t('app.agreement.art3_title') }}</h4>
            <p style="white-space: pre-line;">{{ $t('app.agreement.art3') }}</p>

            <p style="text-align: center; margin-top: 2rem; color: var(--text-secondary); font-size: 0.8rem;">
              {{ $t('app.agreement.version') }}
            </p>
          </div>

          <button class="btn-premium" style="margin-top:1.5rem; width:100%" @click="showAgreement = false">
            {{ $t('app.agreement.confirm') }}
          </button>
        </div>
      </div>
    </transition>

    <!-- Contact Us Modal -->
    <transition name="fade">
      <div v-if="showContact" class="modal-overlay" @click="showContact = false">
        <div class="modal-content glass-panel contact-modal" @click.stop>
          <button class="close-btn" @click="showContact = false">&times;</button>
          <h3 class="modal-title">{{ $t('app.contact.title') }}</h3>
          <div class="modal-body scroll-pane">
            <p class="contact-desc">{{ $t('app.contact.desc') }}</p>

            <div class="contact-channels">
              <!-- WeChat -->
              <div class="contact-card">
                <div class="contact-card-header">
                  <span class="contact-icon wechat-icon">W</span>
                  <span class="contact-label">{{ $t('app.contact.wechat') }}</span>
                </div>
                <img src="/contantus.jpg" alt="WeChat QR" class="contact-qr" />
                <p class="contact-hint">{{ $t('app.contact.wechat_hint') }}</p>
              </div>

              <!-- Email -->
              <div class="contact-card">
                <div class="contact-card-header">
                  <span class="contact-icon email-icon">@</span>
                  <span class="contact-label">{{ $t('app.contact.email') }}</span>
                </div>
                <a :href="'mailto:' + $t('app.contact.email_addr')" class="contact-value contact-link">
                  {{ $t('app.contact.email_addr') }}
                </a>
              </div>

              <!-- Telegram -->
              <div class="contact-card">
                <div class="contact-card-header">
                  <span class="contact-icon tg-icon">T</span>
                  <span class="contact-label">{{ $t('app.contact.telegram') }}</span>
                </div>
                <a href="https://t.me/LoverTrust202500" target="_blank" rel="noopener" class="contact-value contact-link">
                  {{ $t('app.contact.telegram_id') }}
                </a>
              </div>

              <!-- Working Hours -->
              <div class="contact-card">
                <div class="contact-card-header">
                  <span class="contact-icon hours-icon">H</span>
                  <span class="contact-label">{{ $t('app.contact.hours') }}</span>
                </div>
                <span class="contact-value">{{ $t('app.contact.hours_detail') }}</span>
              </div>
            </div>
          </div>
          <button class="btn-premium" style="margin-top:1.5rem;width:100%" @click="showContact = false">
            {{ $t('app.contact.close') }}
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Search from './components/Search.vue'
import Report from './components/Report.vue'
import Appeal from './components/Appeal.vue'
import Admin from './components/Admin.vue'

const { locale } = useI18n()

const langs = [
  { code: 'zh', label: '中文' },
  { code: 'en', label: 'EN' },
  { code: 'ja', label: '日本語' },
  { code: 'ko', label: '한국어' },
  { code: 'th', label: 'ไทย' },
  { code: 'vi', label: 'Tiếng Việt' },
  { code: 'es', label: 'Español' },
  { code: 'hi', label: 'हिन्दी' },
]

const switchLang = (lang) => {
  locale.value = lang
  localStorage.setItem('loveTrust_locale', lang)
}

const currentTab = ref('search')
const showAgreement = ref(false)
const showPrivacy = ref(false)
const showContact = ref(false)

const onHashChange = () => {
  if (window.location.hash === '#admin') currentTab.value = 'admin'
}
onMounted(() => {
  onHashChange()
  window.addEventListener('hashchange', onHashChange)
})
onBeforeUnmount(() => {
  window.removeEventListener('hashchange', onHashChange)
})
</script>

<style scoped>
.header {
  text-align: center;
  margin-bottom: 2rem;
  padding-top: 1rem;
}
.logo {
  font-size: 2.5rem;
  font-weight: 700;
  letter-spacing: -1px;
}
.subtitle {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.nav-tabs {
  display: flex;
  background: var(--silicon-bg);
  border-radius: var(--radius-lg);
  padding: 8px;
  margin-bottom: 2.5rem;
  box-shadow: var(--shadow-in);
  gap: 10px;
}
.tab-btn {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 12px 0;
  font-size: 1rem;
  font-weight: 600;
  border-radius: calc(var(--radius-lg) - 4px);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.2, 0.8, 0.2, 1);
  text-transform: uppercase;
  letter-spacing: 1px;
}
.tab-btn.active {
  background: var(--silicon-surface);
  color: var(--accent-neon);
  box-shadow: var(--shadow-out), 0 0 10px rgba(0, 240, 255, 0.2);
  text-shadow: 0 0 8px rgba(0, 240, 255, 0.5);
}

/* Lang toggle */
.lang-toggle {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: 10px;
}
.lang-btn {
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--text-secondary);
  padding: 3px 10px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s;
  white-space: nowrap;
}
.lang-btn.active {
  border-color: var(--accent-neon);
  color: var(--accent-neon);
  text-shadow: 0 0 6px rgba(0,240,255,0.5);
}
.lang-btn:hover:not(.active) {
  color: var(--text-primary);
  border-color: rgba(255,255,255,0.4);
}

/* Footer Styles */
.glass-footer {
  text-align: center;
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}
.legal-links {
  margin-bottom: 1rem;
}
.legal-sep {
  color: var(--text-secondary);
  opacity: 0.3;
  margin: 0 8px;
}
.legal-link {
  color: var(--accent-neon);
  font-size: 0.85rem;
  text-decoration: none;
  transition: color 0.2s ease;
}
.legal-link:hover {
  color: var(--accent-neon-light);
  text-shadow: 0 0 8px rgba(0, 240, 255, 0.5);
}
.disclaimer {
  color: var(--text-secondary);
  font-size: 0.85rem;
  line-height: 1.6;
  opacity: 0.6;
}

/* Legal Modal Overlays */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(10, 10, 16, 0.85);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}
.modal-content {
  background: var(--silicon-surface);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 0 50px rgba(0, 240, 255, 0.1);
  width: 100%;
  max-width: 600px;
  max-height: 85vh;
  border-radius: var(--radius-lg);
  padding: 30px;
  position: relative;
  display: flex;
  flex-direction: column;
}
.close-btn {
  position: absolute;
  top: 15px;
  right: 20px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.5rem;
  cursor: pointer;
  transition: color 0.2s;
}
.close-btn:hover {
  color: var(--danger-neon);
  text-shadow: 0 0 10px rgba(255, 0, 85, 0.5);
}
.modal-title {
  color: var(--text-primary);
  font-size: 1.3rem;
  font-weight: 600;
  margin-bottom: 20px;
  text-align: center;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.1);
  padding-bottom: 15px;
}
.modal-body {
  overflow-y: auto;
  padding-right: 10px;
  color: #ccc;
  font-size: 0.95rem;
  line-height: 1.7;
}
.modal-body::-webkit-scrollbar {
  width: 6px;
}
.modal-body::-webkit-scrollbar-thumb {
  background: rgba(0, 240, 255, 0.3);
  border-radius: 4px;
}
.modal-body p {
  margin-bottom: 15px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

/* Contact Modal */
.contact-desc {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 1.5rem;
  text-align: center;
}
.contact-channels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.contact-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 1.25rem;
  text-align: center;
  transition: border-color 0.2s;
}
.contact-card:hover {
  border-color: rgba(0, 240, 255, 0.3);
}
.contact-card:first-child {
  grid-column: 1 / -1;
}
.contact-card-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.contact-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.8rem;
  color: #fff;
  flex-shrink: 0;
}
.wechat-icon { background: #07c160; }
.email-icon { background: #ea4335; }
.tg-icon { background: #0088cc; }
.hours-icon { background: #6c5ce7; }
.contact-label {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
}
.contact-qr {
  width: 180px;
  height: 180px;
  border-radius: 10px;
  margin: 0.5rem auto;
  display: block;
  background: #fff;
  padding: 6px;
}
.contact-hint {
  color: var(--text-secondary);
  font-size: 0.8rem;
  margin-top: 0.5rem;
}
.contact-value {
  font-size: 0.9rem;
  color: var(--text-secondary);
}
.contact-link {
  color: var(--accent-neon);
  text-decoration: none;
  transition: text-shadow 0.2s;
  display: block;
}
.contact-link:hover {
  text-shadow: 0 0 8px rgba(0, 240, 255, 0.5);
}

@media (max-width: 480px) {
  .header { margin-bottom: 1rem; padding-top: 0.5rem; }
  .logo { font-size: 1.6rem; }
  .subtitle { font-size: 0.8rem; }
  .lang-toggle { gap: 4px; margin-bottom: 6px; }
  .lang-btn { padding: 2px 7px; font-size: 0.68rem; }
  .nav-tabs { padding: 4px; gap: 4px; margin-bottom: 1.5rem; border-radius: 14px; }
  .tab-btn { padding: 10px 0; font-size: 0.78rem; letter-spacing: 0.3px; }
  .glass-footer { margin-top: 2rem; padding-top: 1rem; }
  .disclaimer { font-size: 0.75rem; }
  .modal-overlay { padding: 10px; }
  .modal-content { padding: 18px 14px; max-height: 90vh; }
  .modal-title { font-size: 1.1rem; margin-bottom: 14px; }
  .modal-body { font-size: 0.85rem; padding-right: 4px; }
  .contact-channels { grid-template-columns: 1fr; gap: 0.75rem; }
  .contact-card:first-child { grid-column: 1; }
  .contact-qr { width: 150px; height: 150px; }
  .contact-card { padding: 1rem; }
}
</style>

<template>
  <!-- PayPal capture status banner -->
  <div v-if="paypalCapturing" class="paypal-capture-banner">
    <span class="spinner"></span> {{ $t('search.pay_paypal_capturing') }}
  </div>

  <div class="search-container glass-panel">
    <!-- Public Stats Bar -->
    <div v-if="!searched && pubStats" class="pub-stats-bar">
      <div class="pub-stat">
        <span class="pub-stat-val">{{ animatedReports.toLocaleString() }}+</span>
        <span class="pub-stat-label">{{ $t('search.stat_reports') }}</span>
      </div>
      <div class="pub-stat-divider"></div>
      <div class="pub-stat">
        <span class="pub-stat-val">{{ animatedCities }}</span>
        <span class="pub-stat-label">{{ $t('search.stat_cities') }}</span>
      </div>
      <div class="pub-stat-divider"></div>
      <div class="pub-stat">
        <span class="pub-stat-val">{{ animatedAlerts.toLocaleString() }}+</span>
        <span class="pub-stat-label">{{ $t('search.stat_alerts') }}</span>
      </div>
    </div>

    <!-- Initial Search View -->
    <div v-if="!searched" class="search-box">
      <h2 style="margin-bottom: 1rem; font-weight: 600;">{{ $t('search.title') }}</h2>
      <p style="color: var(--text-secondary); margin-bottom: 1.5rem; font-size: 0.9rem;">
        {{ $t('search.subtitle') }}
      </p>
      
      <div class="input-group">
        <PhoneInput
          v-model="phone"
          v-model:countryCode="countryCode"
          @validity="v => phoneValid = v"
          :placeholder="$t('search.placeholder')"
        />
      </div>

      <div class="input-group" style="margin-top: 0.75rem;">
        <textarea
          v-model="remark"
          class="input-premium textarea-premium"
          :placeholder="$t('search.remark_placeholder')"
          rows="3"
          maxlength="1000"
        ></textarea>
      </div>
      
      <button 
        class="btn-premium search-btn" 
        @click="handleSearch" 
        :disabled="loading || !phoneValid">
        <span v-if="!loading">{{ $t('search.btn_search') }}</span>
        <span v-else class="loading-state">
          <span class="spinner"></span> {{ $t('search.btn_loading') }}
        </span>
      </button>
      <p v-if="searchError" class="inline-alert" @click="searchError = ''">{{ searchError }}</p>
    </div>

    <!-- Results Area post-fetch -->
    <div v-else class="results-box">
      <div class="actions-top">
        <button class="back-btn" @click="resetSearch">{{ $t('search.btn_reset') }}</button>
      </div>

      <!-- Share Buttons -->
      <div v-if="resultType === 'warning' && hasPaid" class="share-bar">
        <span class="share-label">{{ $t('search.share_title') }}</span>
        <div class="share-btns">
          <button class="share-btn share-twitter" @click="shareTwitter" title="Twitter / X">X</button>
          <button class="share-btn share-whatsapp" @click="shareWhatsApp" title="WhatsApp">WA</button>
          <button class="share-btn share-copy" @click="shareCopy">{{ copyLabel || $t('search.share_copy') }}</button>
        </div>
      </div>

      <!-- Clean Result -->
      <div v-if="resultType === 'clean'" class="result-card clean-card">
        <div class="icon-safe"></div>
        <h3 style="color: var(--safe-color);">{{ $t('search.clean_title') }}</h3>
        <p class="desc">{{ $t('search.clean_desc') }}</p>
        <p class="note">{{ $t('search.clean_note') }}</p>
        <button v-if="pushSupported && !watchSubscribed" class="btn-watch" @click="subscribeWatch">
          {{ $t('search.watch_btn') }}
        </button>
        <p v-if="watchSubscribed" class="watch-done">{{ $t('search.watch_done') }}</p>
      </div>

      <!-- Warning Aggregated Profile Result -->
      <div v-if="resultType === 'warning'" class="result-card warning-card">
        <!-- Render pulsating warning radar -->
        <div class="radar-warning-icon"></div>
        <h3 style="color: var(--danger-color); font-size: 1.5rem; margin-bottom: 0.5rem">
          {{ $t('search.warning_title') }}
        </h3>
        
        <!-- Paywall Logic: Only show full details if hasPaid === true -->
        <div class="agg-profile" :class="{'locked-panel': !hasPaid}">
          <div v-if="!hasPaid" class="lock-overlay">
            <div class="lock-icon">{{ $t('search.lock_title') }}</div>
            <h4 style="color: var(--accent-neon); margin-bottom: 8px; font-size:1.1rem;">{{ $t('search.lock_subtitle') }}</h4>
            <p style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 20px;">
              {{ $t('search.lock_desc') }}<br/>
              {{ $t('search.lock_pay') }} <strong>{{ $t('search.lock_count', { n: profile.total_independent_reports }) }}</strong>
            </p>
            <button class="btn-premium pay-btn" @click="openCodeModal">
              {{ $t('search.pay_btn') }}
            </button>
          </div>
          
          <div v-else class="unlocked-content">
            <div class="success-badge">{{ $t('search.unlocked_badge') }}</div>

            <!-- Aggregated Summary Header -->
            <div class="agg-row">
              <span class="label">{{ $t('search.field_name') }}</span>
              <span class="value">{{ profile.display_name }} {{ $t('search.field_name_suffix') }}</span>
            </div>
            <div class="agg-row">
              <span class="label">{{ $t('search.field_victims') }}</span>
              <span class="value highlight-danger">{{ profile.total_independent_reports }} {{ $t('search.field_victims_suffix') }}</span>
            </div>
            <div class="agg-row">
              <span class="label">{{ $t('search.field_risk') }}</span>
              <span class="value highlight-danger">
                {{ profile.consensus_risk_score }} {{ $t('search.field_risk_suffix') }}
                <span :class="'risk-badge risk-' + (profile.risk_level || 'low')">{{ $t('search.risk_' + (profile.risk_level || 'low')) }}</span>
              </span>
            </div>
            <div class="risk-bar-wrap" v-if="profile.consensus_risk_score">
              <div class="risk-bar-fill" :style="{ width: (profile.consensus_risk_score * 10) + '%' }" :class="'risk-fill-' + (profile.risk_level || 'low')"></div>
            </div>
            <div class="agg-row">
              <span class="label">{{ $t('search.field_tags') }}</span>
              <span class="value tags">
                <span v-for="tag in profile.consolidated_tags || []" :key="tag" class="tag warning-tag">{{ $t('report.tag_' + tag, tag) }}</span>
              </span>
            </div>
            <div class="agg-row" v-if="profile.first_report_at">
              <span class="label">{{ $t('search.field_first_report') }}</span>
              <span class="value">{{ formatDate(profile.first_report_at) }}</span>
            </div>
            <div class="agg-row" v-if="profile.latest_report_at && profile.total_independent_reports > 1">
              <span class="label">{{ $t('search.field_latest_report') }}</span>
              <span class="value">{{ formatDate(profile.latest_report_at) }}</span>
            </div>

            <!-- Per-Record Timeline -->
            <div v-if="records.length > 0" class="record-timeline">
              <h4 class="record-section-title">{{ $t('search.record_section_title') }}</h4>
              <div v-for="(rec, idx) in records" :key="'rec-'+rec.id" class="record-card">
                <div class="record-header">
                  <span class="record-num">{{ $t('search.record_title', { n: idx + 1 }) }}</span>
                  <span class="record-date">{{ formatDate(rec.created_at) }}</span>
                </div>
                <div class="record-body">
                  <div class="record-row record-reporter-row">
                    <span class="record-label">{{ $t('search.reporter_identity') }}</span>
                    <div class="reporter-meta">
                      <p v-if="rec.reporter_city" class="reporter-line">
                        {{ $t('search.reporter_from_city', { city: rec.reporter_city, name: displayReporterName(rec) }) }}
                      </p>
                      <p v-else class="reporter-line">
                        {{ $t('search.reporter_virtual_only', { name: displayReporterName(rec) }) }}
                      </p>
                      <div class="reporter-badges">
                        <span class="verify-chip" :class="'vl-' + (rec.verification_level || 1)">{{ $t('search.verify_level_' + (rec.verification_level || 1)) }}</span>
                        <span class="cred-stars" :title="$t('search.credibility_short')">{{ starRating(rec.verification_level) }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="record-row" v-if="rec.location_city">
                    <span class="record-label">{{ $t('search.record_city') }}</span>
                    <span class="tag">{{ rec.location_city }}</span>
                  </div>
                  <div class="record-row" v-if="rec.tags && rec.tags.length > 0">
                    <span class="record-label">{{ $t('search.record_tags') }}</span>
                    <span class="record-tags">
                      <span v-for="tag in rec.tags" :key="tag" class="tag warning-tag">{{ $t('report.tag_' + tag, tag) }}</span>
                    </span>
                  </div>
                  <div class="record-row record-desc-row" v-if="rec.description">
                    <span class="record-label">{{ $t('search.record_desc') }}</span>
                    <div class="desc-item">{{ rec.description }}</div>
                  </div>
                  <div v-else class="record-row">
                    <span class="record-no-desc">{{ $t('search.record_no_desc') }}</span>
                  </div>
                  <div class="record-row" v-if="rec.evidences && rec.evidences.length > 0">
                    <span class="record-label">{{ $t('search.record_evidence', { n: rec.evidences.length }) }}</span>
                    <div class="evidence-gallery compact-gallery">
                      <div class="gallery-item danger-item" v-for="(ev, eidx) in rec.evidences" :key="'rev-'+idx+'-'+eidx" @click="openPreview(ev)">
                        {{ $t('search.evidence_item', { n: eidx + 1 }) }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Jury & Appeal Section -->
            <div class="jury-section">
              <h4 class="jury-title">{{ $t('search.jury_title') }}</h4>
              <div class="appeal-box" v-if="profile.has_appeal">
                <span class="badge-yellow">{{ $t('search.appeal_badge') }}</span>
                <span v-if="profile.appeal_at" class="appeal-time">{{ $t('search.field_appeal_time') }} {{ formatDate(profile.appeal_at) }}</span>
                <p class="appeal-text">"{{ profile.appeal_reason }}"</p>
                <div class="appeal-evidence-list" v-if="profile.appeal_evidences && profile.appeal_evidences.length > 0">
                  <div class="gallery-item" v-for="(ev, i) in profile.appeal_evidences" :key="'apl-'+i" @click="openPreview(ev)">
                    {{ $t('search.appeal_item', { n: i + 1 }) }}
                  </div>
                </div>
                <p v-else class="empty-evidence">{{ $t('search.appeal_empty_evidence') }}</p>
              </div>
              <p v-else class="empty-evidence" style="margin: 10px 0;">{{ $t('search.no_appeal') }}</p>

              <div v-if="voteResult && voteResult.cleansed" class="cleansed-banner">
                {{ $t('search.cleansed_banner') }}
              </div>

              <div class="vote-area" v-if="profile.has_appeal">
                <p class="vote-hint">{{ $t('search.vote_hint') }}</p>
                <div class="vote-btns">
                  <button :class="['vote-btn', 'support-reporter', { disabled: hasVoted }]" @click="doVote('reporter')" :disabled="hasVoted || votingInProgress">
                    {{ $t('search.vote_reporter', { n: voteCountReporter }) }}
                  </button>
                  <button :class="['vote-btn', 'support-appeal', { disabled: hasVoted }]" @click="doVote('appeal')" :disabled="hasVoted || votingInProgress">
                    {{ $t('search.vote_appeal', { n: voteCountAppeal }) }}
                  </button>
                </div>
                <p v-if="hasVoted" style="text-align:center;color:var(--safe-neon);margin-top:10px;font-size:0.85rem;">{{ $t('search.vote_done') }}</p>
              </div>
            </div>

        </div> <!-- end of unlocked-content -->
      </div> <!-- end of agg-profile -->
      </div> <!-- end of warning-card -->
    </div> <!-- end of results-box -->
  </div> <!-- end of search-container -->

  <!-- Unlock Modal (WeChat Pay + Activation Code) -->
  <transition name="fade">
    <div v-if="showCodeModal" class="code-modal-overlay" @click.self="closeCodeModal">
      <div class="code-modal">
        <button class="close-modal-btn" @click="closeCodeModal" style="position:absolute;top:12px;right:12px;">×</button>
        <h3 class="code-modal-title">{{ $t('search.code_title') }}</h3>

        <!-- Pay Method Tabs (locale-based) -->
        <div class="pay-method-tabs">
          <button v-for="m in availablePayMethods" :key="m"
            :class="['pay-tab', { active: payMethod === m }]"
            @click="payMethod = m">
            {{ m === 'wechat' ? $t('search.pay_wechat') : m === 'alipay' ? $t('search.pay_alipay') : m === 'paypal' ? $t('search.pay_paypal') : $t('search.pay_code_tab') }}
          </button>
        </div>

        <!-- WeChat / Alipay Pay Panel (Xunhupay) -->
        <div v-if="payMethod === 'wechat' || payMethod === 'alipay'" class="pay-panel">
          <div v-if="!payOrderNo" class="pay-start">
            <p class="pay-price-label">{{ $t('search.pay_amount') }}</p>
            <div class="pay-price-value">{{ $t('search.pay_currency_cny') }}{{ $t('search.pay_price') }}</div>
            <button class="btn-premium wechat-pay-btn" @click="startXunhuPay(payMethod)" :disabled="payCreating">
              <span v-if="!payCreating">{{ payMethod === 'alipay' ? $t('search.pay_alipay_btn') : $t('search.pay_wechat_btn') }}</span>
              <span v-else><span class="spinner"></span> {{ $t('search.pay_creating') }}</span>
            </button>
            <p v-if="payError" class="code-error" style="margin-top:0.5rem">{{ payError }}</p>
          </div>

          <div v-else class="pay-pending">
            <div v-if="payQrUrl && !isMobile" class="pay-qr-section">
              <p class="pay-qr-hint">{{ payMethod === 'alipay' ? $t('search.pay_alipay_qr_hint') : $t('search.pay_qr_hint') }}</p>
              <img :src="payQrUrl" class="pay-qr-img" :alt="payMethod === 'alipay' ? 'Alipay QR' : 'WeChat Pay QR'" />
            </div>
            <div v-else class="pay-redirect-hint">
              <p>{{ $t('search.pay_redirect_hint') }}</p>
            </div>
            <div class="pay-status-bar">
              <span class="spinner"></span>
              <span>{{ payStatusMsg || $t('search.pay_waiting') }}</span>
            </div>
            <p v-if="payError" class="code-error" style="margin-top:0.5rem">{{ payError }}</p>
          </div>
        </div>

        <!-- PayPal Panel -->
        <div v-if="payMethod === 'paypal'" class="pay-panel">
          <div class="pay-start">
            <p class="pay-price-label">{{ $t('search.pay_amount') }}</p>
            <div class="pay-price-value">{{ $t('search.pay_currency_usd') }}{{ $t('search.pay_price_usd') }}</div>
            <button class="btn-premium paypal-pay-btn" @click="startPayPalPay" :disabled="payCreating">
              <span v-if="!payCreating">{{ $t('search.pay_paypal_btn') }}</span>
              <span v-else><span class="spinner"></span> {{ $t('search.pay_creating') }}</span>
            </button>
            <p v-if="payError" class="code-error" style="margin-top:0.5rem">{{ payError }}</p>
            <p v-if="payStatusMsg" class="pay-status-msg" style="margin-top:0.5rem;color:var(--text-secondary)">{{ payStatusMsg }}</p>
          </div>
        </div>

        <!-- Activation Code Panel (existing flow) -->
        <div v-if="payMethod === 'code'" class="pay-panel">
          <div class="code-step">
            <span class="code-step-num">1</span>
            <div style="width:100%">
              <p class="code-step-label">{{ $t('search.code_step1') }}</p>
              <p class="code-step-desc">{{ $t('search.code_step1_desc') }}</p>
              <div class="buy-links">
                <a v-for="p in buyPlatforms" :key="p.id" :href="p.url" target="_blank" rel="noopener" class="btn-premium buy-link">
                  {{ p.icon ? p.icon + ' ' : '' }}{{ p.name }}
                </a>
                <span v-if="!buyPlatforms.length" class="empty-hint" style="font-size:0.82rem;color:var(--text-secondary)">{{ $t('search.code_no_platforms') }}</span>
              </div>
            </div>
          </div>

          <div class="code-step">
            <span class="code-step-num">2</span>
            <div style="width:100%">
              <p class="code-step-label">{{ $t('search.code_step2') }}</p>
              <input
                v-model="activationCode"
                type="text"
                class="input-premium code-input"
                placeholder="LT-XXXX-XXXX-XXXX"
                maxlength="17"
                @keyup.enter="submitActivationCode"
              />
              <p v-if="codeError" class="code-error">{{ codeError }}</p>
            </div>
          </div>

          <button class="btn-premium unlock-btn" @click="submitActivationCode" :disabled="activating || !activationCode.trim()">
            <span v-if="!activating">{{ $t('search.code_unlock_btn') }}</span>
            <span v-else><span class="spinner"></span> {{ $t('search.code_verifying') }}</span>
          </button>
        </div>
      </div>
    </div>
  </transition>

  <!-- Fullscreen Evidence Image Preview Modal -->
  <transition name="fade">
    <div v-if="previewImage" class="image-modal" @click="closePreview">
      <button class="close-modal-btn" @click="closePreview">×</button>
      <img :src="previewImage" class="preview-img" @click.stop :alt="$t('search.evidence_alt')" />
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import PhoneInput from './PhoneInput.vue'
import { getDefaultCountry, getDialCode } from '../data/countryCodes.js'
import { trackEvent } from '../utils/analytics.js'

const { t, locale } = useI18n()

const phone = ref('')
const remark = ref('')
const countryCode = ref(getDefaultCountry(locale.value))
const phoneValid = ref(false)
const loading = ref(false)
const searched = ref(false)
const resultType = ref('') // expected variables logic: 'clean' or 'warning'
const profile = ref(null)
const records = ref([])
const hasPaid = ref(false)
const showCodeModal = ref(false)
const activationCode = ref('')
const activating = ref(false)
const codeError = ref('')
const queryToken = ref('')
const buyPlatforms = ref([])
const hasVoted = ref(false)
const votingInProgress = ref(false)
const voteResult = ref(null)
const searchError = ref('')
const previewImage = ref(null)
const payOptions = ref({
  code_enabled: true,
  providers: {
    wechat: false,
    alipay: false,
    paypal: false,
  },
})

const copyLabel = ref('')
const pushSupported = ref('serviceWorker' in navigator && 'PushManager' in window)
const watchSubscribed = ref(false)

const payCreating = ref(false)
const payOrderNo = ref('')
const payQrUrl = ref('')
const payError = ref('')
const payStatusMsg = ref('')
const payPollTimer = ref(null)
const isMobile = ref(/Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent))
const paypalCapturing = ref(false)

const availablePayMethods = computed(() => {
  const methods = []
  const providers = payOptions.value?.providers || {}

  if (locale.value === 'zh') {
    if (providers.wechat) methods.push('wechat')
    if (providers.alipay) methods.push('alipay')
  } else if (providers.paypal) {
    methods.push('paypal')
  }

  if (payOptions.value?.code_enabled !== false || methods.length === 0) {
    methods.push('code')
  }

  return methods
})

const voteCountReporter = computed(() => {
  const n = profile.value?.reporter_votes
  return typeof n === 'number' && !Number.isNaN(n) ? n : 0
})
const voteCountAppeal = computed(() => {
  const n = profile.value?.appeal_votes
  return typeof n === 'number' && !Number.isNaN(n) ? n : 0
})

const payMethod = ref('')
const initPayMethod = () => {
  payMethod.value = availablePayMethods.value[0]
}
const pubStats = ref(null)
const animatedReports = ref(0)
const animatedCities = ref(0)
const animatedAlerts = ref(0)

function animateNumber(target, setter, duration = 1500) {
  const start = 0
  const startTime = performance.now()
  function step(now) {
    const elapsed = now - startTime
    const progress = Math.min(elapsed / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    setter(Math.floor(start + (target - start) * eased))
    if (progress < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return '-'
  return d.toLocaleDateString(locale.value, { year: 'numeric', month: 'long', day: 'numeric' })
}

function displayReporterName(rec) {
  if (rec && rec.reporter_display_name) return rec.reporter_display_name
  return t('search.reporter_fallback')
}

function starRating(level) {
  const lv = Number(level) || 1
  const n = lv >= 3 ? 5 : lv >= 2 ? 4 : 3
  return '\u2605'.repeat(n) + '\u2606'.repeat(5 - n)
}

const accessStorageKey = (targetHash) => `lt_access_${targetHash}`
const getStoredAccessToken = (targetHash) => targetHash ? localStorage.getItem(accessStorageKey(targetHash)) || '' : ''
const storeAccessToken = (targetHash, token) => {
  if (targetHash && token) {
    localStorage.setItem(accessStorageKey(targetHash), token)
  }
}
const clearAccessToken = (targetHash) => {
  if (targetHash) {
    localStorage.removeItem(accessStorageKey(targetHash))
  }
}

const shareText = () => {
  const name = profile.value?.display_name || ''
  const score = profile.value?.consensus_risk_score || 0
  return `LoverTrust Risk Alert: ${name} - Risk Score ${score}/10. Check now: ${window.location.origin}`
}
const shareTwitter = () => {
  window.open(`https://x.com/intent/tweet?text=${encodeURIComponent(shareText())}`, '_blank')
}
const shareWhatsApp = () => {
  window.open(`https://wa.me/?text=${encodeURIComponent(shareText())}`, '_blank')
}
const shareCopy = async () => {
  try {
    await navigator.clipboard.writeText(shareText())
    copyLabel.value = t('search.share_copied')
    setTimeout(() => { copyLabel.value = '' }, 2000)
  } catch { /* fallback ignored */ }
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = atob(base64)
  return Uint8Array.from([...rawData].map(c => c.charCodeAt(0)))
}

const subscribeWatch = async () => {
  try {
    const vapidRes = await fetch('/api/v1/push/vapid-key')
    if (!vapidRes.ok) return
    const { public_key } = await vapidRes.json()
    if (!public_key) return

    const reg = await navigator.serviceWorker.ready
    const sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(public_key),
    })
    const subJson = sub.toJSON()
    const fullPhone = getDialCode(countryCode.value) + phone.value
    await fetch('/api/v1/push/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        phone: fullPhone,
        phone_local: phone.value,
        endpoint: subJson.endpoint,
        key_auth: subJson.keys.auth,
        key_p256dh: subJson.keys.p256dh,
      }),
    })
    watchSubscribed.value = true
    trackEvent('push_subscribe')
  } catch { /* permission denied or not supported */ }
}

onMounted(async () => {
  try {
    const payRes = await fetch('/api/v1/pay/options')
    if (payRes.ok) {
      payOptions.value = await payRes.json()
    }
  } catch {}
  initPayMethod()

  const urlParams = new URLSearchParams(window.location.search)
  const ppCapture = urlParams.get('paypal_capture')
  const ppOrderNo = urlParams.get('order_no')
  if (ppCapture === '1' && ppOrderNo) {
    const cleanURL = window.location.pathname + window.location.hash
    window.history.replaceState({}, '', cleanURL)
    capturePayPalOrder(ppOrderNo)
  }

  try {
    const res = await fetch('/api/v1/stats/public')
    if (res.ok) {
      const data = await res.json()
      pubStats.value = data
      animateNumber(data.reports, v => { animatedReports.value = v })
      animateNumber(data.cities, v => { animatedCities.value = v })
      animateNumber(data.alerts, v => { animatedAlerts.value = v })
    }
  } catch {}
})

watch(availablePayMethods, (methods) => {
  if (!methods.includes(payMethod.value)) {
    payMethod.value = methods[0] || 'code'
  }
}, { immediate: true })

const mockAppealImage = computed(() =>
  "data:image/svg+xml;charset=utf-8," + encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="800"><rect width="600" height="800" fill="#131419"/><text x="50%" y="45%" fill="#facc15" font-size="26" text-anchor="middle" font-family="sans-serif">${t('search.mock_image_title')}</text><text x="50%" y="52%" fill="#888" font-size="18" text-anchor="middle" font-family="sans-serif">${t('search.mock_image_desc')}</text></svg>`)
)

const openPreview = (url) => {
  if (!url) {
    previewImage.value = mockAppealImage.value
    return
  }
  // Let Vite Proxy redirect it to the backend `GET /api/v1/evidence/:filename` route!
  previewImage.value = `/api/v1/evidence/${url}`
}
const closePreview = () => {
  previewImage.value = null
}

const requestSearch = async (fullPhone, accessToken = '') => {
  const queryParams = new URLSearchParams({ phone: fullPhone, phone_local: phone.value })
  if (remark.value.trim()) queryParams.set('remark', remark.value.trim())

  const headers = {}
  if (accessToken) {
    headers['X-Access-Token'] = accessToken
  }

  const res = await fetch(`/api/v1/query?${queryParams}`, { headers })
  const data = await res.json()
  return { res, data }
}

const applyWarningResult = (data) => {
  resultType.value = 'warning'
  profile.value = data.aggregated_profile || {}
  records.value = data.records || []
  queryToken.value = data.query_token || ''
  hasPaid.value = !!data.unlocked
  searched.value = true
}

const refreshUnlockedWarning = async (accessToken) => {
  if (!accessToken || !phoneValid.value) return false

  const fullPhone = getDialCode(countryCode.value) + phone.value
  try {
    const { res, data } = await requestSearch(fullPhone, accessToken)
    if (res.ok && data.status === 'warning' && data.unlocked) {
      applyWarningResult(data)
      return true
    }
  } catch {}
  return false
}

const unlockCurrentResult = async (accessToken) => {
  if (!accessToken || !queryToken.value) return false
  storeAccessToken(queryToken.value, accessToken)
  const ok = await refreshUnlockedWarning(accessToken)
  if (!ok) {
    clearAccessToken(queryToken.value)
  }
  return ok
}

const startXunhuPay = async (provider) => {
  payCreating.value = true
  payError.value = ''
  try {
    const res = await fetch('/api/v1/pay/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_hash: queryToken.value, provider }),
    })
    const data = await res.json()
    if (!res.ok || data.error) {
      payError.value = t('search.pay_failed')
      return
    }
    payOrderNo.value = data.order_no
    payQrUrl.value = data.qr_url || ''

    if (isMobile.value && data.pay_url) {
      window.location.href = data.pay_url
    }

    startPayPolling()
    trackEvent(provider + '_pay_initiated')
  } catch {
    payError.value = t('search.pay_failed')
  } finally {
    payCreating.value = false
  }
}

const startPayPalPay = async () => {
  payCreating.value = true
  payError.value = ''
  try {
    const res = await fetch('/api/v1/pay/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_hash: queryToken.value, provider: 'paypal' }),
    })
    const data = await res.json()
    if (!res.ok || data.error) {
      payError.value = t('search.pay_failed')
      return
    }
    payOrderNo.value = data.order_no
    if (data.approve_url) {
      payStatusMsg.value = t('search.pay_paypal_redirect')
      window.location.href = data.approve_url
    } else {
      payError.value = t('search.pay_failed')
    }
    trackEvent('paypal_pay_initiated')
  } catch {
    payError.value = t('search.pay_failed')
  } finally {
    payCreating.value = false
  }
}

const capturePayPalOrder = async (orderNo) => {
  paypalCapturing.value = true
  payError.value = ''
  payStatusMsg.value = t('search.pay_paypal_capturing')
  try {
    const res = await fetch('/api/v1/pay/paypal-capture', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ order_no: orderNo }),
    })
    const data = await res.json()
    if (data.paid && data.access_token) {
      const unlocked = await unlockCurrentResult(data.access_token)
      if (!unlocked) {
        payError.value = t('search.pay_failed')
        return
      }
      payStatusMsg.value = t('search.pay_success')
      trackEvent('paypal_pay_success')
    } else {
      payError.value = t('search.pay_failed')
    }
  } catch {
    payError.value = t('search.pay_failed')
  } finally {
    paypalCapturing.value = false
  }
}

const startPayPolling = () => {
  stopPayPolling()
  let attempts = 0
  payPollTimer.value = setInterval(async () => {
    attempts++
    if (attempts > 200) {
      stopPayPolling()
      payStatusMsg.value = t('search.pay_expired')
      return
    }
    try {
      const res = await fetch(`/api/v1/pay/status?order_no=${payOrderNo.value}`)
      if (!res.ok) return
      const data = await res.json()
      if (data.paid && data.access_token) {
        stopPayPolling()
        const unlocked = await unlockCurrentResult(data.access_token)
        if (!unlocked) {
          payError.value = t('search.pay_failed')
          return
        }
        showCodeModal.value = false
        payStatusMsg.value = ''
        trackEvent('wechat_pay_success')
      }
    } catch { /* continue polling */ }
  }, 3000)
}

const stopPayPolling = () => {
  if (payPollTimer.value) {
    clearInterval(payPollTimer.value)
    payPollTimer.value = null
  }
}

onBeforeUnmount(() => { stopPayPolling() })

const resetPayState = () => {
  payOrderNo.value = ''
  payQrUrl.value = ''
  payError.value = ''
  payStatusMsg.value = ''
  payCreating.value = false
  stopPayPolling()
}

const openCodeModal = async () => {
  activationCode.value = ''
  codeError.value = ''
  initPayMethod()
  resetPayState()
  showCodeModal.value = true
  try {
    const res = await fetch('/api/v1/platforms')
    if (res.ok) buyPlatforms.value = await res.json()
  } catch { /* ignore */ }
}

const closeCodeModal = () => {
  showCodeModal.value = false
  stopPayPolling()
}

const submitActivationCode = async () => {
  if (!activationCode.value.trim()) return
  activating.value = true
  codeError.value = ''
  try {
    const res = await fetch('/api/v1/activate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: activationCode.value.trim(),
        target_hash: queryToken.value,
      }),
    })
    const data = await res.json()
    if (data.success && data.access_token) {
      const unlocked = await unlockCurrentResult(data.access_token)
      if (!unlocked) {
        codeError.value = t('search.code_error_unknown')
        return
      }
      showCodeModal.value = false
      trackEvent('code_activated')
    } else {
      codeError.value = t('search.code_error_' + (data.error || 'unknown'))
    }
  } catch {
    codeError.value = t('search.code_error_network')
  } finally {
    activating.value = false
  }
}

const doVote = async (side) => {
  if (hasVoted.value || votingInProgress.value) return
  votingInProgress.value = true
  const fullPhone = getDialCode(countryCode.value) + phone.value
  try {
    const res = await fetch('/api/v1/vote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        target_phone: fullPhone,
        target_phone_local: phone.value,
        side: side,
      }),
    })
    const data = await res.json()
    if (res.ok && data.message === 'vote_recorded') {
      voteResult.value = data
      hasVoted.value = true
      if (side === 'reporter') {
        profile.value.reporter_votes = voteCountReporter.value + 1
      } else {
        profile.value.appeal_votes = voteCountAppeal.value + 1
      }
    }
  } catch {
    /* network error, silently ignore */
  } finally {
    votingInProgress.value = false
  }
}

// handleSearch makes exact hit query via our rewritten merged query hook.
// 触发 API 异步调用以向我们在 Phase 2 构建的主脑发号命令。
const handleSearch = async () => {
  if (!phoneValid.value) return
  loading.value = true
  searchError.value = ''
  const fullPhone = getDialCode(countryCode.value) + phone.value
  
  try {
    const { res, data } = await requestSearch(fullPhone)
    
    // Evaluate status via response standard format handling
    if (res.status === 429) {
      searchError.value = t('search.alert_rate_limit')
    } else if (data.status === 'clean') {
      resultType.value = 'clean'
      profile.value = null
      records.value = []
      queryToken.value = ''
      hasPaid.value = false
      searched.value = true
      trackEvent('search_clean')
    } else if (data.status === 'warning') {
      applyWarningResult(data)
      trackEvent('search_hit', { reports: profile.value.total_independent_reports })

      const storedAccessToken = getStoredAccessToken(queryToken.value)
      if (!data.unlocked && storedAccessToken) {
        const unlocked = await refreshUnlockedWarning(storedAccessToken)
        if (!unlocked) {
          clearAccessToken(queryToken.value)
        }
      }
    }
  } catch(e) {
    searchError.value = t('search.alert_network_error', { e: e.message || e })
  } finally {
    loading.value = false
  }
}

// Resets view
const resetSearch = () => {
  searched.value = false
  resultType.value = ''
  profile.value = null
  records.value = []
  phone.value = ''
  remark.value = ''
  countryCode.value = getDefaultCountry(locale.value)
  hasPaid.value = false
  showCodeModal.value = false
  activationCode.value = ''
  queryToken.value = ''
  hasVoted.value = false
  voteResult.value = null
}
</script>

<style scoped>
.textarea-premium {
  width: 100%;
  resize: vertical;
  min-height: 60px;
  font-family: inherit;
  font-size: 0.95rem;
  line-height: 1.5;
}
.desc-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
  margin-top: 0.25rem;
}
.desc-item {
  background: rgba(255,255,255,0.04);
  border-left: 3px solid var(--danger-neon);
  padding: 0.6rem 0.8rem;
  border-radius: 6px;
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.record-timeline {
  margin-top: 1.5rem;
  border-top: 1px solid rgba(255,255,255,0.08);
  padding-top: 1rem;
}
.record-section-title {
  font-size: 1rem;
  color: var(--danger-neon);
  margin-bottom: 1rem;
  font-weight: 600;
}
.record-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.08);
  border-left: 3px solid var(--accent-neon);
  border-radius: 8px;
  margin-bottom: 1rem;
  overflow: hidden;
}
.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 1rem;
  background: rgba(255,255,255,0.04);
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.record-num {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--accent-neon);
}
.record-date {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.record-body {
  padding: 0.8rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.record-row {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.88rem;
}
.record-desc-row {
  flex-direction: column;
}
.record-label {
  color: var(--text-secondary);
  white-space: nowrap;
  min-width: fit-content;
}
.record-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.record-no-desc {
  color: var(--text-secondary);
  font-size: 0.82rem;
  font-style: italic;
}
.record-reporter-row {
  flex-direction: column;
  align-items: stretch;
  gap: 0.35rem;
}
.reporter-meta {
  width: 100%;
}
.reporter-line {
  margin: 0 0 0.4rem;
  font-size: 0.88rem;
  color: var(--text-primary);
  line-height: 1.45;
}
.reporter-badges {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
}
.verify-chip {
  font-size: 0.75rem;
  padding: 3px 8px;
  border-radius: 4px;
  border: 1px solid rgba(15, 255, 102, 0.35);
  color: var(--safe-neon);
  background: rgba(15, 255, 102, 0.08);
}
.verify-chip.vl-2 {
  border-color: rgba(0, 240, 255, 0.4);
  color: var(--accent-neon);
  background: rgba(0, 240, 255, 0.08);
}
.verify-chip.vl-3 {
  border-color: rgba(250, 204, 21, 0.45);
  color: #facc15;
  background: rgba(250, 204, 21, 0.1);
}
.cred-stars {
  font-size: 0.85rem;
  letter-spacing: 1px;
  color: #facc15;
}
.compact-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
}
.compact-gallery .gallery-item {
  font-size: 0.8rem;
  padding: 4px 10px;
}

@media (max-width: 600px) {
  .record-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }
  .record-row {
    flex-direction: column;
    gap: 2px;
  }
}

.search-btn {
  width: 100%;
  margin-top: 1.5rem;
}
.back-btn {
  background: transparent;
  border: 1px solid rgba(255,255,255,0.2);
  color: var(--text-secondary);
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 1.5rem;
}
.back-btn:hover {
  background: rgba(255,255,255,0.1);
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-radius: 50%;
  border-top-color: #fff;
  animation: spin 1s ease-in-out infinite;
  vertical-align: middle;
  margin-right: 8px;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Result Cards Pop up Interactions */
.result-card {
  text-align: center;
  animation: slideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes slideUp {
  0% { transform: translateY(20px); opacity: 0; }
  100% { transform: translateY(0); opacity: 1; }
}

.icon-safe {
  font-size: 3rem;
  color: var(--safe-color);
  margin-bottom: 1rem;
}
.radar-warning-icon {
  font-size: 3.5rem;
  margin-bottom: 1rem;
  display: inline-block;
  animation: pulse-ring 1.5s infinite alternate ease-in-out;
}
@keyframes pulse-ring {
  0% { transform: scale(0.9); text-shadow: 0 0 10px rgba(239, 68, 68, 0.5); }
  100% { transform: scale(1.1); text-shadow: 0 0 30px rgba(239, 68, 68, 1); }
}

.desc {
  color: var(--text-secondary);
  font-size: 0.95rem;
  margin-top: 1rem;
}
.note {
  color: var(--text-secondary);
  font-size: 0.8rem;
  opacity: 0.7;
  margin-top: 2rem;
}

/* Aggregated Profile Base View Structure */
.agg-profile {
  background: var(--silicon-bg);
  box-shadow: var(--shadow-in);
  border-radius: var(--radius-lg);
  padding: 24px;
  text-align: left;
  margin-top: 2rem;
  border: 1px solid rgba(255, 0, 85, 0.1);
  position: relative;
  overflow: hidden;
}

/* Paywall Lock Mode Styles */
.locked-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 40px 20px;
}
.lock-overlay {
  z-index: 10;
}
.lock-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
  color: var(--text-secondary);
}
.success-badge {
  background: rgba(0, 255, 102, 0.1);
  color: var(--safe-neon);
  border: 1px solid rgba(0, 255, 102, 0.3);
  padding: 6px 12px;
  display: inline-block;
  border-radius: 20px;
  font-size: 0.8rem;
  margin-bottom: 1.5rem;
}

.agg-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1.2rem;
  align-items: center;
}
.agg-row:last-child { margin-bottom: 0; }
.label {
  color: var(--text-secondary);
  font-size: 0.95rem;
}
.value {
  font-weight: 600;
  text-align: right;
  max-width: 60%;
  color: var(--text-primary);
}
.highlight-danger {
  color: var(--danger-neon);
  font-size: 1.2rem;
  text-shadow: 0 0 8px rgba(255, 0, 85, 0.4);
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}
.tag {
  background: var(--silicon-bg);
  box-shadow: var(--shadow-out);
  padding: 6px 14px;
  border-radius: var(--radius-xl);
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-primary);
}
.warning-tag {
  box-shadow: var(--shadow-in);
  color: var(--danger-neon);
  border: 1px solid rgba(255, 0, 85, 0.2);
}

/* Jury and Appeal Box Styles */
.jury-section {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px dashed rgba(255, 255, 255, 0.1);
}
.jury-title {
  color: var(--text-primary);
  font-size: 1.1rem;
  margin-bottom: 1rem;
  text-align: center;
}
.appeal-box {
  background: rgba(250, 204, 21, 0.05);
  border: 1px solid rgba(250, 204, 21, 0.3);
  border-radius: var(--radius-md);
  padding: 1rem;
  margin-bottom: 1.5rem;
}
.badge-yellow {
  display: inline-block;
  background: rgba(250, 204, 21, 0.2);
  color: #facc15;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
  margin-bottom: 0.8rem;
}
.appeal-time {
  display: block;
  color: var(--text-secondary);
  font-size: 0.8rem;
  margin-bottom: 0.6rem;
}
.appeal-text {
  color: var(--text-secondary);
  font-size: 0.95rem;
  font-style: italic;
  margin-bottom: 1rem;
}
.appeal-evidence-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.evidence-gallery {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 10px;
}
.gallery-item {
  background: var(--silicon-bg);
  border: 1px solid rgba(250, 204, 21, 0.3);
  color: #facc15;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}
.gallery-item:hover {
  background: rgba(250, 204, 21, 0.1);
  box-shadow: 0 0 10px rgba(250, 204, 21, 0.2);
}
.gallery-item.danger-item {
  border-color: rgba(255, 0, 85, 0.3);
  color: var(--danger-neon);
}
.gallery-item.danger-item:hover {
  background: rgba(255, 0, 85, 0.1);
  box-shadow: 0 0 10px rgba(255, 0, 85, 0.2);
}
.empty-evidence {
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.85rem;
  padding: 10px 0;
  font-style: italic;
}

.cleansed-banner {
  background: rgba(0, 255, 102, 0.1);
  border: 1px solid rgba(0, 255, 102, 0.4);
  color: var(--safe-neon);
  padding: 12px 16px;
  border-radius: var(--radius-md);
  text-align: center;
  font-weight: 600;
  margin: 12px 0;
  font-size: 0.95rem;
  animation: pulse-ring 2s infinite alternate ease-in-out;
}

.vote-area {
  text-align: center;
}
.vote-hint {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 1rem;
}
.vote-btns {
  display: flex;
  gap: 10px;
}
.vote-btn {
  flex: 1;
  padding: 12px 10px;
  border-radius: var(--radius-md);
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.vote-btn.disabled {
  opacity: 0.6;
  cursor: default;
}
.support-reporter {
  background: rgba(255, 0, 85, 0.1);
  color: var(--danger-neon);
  border: 1px solid rgba(255, 0, 85, 0.3);
}
.support-reporter:not(.disabled):hover {
  background: var(--danger-neon);
  color: #fff;
  box-shadow: 0 0 15px rgba(255, 0, 85, 0.4);
}
.support-appeal {
  background: rgba(250, 204, 21, 0.1);
  color: #facc15;
  border: 1px solid rgba(250, 204, 21, 0.3);
}
.support-appeal:not(.disabled):hover {
  background: #facc15;
  color: #000;
  box-shadow: 0 0 15px rgba(250, 204, 21, 0.4);
}

/* Image Preview Modal CSS */
.image-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(19, 20, 25, 0.95);
  backdrop-filter: blur(15px);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}
.close-modal-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(255,255,255,0.1);
  color: #fff;
  border: none;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  z-index: 10000;
  transition: all 0.3s;
}
.close-modal-btn:hover {
  background: var(--danger-neon);
  transform: rotate(90deg);
  box-shadow: 0 0 20px rgba(255, 0, 85, 0.6);
}
.preview-img {
  max-width: 100%;
  max-height: 90vh;
  border-radius: 12px;
  box-shadow: 0 0 40px rgba(0, 240, 255, 0.3);
  animation: zoomIn 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
  object-fit: contain;
  border: 1px solid rgba(0, 240, 255, 0.4);
}
@keyframes zoomIn {
  from { transform: scale(0.9); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

/* Activation Code Modal */
.code-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}
.code-modal {
  position: relative;
  background: var(--card-bg, #1a1b23);
  border: 1px solid rgba(0, 240, 255, 0.2);
  border-radius: 16px;
  padding: 2rem;
  max-width: 440px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6), 0 0 30px rgba(0, 240, 255, 0.1);
  animation: zoomIn 0.3s ease;
}
.code-modal-title {
  color: var(--accent-neon, #00f0ff);
  font-size: 1.25rem;
  margin-bottom: 1.5rem;
  text-align: center;
}
.code-step {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  align-items: flex-start;
}
.code-step-num {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(0, 240, 255, 0.15);
  color: var(--accent-neon, #00f0ff);
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  margin-top: 2px;
}
.code-step-label {
  color: var(--text-primary, #eee);
  font-weight: 600;
  font-size: 0.95rem;
  margin-bottom: 0.25rem;
}
.code-step-desc {
  color: var(--text-secondary, #999);
  font-size: 0.82rem;
  line-height: 1.4;
  margin-bottom: 0.5rem;
}
.buy-links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 0.25rem;
}
.buy-link {
  display: inline-block;
  font-size: 0.85rem;
  padding: 6px 16px;
  text-decoration: none;
  text-align: center;
}
.code-input {
  width: 100%;
  text-align: center;
  font-size: 1.1rem;
  font-family: 'Courier New', monospace;
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-top: 0.35rem;
}
.code-error {
  color: var(--danger-neon, #ff0055);
  font-size: 0.82rem;
  margin-top: 0.4rem;
}
.unlock-btn {
  width: 100%;
  margin-top: 0.5rem;
  font-size: 1rem;
}

.pub-stats-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1.2rem;
  padding: 1rem 1.5rem;
  margin-bottom: 1.5rem;
  background: linear-gradient(135deg, rgba(0, 240, 255, 0.05), rgba(120, 80, 255, 0.05));
  border: 1px solid rgba(0, 240, 255, 0.1);
  border-radius: 12px;
}
.pub-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.pub-stat-val {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--accent-neon);
  font-variant-numeric: tabular-nums;
}
.pub-stat-label {
  font-size: 0.7rem;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
}
.pub-stat-divider {
  width: 1px;
  height: 28px;
  background: rgba(0, 240, 255, 0.15);
}

.btn-watch {
  margin-top: 1.5rem;
  background: rgba(0,240,255,0.08);
  border: 1px solid rgba(0,240,255,0.2);
  color: var(--accent-neon);
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-watch:hover {
  background: rgba(0,240,255,0.15);
  box-shadow: 0 0 12px rgba(0,240,255,0.2);
}
.watch-done {
  margin-top: 1rem;
  color: var(--safe-neon);
  font-size: 0.85rem;
  text-align: center;
}

.share-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.6rem 1rem;
  margin-bottom: 1rem;
  background: rgba(0,240,255,0.04);
  border: 1px solid rgba(0,240,255,0.1);
  border-radius: 10px;
}
.share-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  white-space: nowrap;
}
.share-btns {
  display: flex;
  gap: 6px;
}
.share-btn {
  padding: 5px 12px;
  border: 1px solid rgba(255,255,255,0.15);
  background: transparent;
  color: var(--text-primary);
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.share-btn:hover {
  background: rgba(255,255,255,0.08);
  border-color: var(--accent-neon);
  color: var(--accent-neon);
}
.share-twitter:hover { border-color: #1da1f2; color: #1da1f2; }
.share-whatsapp:hover { border-color: #25d366; color: #25d366; }

.risk-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-left: 8px;
  vertical-align: middle;
}
.risk-low { background: rgba(0,200,100,0.15); color: #00c864; }
.risk-medium { background: rgba(250,204,21,0.15); color: #facc15; }
.risk-high { background: rgba(255,140,0,0.15); color: #ff8c00; }
.risk-critical { background: rgba(255,0,85,0.2); color: var(--danger-neon); animation: pulse-ring 1.5s infinite alternate; }

.risk-bar-wrap {
  height: 6px;
  background: rgba(255,255,255,0.06);
  border-radius: 3px;
  overflow: hidden;
  margin: 4px 0 1rem;
}
.risk-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 1s cubic-bezier(0.2,0.8,0.2,1);
}
.risk-fill-low { background: linear-gradient(90deg, #00c864, #00e676); }
.risk-fill-medium { background: linear-gradient(90deg, #facc15, #fbbf24); }
.risk-fill-high { background: linear-gradient(90deg, #ff8c00, #ff6b00); }
.risk-fill-critical { background: linear-gradient(90deg, #ff0055, #ff4488); box-shadow: 0 0 8px rgba(255,0,85,0.4); }

.inline-alert {
  color: var(--danger-neon);
  font-size: 0.85rem;
  text-align: center;
  margin-top: 1rem;
  padding: 10px;
  background: rgba(255, 0, 85, 0.08);
  border: 1px solid rgba(255, 0, 85, 0.2);
  border-radius: 8px;
  cursor: pointer;
}

/* Payment Method Tabs */
.pay-method-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 1.25rem;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid rgba(255,255,255,0.1);
}
.pay-tab {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 10px 0;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.pay-tab.active {
  background: rgba(0,240,255,0.08);
  color: var(--accent-neon);
}
.pay-tab:not(.active):hover {
  background: rgba(255,255,255,0.03);
}
.pay-panel { animation: fadeIn 0.3s ease; }
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.pay-start { text-align: center; padding: 1rem 0; }
.pay-price-label {
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-bottom: 0.3rem;
}
.pay-price-value {
  font-size: 2.2rem;
  font-weight: 700;
  color: var(--accent-neon);
  margin-bottom: 1.25rem;
  font-variant-numeric: tabular-nums;
}
.wechat-pay-btn {
  width: 100%;
  background: #07c160;
  color: #fff;
  border-color: #07c160;
  font-size: 1rem;
  letter-spacing: 0.5px;
}
.wechat-pay-btn:hover:not(:disabled) {
  background: #06ae56;
  border-color: #06ae56;
  box-shadow: 0 0 16px rgba(7,193,96,0.4);
  text-shadow: none;
  color: #fff;
}
.paypal-pay-btn {
  width: 100%;
  background: #0070ba;
  color: #fff;
  font-size: 1rem;
  letter-spacing: 0.5px;
}
.paypal-pay-btn:hover:not(:disabled) {
  background: #005ea6;
  border-color: #005ea6;
  box-shadow: 0 0 16px rgba(0,112,186,0.4);
  text-shadow: none;
  color: #fff;
}
.paypal-capture-banner {
  background: rgba(0,112,186,0.15);
  border: 1px solid rgba(0,112,186,0.3);
  border-radius: 12px;
  padding: 1rem 1.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.pay-pending { text-align: center; padding: 0.5rem 0; }
.pay-qr-section { margin-bottom: 1rem; }
.pay-qr-hint {
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}
.pay-qr-img {
  width: 220px;
  height: 220px;
  border-radius: 8px;
  border: 2px solid rgba(7,193,96,0.3);
  background: #fff;
  padding: 8px;
}
.pay-redirect-hint {
  color: var(--text-secondary);
  font-size: 0.9rem;
  padding: 1rem 0;
}
.pay-status-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--accent-neon);
  font-size: 0.85rem;
  padding: 0.75rem 0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 480px) {
  .pub-stats-bar { gap: 0.6rem; padding: 0.7rem 0.8rem; }
  .pub-stat-val { font-size: 1.1rem; }
  .pub-stat-label { font-size: 0.6rem; }
  .pub-stat-divider { height: 22px; }

  .agg-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  .value {
    max-width: 100%;
    text-align: left;
  }
  .tags { justify-content: flex-start; }
  .tag { padding: 4px 10px; font-size: 0.78rem; }

  .agg-profile { padding: 16px 12px; }
  .locked-panel { padding: 24px 12px; }
  .lock-icon { font-size: 2rem; }

  .code-modal { padding: 1.25rem; }
  .code-modal-title { font-size: 1.05rem; margin-bottom: 1rem; }
  .code-step-num { width: 24px; height: 24px; font-size: 0.78rem; }
  .code-input { font-size: 0.95rem; }

  .share-bar { flex-direction: column; gap: 0.4rem; align-items: stretch; }
  .share-btns { justify-content: center; }

  .vote-btns { flex-direction: column; gap: 8px; }
  .vote-btn { padding: 10px; font-size: 0.85rem; }

  .highlight-danger { font-size: 1rem; }
  .risk-badge { font-size: 0.65rem; padding: 2px 6px; }

  .close-modal-btn { width: 36px; height: 36px; font-size: 1.2rem; top: 12px; right: 12px; }

  .pay-tab { font-size: 0.8rem; padding: 8px 0; }
  .pay-price-value { font-size: 1.8rem; }
  .pay-qr-img { width: 180px; height: 180px; }
}
</style>

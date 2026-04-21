<template>
  <div class="admin-container glass-panel">
    <!-- Login Gate -->
    <div v-if="!authed" class="login-gate">
      <h2>{{ $t('admin.login_title') }}</h2>
      <form @submit.prevent="doLogin" autocomplete="on">
        <input v-model="secret" type="password" class="input-premium" :placeholder="$t('admin.login_placeholder')" autocomplete="current-password" />
        <button type="submit" class="btn-premium" style="width:100%;margin-top:1rem" :disabled="!secret">{{ $t('admin.login_btn') }}</button>
      </form>
      <p v-if="loginError" class="error-msg">{{ loginError }}</p>
    </div>

    <!-- Admin Panel -->
    <div v-else>
      <div class="admin-header">
        <h2>{{ $t('admin.title') }}</h2>
        <button class="btn-text" @click="doLogout">{{ $t('admin.logout') }}</button>
      </div>

      <!-- Tab Bar -->
      <div class="admin-tabs">
        <button :class="['admin-tab', { active: tab === 'dashboard' }]" @click="tab = 'dashboard'">{{ $t('admin.tab_dashboard') }}</button>
        <button :class="['admin-tab', { active: tab === 'moderation' }]" @click="openModerationTab">{{ $t('admin.tab_moderation') }}</button>
        <button :class="['admin-tab', { active: tab === 'platforms' }]" @click="tab = 'platforms'">{{ $t('admin.tab_platforms') }}</button>
        <button :class="['admin-tab', { active: tab === 'codes' }]" @click="tab = 'codes'">{{ $t('admin.tab_codes') }}</button>
      </div>

      <!-- Dashboard Tab -->
      <div v-if="tab === 'dashboard'" class="tab-content">
        <div v-if="stats" class="stats-grid">
          <div class="stat-card">
            <div class="stat-value highlight-danger">{{ stats.reports?.total || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_total_reports') }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-value" style="color: var(--accent-neon)">{{ stats.reports?.active || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_active_reports') }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ stats.reports?.last_7d || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_reports_7d') }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-value" style="color:#facc15">{{ stats.appeals?.total || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_appeals') }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-value" style="color:var(--safe-neon)">{{ stats.codes?.used || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_codes_used') }}</div>
          </div>
          <div class="stat-card stat-card-link" @click="viewUnusedCodes">
            <div class="stat-value" style="color:var(--accent-neon)">{{ stats.codes?.unused || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_codes_unused') }} &rsaquo;</div>
          </div>
          <div class="stat-card">
            <div class="stat-value" style="color:var(--safe-neon)">{{ stats.codes?.last_7d || 0 }}</div>
            <div class="stat-label">{{ $t('admin.stat_activations_7d') }}</div>
          </div>
          <div class="stat-card wide">
            <div class="stat-value" style="color:#facc15;font-size:2rem">¥{{ (stats.revenue?.estimated_total || 0).toFixed(1) }}</div>
            <div class="stat-label">{{ $t('admin.stat_revenue') }}</div>
          </div>
        </div>

        <!-- Recent Activations -->
        <h3 style="margin-top:2rem;margin-bottom:1rem;color:var(--text-primary)">{{ $t('admin.recent_title') }}</h3>
        <div class="recent-list" v-if="stats?.recent_activations?.length">
          <div class="recent-item" v-for="(item, i) in stats.recent_activations" :key="i">
            <span class="recent-code">{{ item.code }}</span>
            <span class="recent-ip">{{ item.ip }}</span>
            <span class="recent-time">{{ formatTime(item.activated_at) }}</span>
          </div>
        </div>
        <p v-else class="empty-hint">{{ $t('admin.recent_empty') }}</p>
      </div>

      <div v-if="tab === 'moderation'" class="tab-content">
        <div class="section-header moderation-header">
          <h3>{{ $t('admin.moderation_title') }}</h3>
          <div class="moderation-actions">
            <input v-model="moderationSearch" class="input-premium moderation-search" :placeholder="$t('admin.moderation_search')" @keyup.enter="loadModerationReports(1)" />
            <button class="btn-premium btn-sm" @click="loadModerationReports(1)">{{ $t('admin.moderation_refresh') }}</button>
          </div>
        </div>

        <div class="moderation-filters">
          <div class="filter-group">
            <button :class="['btn-sm', 'filter-btn', { active: moderationKind === 'person' }]" @click="setModerationKind('person')">{{ $t('admin.moderation_person') }}</button>
            <button :class="['btn-sm', 'filter-btn', { active: moderationKind === 'company' }]" @click="setModerationKind('company')">{{ $t('admin.moderation_company') }}</button>
          </div>
          <div class="filter-group">
            <button :class="['btn-sm', 'filter-btn', { active: moderationStatus === 'active' }]" @click="setModerationStatus('active')">{{ $t('admin.status_active') }}</button>
            <button :class="['btn-sm', 'filter-btn', { active: moderationStatus === 'hidden' }]" @click="setModerationStatus('hidden')">{{ $t('admin.status_hidden') }}</button>
            <button :class="['btn-sm', 'filter-btn', { active: moderationStatus === 'all' }]" @click="setModerationStatus('all')">{{ $t('admin.status_all') }}</button>
          </div>
        </div>

        <div v-if="moderationLoading" style="text-align:center;padding:2rem 0">
          <span class="spinner"></span>
        </div>

        <div v-else class="moderation-list">
          <div v-for="item in moderationItems" :key="`${item.kind}-${item.id}`" class="moderation-card">
            <div class="moderation-card-header">
              <div>
                <div class="moderation-title-row">
                  <span class="moderation-name">{{ item.display_name }}</span>
                  <span :class="['status-pill', `status-${item.status}`]">{{ formatModerationStatus(item.status) }}</span>
                </div>
                <div class="moderation-meta">
                  <span>{{ formatTime(item.created_at) }}</span>
                  <span v-if="item.location_city">{{ item.location_city }}</span>
                  <span v-if="item.registration_no">{{ item.registration_no }}</span>
                  <span v-if="item.industry">{{ item.industry }}</span>
                </div>
              </div>
              <div class="moderation-buttons">
                <button v-if="item.status !== 'hidden'" class="btn-sm btn-text moderation-danger" @click="updateModerationStatus(item, 'hidden')">{{ $t('admin.action_hide') }}</button>
                <button v-else class="btn-sm btn-text" @click="updateModerationStatus(item, 'active')">{{ $t('admin.action_restore') }}</button>
              </div>
            </div>

            <div class="moderation-row" v-if="item.tags?.length">
              <span class="moderation-label">{{ $t('admin.moderation_tags') }}</span>
              <div class="moderation-tags">
                <span v-for="tag in item.tags" :key="tag" class="tag">{{ tag }}</span>
              </div>
            </div>

            <div class="moderation-row" v-if="item.description">
              <span class="moderation-label">{{ $t('admin.moderation_desc') }}</span>
              <p class="moderation-desc">{{ item.description }}</p>
            </div>

            <div class="moderation-row" v-if="item.reporter_display_name || item.reporter_city">
              <span class="moderation-label">{{ $t('admin.moderation_reporter') }}</span>
              <div class="moderation-meta-block">
                <span>{{ item.reporter_display_name || '-' }}</span>
                <span v-if="item.reporter_city">{{ item.reporter_city }}</span>
                <span v-if="item.verification_level">{{ $t('admin.moderation_verify', { n: item.verification_level }) }}</span>
              </div>
            </div>

            <div class="moderation-row" v-if="item.has_appeal">
              <span class="moderation-label">{{ $t('admin.moderation_appeal') }}</span>
              <p class="moderation-desc">{{ item.appeal_reason || $t('admin.moderation_no_appeal_reason') }}</p>
            </div>

            <div class="moderation-row" v-if="item.evidences?.length">
              <span class="moderation-label">{{ $t('admin.moderation_evidence') }}</span>
              <div class="moderation-links">
                <a v-for="(ev, idx) in item.evidences" :key="`${item.id}-${idx}`" :href="`/api/v1/evidence/${ev}`" target="_blank" rel="noopener" class="platform-url">
                  {{ $t('admin.moderation_evidence_item', { n: idx + 1 }) }}
                </a>
              </div>
            </div>
          </div>

          <p v-if="!moderationItems.length" class="empty-hint">{{ $t('admin.moderation_empty') }}</p>
        </div>

        <div v-if="moderationTotal > moderationPageSize" class="moderation-pagination">
          <button class="btn-sm btn-text" @click="loadModerationReports(moderationPage - 1)" :disabled="moderationPage <= 1">{{ $t('admin.page_prev') }}</button>
          <span>{{ moderationPage }} / {{ moderationPageCount }}</span>
          <button class="btn-sm btn-text" @click="loadModerationReports(moderationPage + 1)" :disabled="moderationPage >= moderationPageCount">{{ $t('admin.page_next') }}</button>
        </div>
      </div>

      <!-- Platforms Tab -->
      <div v-if="tab === 'platforms'" class="tab-content">
        <div class="section-header">
          <h3>{{ $t('admin.platforms_title') }}</h3>
          <button class="btn-premium btn-sm" @click="showPlatformForm = true">+ {{ $t('admin.platform_add') }}</button>
        </div>

        <div class="platform-list">
          <div class="platform-card" v-for="p in platforms" :key="p.id" :class="{ inactive: !p.active }">
            <div class="platform-info">
              <span class="platform-name">{{ p.icon }} {{ p.name }}</span>
              <span class="platform-region" v-if="p.region">{{ p.region }}</span>
              <a :href="p.url" target="_blank" class="platform-url">{{ p.url }}</a>
            </div>
            <div class="platform-actions">
              <button class="btn-sm btn-text" @click="editPlatform(p)">{{ $t('admin.platform_edit') }}</button>
              <button class="btn-sm btn-text" style="color:var(--danger-neon)" @click="deletePlatform(p.id)">{{ $t('admin.platform_delete') }}</button>
            </div>
          </div>
          <p v-if="!platforms.length" class="empty-hint">{{ $t('admin.platforms_empty') }}</p>
        </div>

        <!-- Platform Form Modal -->
        <div v-if="showPlatformForm" class="form-overlay" @click.self="closePlatformForm">
          <div class="form-modal glass-panel">
            <h3>{{ editingPlatform ? $t('admin.platform_edit') : $t('admin.platform_add') }}</h3>
            <div class="form-group"><label>{{ $t('admin.field_name') }}</label><input v-model="platForm.name" class="input-premium" /></div>
            <div class="form-group"><label>{{ $t('admin.field_url') }}</label><input v-model="platForm.url" class="input-premium" placeholder="https://..." /></div>
            <div class="form-group"><label>{{ $t('admin.field_icon') }}</label><input v-model="platForm.icon" class="input-premium" placeholder="e.g. WeChat / Alipay / PayPal" /></div>
            <div class="form-group"><label>{{ $t('admin.field_region') }}</label><input v-model="platForm.region" class="input-premium" placeholder="e.g. CN / Global" /></div>
            <div class="form-group"><label>{{ $t('admin.field_sort') }}</label><input v-model.number="platForm.sort_order" type="number" class="input-premium" /></div>
            <div class="form-group" style="flex-direction:row;align-items:center;gap:0.5rem">
              <input type="checkbox" v-model="platForm.active" id="plat-active" />
              <label for="plat-active">{{ $t('admin.field_active') }}</label>
            </div>
            <div class="form-actions">
              <button class="btn-text" @click="closePlatformForm">{{ $t('admin.cancel') }}</button>
              <button class="btn-premium" @click="savePlatform">{{ $t('admin.save') }}</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Codes Tab -->
      <div v-if="tab === 'codes'" class="tab-content">
        <div class="section-header">
          <h3>{{ $t('admin.codes_title') }}</h3>
        </div>
        <div class="gen-row">
          <button v-for="n in [200, 500, 1000]" :key="n"
            :class="['btn-premium', 'btn-sm', { 'btn-outline': genCount !== n }]"
            @click="genCount = n"
          >{{ n }}</button>
          <button class="btn-premium btn-sm" style="margin-left:auto" @click="generateCodes" :disabled="generating">
            <span v-if="!generating">{{ $t('admin.gen_btn') }} ({{ genCount }})</span>
            <span v-else><span class="spinner"></span></span>
          </button>
        </div>
        <div v-if="generatedCodes.length" class="gen-result">
          <div class="gen-header">
            <span>{{ $t('admin.gen_result', { n: generatedCodes.length }) }}</span>
            <button class="btn-text btn-sm" @click="copyAllCodes">{{ $t('admin.gen_copy') }}</button>
          </div>
          <textarea class="code-output" readonly :value="generatedCodes.join('\n')" rows="10"></textarea>
        </div>
      </div>

      <!-- Unused Codes Modal -->
      <div v-if="showUnusedCodes" class="form-overlay" @click.self="showUnusedCodes = false">
        <div class="form-modal glass-panel unused-codes-modal">
          <div class="unused-codes-header">
            <h3>{{ $t('admin.unused_codes_title') }} ({{ unusedCodesList.length }})</h3>
            <button class="btn-text btn-sm" @click="copyUnusedCodes">{{ $t('admin.gen_copy') }}</button>
          </div>
          <div v-if="unusedCodesLoading" style="text-align:center;padding:2rem 0">
            <span class="spinner"></span>
          </div>
          <template v-else>
            <textarea class="code-output" readonly :value="unusedCodesList.map(c => c.code).join('\n')" rows="15"></textarea>
          </template>
          <button class="btn-premium" style="margin-top:1rem;width:100%" @click="showUnusedCodes = false">{{ $t('admin.close_btn') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const authed = ref(false)
const secret = ref('')
const loginError = ref('')
const tab = ref('dashboard')

const stats = ref(null)
const platforms = ref([])
const showPlatformForm = ref(false)
const editingPlatform = ref(null)
const platForm = ref({ name: '', url: '', icon: '', region: '', sort_order: 0, active: true })

const genCount = ref(200)
const generating = ref(false)
const generatedCodes = ref([])
const showUnusedCodes = ref(false)
const unusedCodesList = ref([])
const unusedCodesLoading = ref(false)
const moderationKind = ref('person')
const moderationStatus = ref('active')
const moderationSearch = ref('')
const moderationItems = ref([])
const moderationTotal = ref(0)
const moderationPage = ref(1)
const moderationPageSize = 20
const moderationLoading = ref(false)

const moderationPageCount = computed(() => Math.max(1, Math.ceil(moderationTotal.value / moderationPageSize)))

const headers = () => ({ 'X-Admin-Secret': secret.value, 'Content-Type': 'application/json' })

const doLogin = async () => {
  loginError.value = ''
  try {
    const res = await fetch(`/api/v1/admin/stats`, { headers: { 'X-Admin-Secret': secret.value } })
    if (res.status === 403) {
      loginError.value = t('admin.login_error')
      return
    }
    stats.value = await res.json()
    authed.value = true
    sessionStorage.setItem('lt_admin_secret', secret.value)
    sessionStorage.setItem('lt_admin_ts', Date.now().toString())
    loadPlatforms()
  } catch {
    loginError.value = t('admin.login_network')
  }
}

const doLogout = () => {
  authed.value = false
  secret.value = ''
  stats.value = null
  platforms.value = []
  generatedCodes.value = []
  moderationItems.value = []
  moderationTotal.value = 0
  sessionStorage.removeItem('lt_admin_secret')
}

const loadStats = async () => {
  const res = await fetch('/api/v1/admin/stats', { headers: headers() })
  if (res.ok) stats.value = await res.json()
}

const loadPlatforms = async () => {
  const res = await fetch('/api/v1/admin/platforms', { headers: headers() })
  if (res.ok) platforms.value = await res.json()
}

const openModerationTab = () => {
  tab.value = 'moderation'
  loadModerationReports(1)
}

const setModerationKind = (kind) => {
  moderationKind.value = kind
  loadModerationReports(1)
}

const setModerationStatus = (status) => {
  moderationStatus.value = status
  loadModerationReports(1)
}

const loadModerationReports = async (page = moderationPage.value) => {
  moderationLoading.value = true
  moderationPage.value = Math.min(Math.max(page, 1), moderationPageCount.value || 1)
  try {
    const params = new URLSearchParams({
      kind: moderationKind.value,
      status: moderationStatus.value,
      page: String(moderationPage.value),
      page_size: String(moderationPageSize),
    })
    if (moderationSearch.value.trim()) {
      params.set('q', moderationSearch.value.trim())
    }
    const res = await fetch(`/api/v1/admin/reports?${params.toString()}`, { headers: headers() })
    if (res.ok) {
      const data = await res.json()
      moderationItems.value = data.items || []
      moderationTotal.value = data.total || 0
    }
  } finally {
    moderationLoading.value = false
  }
}

const updateModerationStatus = async (item, status) => {
  await fetch(`/api/v1/admin/reports/${item.kind}/${item.id}/status`, {
    method: 'PUT',
    headers: headers(),
    body: JSON.stringify({ status }),
  })
  loadModerationReports(moderationPage.value)
}

const editPlatform = (p) => {
  editingPlatform.value = p.id
  platForm.value = { ...p }
  showPlatformForm.value = true
}

const closePlatformForm = () => {
  showPlatformForm.value = false
  editingPlatform.value = null
  platForm.value = { name: '', url: '', icon: '', region: '', sort_order: 0, active: true }
}

const savePlatform = async () => {
  if (editingPlatform.value) {
    await fetch(`/api/v1/admin/platforms/${editingPlatform.value}`, {
      method: 'PUT', headers: headers(), body: JSON.stringify(platForm.value)
    })
  } else {
    await fetch('/api/v1/admin/platforms', {
      method: 'POST', headers: headers(), body: JSON.stringify(platForm.value)
    })
  }
  closePlatformForm()
  loadPlatforms()
}

const deletePlatform = async (id) => {
  await fetch(`/api/v1/admin/platforms/${id}`, { method: 'DELETE', headers: headers() })
  loadPlatforms()
}

const generateCodes = async () => {
  generating.value = true
  generatedCodes.value = []
  try {
    const res = await fetch(`/api/v1/admin/generate-codes?count=${genCount.value}`, { headers: headers() })
    const data = await res.json()
    generatedCodes.value = data.codes || []
  } finally {
    generating.value = false
    loadStats()
  }
}

const copyAllCodes = () => {
  navigator.clipboard.writeText(generatedCodes.value.join('\n'))
}

const viewUnusedCodes = async () => {
  showUnusedCodes.value = true
  unusedCodesLoading.value = true
  try {
    const res = await fetch('/api/v1/admin/unused-codes', { headers: headers() })
    if (res.ok) {
      const data = await res.json()
      unusedCodesList.value = data.codes || []
    }
  } finally {
    unusedCodesLoading.value = false
  }
}

const copyUnusedCodes = () => {
  const text = unusedCodesList.value.map(c => c.code).join('\n')
  navigator.clipboard.writeText(text)
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

const formatModerationStatus = (status) => {
  if (status === 'hidden') return t('admin.status_hidden')
  if (status === 'active') return t('admin.status_active')
  return status || '-'
}

onMounted(() => {
  const saved = sessionStorage.getItem('lt_admin_secret')
  if (saved) {
    // Check session age (auto-expire after 2 hours)
    const savedAt = sessionStorage.getItem('lt_admin_ts')
    if (savedAt && Date.now() - parseInt(savedAt) > 2 * 60 * 60 * 1000) {
      sessionStorage.removeItem('lt_admin_secret')
      sessionStorage.removeItem('lt_admin_ts')
      return
    }
    secret.value = saved
    doLogin()
  }

  // Clear session on tab/window close
  window.addEventListener('beforeunload', () => {
    sessionStorage.removeItem('lt_admin_secret')
    sessionStorage.removeItem('lt_admin_ts')
  })
})
</script>

<style scoped>
.admin-container { min-height: 400px; }
.login-gate {
  max-width: 360px;
  margin: 3rem auto;
  text-align: center;
}
.login-gate h2 { margin-bottom: 1.5rem; color: var(--text-primary); }
.error-msg { color: var(--danger-neon); margin-top: 0.75rem; font-size: 0.85rem; }

.admin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
.admin-header h2 { color: var(--text-primary); font-size: 1.3rem; }

.admin-tabs {
  display: flex;
  gap: 4px;
  background: rgba(255,255,255,0.03);
  border-radius: 10px;
  padding: 4px;
  margin-bottom: 1.5rem;
}
.admin-tab {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 8px 0;
  font-size: 0.9rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.admin-tab.active {
  background: var(--silicon-surface);
  color: var(--accent-neon);
  box-shadow: 0 0 8px rgba(0,240,255,0.15);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}
.stat-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 12px;
  padding: 1rem;
  text-align: center;
}
.stat-card.wide { grid-column: span 2; }
.stat-card-link { cursor: pointer; transition: border-color 0.2s, box-shadow 0.2s; }
.stat-card-link:hover { border-color: rgba(0,240,255,0.4); box-shadow: 0 0 12px rgba(0,240,255,0.15); }
.stat-card-link .stat-label { color: var(--accent-neon); }
.unused-codes-modal { max-width: 500px; max-height: 80vh; display: flex; flex-direction: column; }
.unused-codes-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.unused-codes-header h3 { color: var(--text-primary); font-size: 1rem; }
.unused-codes-modal .code-output { flex: 1; min-height: 200px; }
.stat-value { font-size: 1.6rem; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 0.75rem; color: var(--text-secondary); margin-top: 0.25rem; }
.btn-outline {
  background: transparent !important;
  border: 1px solid rgba(0,240,255,0.3) !important;
  color: var(--text-secondary) !important;
  box-shadow: none !important;
}
.btn-outline:hover { border-color: var(--accent-neon) !important; color: var(--accent-neon) !important; }

.recent-list { display: flex; flex-direction: column; gap: 6px; }
.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255,255,255,0.03);
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 0.82rem;
}
.recent-code { font-family: 'Courier New', monospace; color: var(--accent-neon); }
.recent-ip { color: var(--text-secondary); }
.recent-time { color: var(--text-secondary); font-size: 0.75rem; }

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
.section-header h3 { color: var(--text-primary); }
.btn-sm { font-size: 0.82rem; padding: 6px 14px; }

.platform-list { display: flex; flex-direction: column; gap: 8px; }
.platform-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 10px;
  padding: 12px 16px;
}
.platform-card.inactive { opacity: 0.5; }
.platform-info { display: flex; flex-direction: column; gap: 2px; }
.platform-name { color: var(--text-primary); font-weight: 600; }
.platform-region { font-size: 0.75rem; color: var(--accent-neon); }
.platform-url {
  font-size: 0.78rem;
  color: var(--text-secondary);
  text-decoration: none;
  word-break: break-all;
}
.platform-url:hover { color: var(--accent-neon); }
.platform-actions { display: flex; gap: 6px; flex-shrink: 0; }

.moderation-header {
  gap: 12px;
}
.moderation-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.moderation-search {
  min-width: 240px;
}
.moderation-filters {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.filter-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.filter-btn {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  color: var(--text-secondary);
  border-radius: 8px;
  cursor: pointer;
}
.filter-btn.active {
  color: var(--accent-neon);
  border-color: rgba(0,240,255,0.4);
  box-shadow: 0 0 10px rgba(0,240,255,0.12);
}
.moderation-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.moderation-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 12px;
  padding: 14px 16px;
}
.moderation-card-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.moderation-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.moderation-name {
  color: var(--text-primary);
  font-weight: 600;
}
.status-pill {
  font-size: 0.72rem;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid rgba(255,255,255,0.12);
}
.status-active {
  color: var(--safe-neon);
  border-color: rgba(0,255,102,0.3);
}
.status-hidden {
  color: var(--danger-neon);
  border-color: rgba(255,0,85,0.3);
}
.moderation-meta,
.moderation-meta-block {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  color: var(--text-secondary);
  font-size: 0.8rem;
}
.moderation-buttons {
  display: flex;
  align-items: flex-start;
}
.moderation-danger {
  color: var(--danger-neon);
}
.moderation-row {
  display: flex;
  gap: 12px;
  margin-top: 10px;
  align-items: flex-start;
}
.moderation-label {
  width: 88px;
  flex-shrink: 0;
  color: var(--text-secondary);
  font-size: 0.82rem;
}
.moderation-tags,
.moderation-links {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.moderation-desc {
  margin: 0;
  color: var(--text-primary);
  line-height: 1.5;
  white-space: pre-wrap;
}
.moderation-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 1rem;
}

.form-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}
.form-modal {
  max-width: 420px;
  width: 100%;
  padding: 1.5rem;
}
.form-modal h3 { color: var(--text-primary); margin-bottom: 1rem; }
.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 0.75rem;
}
.form-group label { font-size: 0.82rem; color: var(--text-secondary); }
.form-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 1rem; }

.gen-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1rem;
}
.gen-row label { color: var(--text-secondary); font-size: 0.9rem; white-space: nowrap; }
.gen-result { margin-top: 1rem; }
.gen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
}
.code-output {
  width: 100%;
  background: rgba(0,0,0,0.3);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  color: var(--accent-neon);
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  padding: 12px;
  resize: vertical;
}
.empty-hint { color: var(--text-secondary); font-size: 0.85rem; text-align: center; padding: 2rem 0; }

@media (max-width: 480px) {
  .admin-header h2 { font-size: 1.1rem; }
  .admin-tabs { gap: 2px; padding: 3px; }
  .admin-tab { font-size: 0.78rem; padding: 7px 0; }

  .stats-grid { grid-template-columns: repeat(2, 1fr); gap: 8px; }
  .stat-card { padding: 0.7rem; }
  .stat-value { font-size: 1.3rem; }
  .stat-card.wide .stat-value { font-size: 1.5rem !important; }
  .stat-label { font-size: 0.68rem; }

  .recent-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    padding: 8px 10px;
  }
  .recent-code { font-size: 0.78rem; }
  .recent-ip { font-size: 0.72rem; }
  .recent-time { font-size: 0.68rem; }

  .gen-row { flex-wrap: wrap; gap: 8px; }
  .gen-row .btn-sm { flex: 1; min-width: 50px; text-align: center; }
  .gen-row .btn-sm:last-child { flex-basis: 100%; margin-left: 0; }

  .platform-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .platform-actions { width: 100%; justify-content: flex-end; }
  .platform-url { font-size: 0.72rem; }

  .section-header { flex-direction: column; align-items: flex-start; gap: 8px; }
  .moderation-actions { width: 100%; flex-direction: column; align-items: stretch; }
  .moderation-search { min-width: 0; width: 100%; }
  .moderation-card-header,
  .moderation-row { flex-direction: column; }
  .moderation-label { width: auto; }

  .form-modal { padding: 1.2rem; }
  .code-output { font-size: 0.78rem; padding: 8px; }
}
</style>

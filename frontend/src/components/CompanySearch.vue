<template>
  <div class="company-search-container glass-panel">
    <h2 style="margin-bottom: 1rem; font-weight: 600;">{{ $t('companySearch.title') }}</h2>
    <p style="color: var(--text-secondary); margin-bottom: 2rem; font-size: 0.95rem;">
      {{ $t('companySearch.subtitle') }}
    </p>

    <!-- Search Form -->
    <div v-if="!searched" class="search-box">
      <div class="form-group">
        <label>{{ $t('companySearch.registrationNoPlaceholder') }}</label>
        <input
          v-model="searchForm.registrationNo"
          type="text"
          class="input-premium"
          :placeholder="$t('companySearch.registrationNoPlaceholder')"
          @keyup.enter="searchCompany"
        />
      </div>
      <div class="form-group">
        <label>{{ $t('companySearch.companyNamePlaceholder') }}</label>
        <input
          v-model="searchForm.companyName"
          type="text"
          class="input-premium"
          :placeholder="$t('companySearch.companyNamePlaceholder')"
          @keyup.enter="searchCompany"
        />
      </div>
      <button class="btn-premium search-btn" @click="searchCompany" :disabled="isSearching">
        <span v-if="!isSearching">{{ $t('companySearch.search') }}</span>
        <span v-else class="loading-state">
          <span class="spinner"></span> {{ $t('companySearch.searching') }}
        </span>
      </button>
      <p v-if="errorMessage" class="inline-alert" @click="errorMessage = ''">{{ errorMessage }}</p>
    </div>

    <!-- Results Area -->
    <div v-else class="results-box">
      <div class="actions-top">
        <button class="back-btn" @click="resetSearch">{{ $t('search.btn_reset') }}</button>
      </div>

      <!-- Clean Result -->
      <div v-if="searchResult && searchResult.status === 'clean'" class="result-card clean-card">
        <div class="icon-safe"></div>
        <h3 style="color: var(--safe-color);">{{ $t('companySearch.cleanTitle') }}</h3>
        <p class="desc">{{ searchResult.message }}</p>
      </div>

      <!-- Warning Result -->
      <div v-if="searchResult && searchResult.status === 'warning'" class="result-card warning-card">
        <div class="radar-warning-icon"></div>
        <h3 style="color: var(--danger-color); font-size: 1.5rem; margin-bottom: 1rem">
          {{ searchResult.message }}
        </h3>

        <!-- Aggregated Profile -->
        <div class="profile-section">
          <h4 style="color: var(--text-primary); margin-bottom: 1rem;">{{ $t('companySearch.companyProfile') }}</h4>
          <div class="info-grid">
            <div class="info-row">
              <span class="info-label">{{ $t('companySearch.companyName') }}</span>
              <span class="info-value">{{ searchResult.aggregated_profile.display_name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ $t('companySearch.totalReports') }}</span>
              <span class="info-value danger-text">{{ searchResult.aggregated_profile.total_independent_reports }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ $t('companySearch.riskScore') }}</span>
              <span class="info-value" :class="getRiskClass(searchResult.aggregated_profile.risk_level)">
                {{ searchResult.aggregated_profile.consensus_risk_score }} / 100
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ $t('companySearch.riskLevel') }}</span>
              <span class="info-value" :class="getRiskClass(searchResult.aggregated_profile.risk_level)">
                {{ getRiskLevelText(searchResult.aggregated_profile.risk_level) }}
              </span>
            </div>
            <div class="info-row full-width">
              <span class="info-label">{{ $t('companySearch.locations') }}</span>
              <span class="info-value">{{ searchResult.aggregated_profile.active_cities.join(', ') || 'N/A' }}</span>
            </div>
            <div class="info-row full-width">
              <span class="info-label">{{ $t('companySearch.tags') }}</span>
              <div class="tags-display">
                <span v-for="tag in searchResult.aggregated_profile.consolidated_tags" :key="tag" class="tag-badge">
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>

          <!-- Time Range -->
          <div class="time-info">
            <p>{{ $t('companySearch.firstReport') }}: {{ formatDate(searchResult.aggregated_profile.first_report_at) }}</p>
            <p>{{ $t('companySearch.latestReport') }}: {{ formatDate(searchResult.aggregated_profile.latest_report_at) }}</p>
          </div>

          <!-- Appeal Info -->
          <div v-if="searchResult.aggregated_profile.has_appeal" class="appeal-box">
            <h5 style="color: var(--text-primary); margin-bottom: 0.75rem;">{{ $t('companySearch.companyAppeal') }}</h5>
            <p class="appeal-text">{{ searchResult.aggregated_profile.appeal_reason }}</p>
            <div class="vote-display">
              <div class="vote-stat">
                <span class="vote-label">{{ $t('companySearch.reporterVotes') }}</span>
                <span class="vote-number">{{ searchResult.aggregated_profile.reporter_votes }}</span>
              </div>
              <div class="vote-stat">
                <span class="vote-label">{{ $t('companySearch.companyVotes') }}</span>
                <span class="vote-number">{{ searchResult.aggregated_profile.company_votes }}</span>
              </div>
            </div>
            <div class="vote-buttons">
              <button class="btn-vote" @click="vote('reporter')" :disabled="hasVoted">
                {{ $t('companySearch.supportReporter') }}
              </button>
              <button class="btn-vote" @click="vote('company')" :disabled="hasVoted">
                {{ $t('companySearch.supportCompany') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Individual Reports -->
        <div class="records-section">
          <h4 style="color: var(--text-primary); margin: 1.5rem 0 1rem;">
            {{ $t('companySearch.individualReports') }} ({{ searchResult.records.length }})
          </h4>
          <div v-for="record in searchResult.records" :key="record.id" class="record-item">
            <div class="record-header">
              <span class="record-reporter">{{ record.reporter_display_name }}</span>
              <span class="record-date">{{ formatDate(record.created_at) }}</span>
            </div>
            <div class="record-meta">
              <span v-if="record.position">{{ $t('companySearch.position') }}: {{ record.position }}</span>
              <span v-if="record.employment_period">{{ $t('companySearch.period') }}: {{ record.employment_period }}</span>
              <span v-if="record.location_city">{{ $t('companySearch.location') }}: {{ record.location_city }}</span>
            </div>
            <div class="record-tags">
              <span v-for="tag in record.tags" :key="tag" class="tag-badge">{{ tag }}</span>
            </div>
            <p class="record-desc">{{ record.description }}</p>
            <div v-if="record.evidences && record.evidences.length > 0" class="evidence-grid">
              <img
                v-for="(evidence, idx) in record.evidences"
                :key="idx"
                :src="`/api/v1/evidence/${evidence}`"
                @click="openEvidence(evidence)"
                class="evidence-img"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const searchForm = ref({
  registrationNo: '',
  companyName: ''
})
const isSearching = ref(false)
const searched = ref(false)
const searchResult = ref(null)
const errorMessage = ref('')
const hasVoted = ref(false)

const searchCompany = async () => {
  if (!searchForm.value.registrationNo.trim()) {
    errorMessage.value = t('companySearch.registrationNoRequired')
    return
  }

  isSearching.value = true
  errorMessage.value = ''
  searchResult.value = null

  try {
    const params = new URLSearchParams({
      registration_no: searchForm.value.registrationNo,
      company_name: searchForm.value.companyName || ''
    })

    const response = await fetch(`/api/v1/company/query?${params}`)
    const data = await response.json()

    if (response.ok) {
      searchResult.value = data
      searched.value = true
    } else {
      errorMessage.value = data.error || t('companySearch.searchError')
    }
  } catch (error) {
    errorMessage.value = t('companySearch.networkError')
  } finally {
    isSearching.value = false
  }
}

const resetSearch = () => {
  searched.value = false
  searchResult.value = null
  errorMessage.value = ''
  hasVoted.value = false
}

const vote = async (side) => {
  try {
    const response = await fetch('/api/v1/company/vote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        registration_no: searchForm.value.registrationNo,
        side: side,
        fingerprint: getFingerprint()
      })
    })

    const data = await response.json()

    if (response.ok) {
      hasVoted.value = true
      alert(t('companySearch.voteSuccess'))
      searchCompany() // Refresh results
    } else {
      alert(data.error || t('companySearch.voteError'))
    }
  } catch (error) {
    alert(t('companySearch.networkError'))
  }
}

const getRiskClass = (level) => {
  if (level === 'high') return 'risk-high'
  if (level === 'medium') return 'risk-medium'
  return 'risk-low'
}

const getRiskLevelText = (level) => {
  const levels = {
    high: t('companySearch.riskHigh'),
    medium: t('companySearch.riskMedium'),
    low: t('companySearch.riskLow')
  }
  return levels[level] || level
}

const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

const openEvidence = (evidence) => {
  window.open(`/api/v1/evidence/${evidence}`, '_blank')
}

const getFingerprint = () => {
  const nav = navigator
  return btoa(`${nav.userAgent}|${nav.language}|${screen.width}x${screen.height}`)
}
</script>

<style scoped>
.company-search-container {
  max-width: 900px;
  margin: 0 auto;
}

.search-box {
  margin-bottom: 2rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.search-btn {
  width: 100%;
  margin-top: 1rem;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.inline-alert {
  margin-top: 1rem;
  padding: 0.75rem;
  background: rgba(255, 0, 85, 0.1);
  border: 1px solid var(--danger-neon);
  border-radius: var(--radius-md);
  color: var(--danger-neon);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.inline-alert:hover {
  background: rgba(255, 0, 85, 0.15);
}

.results-box {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.actions-top {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 1.5rem;
}

.back-btn {
  background: var(--silicon-surface);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary);
  padding: 0.5rem 1rem;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
}

.back-btn:hover {
  color: var(--accent-neon);
  border-color: var(--accent-neon);
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
}

.result-card {
  background: var(--silicon-surface);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-lg);
  padding: 2rem;
  box-shadow: var(--shadow-out);
}

.clean-card {
  text-align: center;
  padding: 3rem 2rem;
}

.icon-safe {
  width: 80px;
  height: 80px;
  margin: 0 auto 1.5rem;
  background: radial-gradient(circle, rgba(76, 175, 80, 0.2) 0%, transparent 70%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.icon-safe::before {
  content: '✓';
  font-size: 3rem;
  color: var(--safe-color);
  text-shadow: 0 0 20px rgba(76, 175, 80, 0.5);
}

.desc {
  color: var(--text-secondary);
  font-size: 1rem;
  line-height: 1.6;
  margin-top: 1rem;
}

.warning-card {
  text-align: center;
}

.radar-warning-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  background: radial-gradient(circle, rgba(255, 0, 85, 0.2) 0%, transparent 70%);
  border-radius: 50%;
  position: relative;
  animation: pulse 2s ease-in-out infinite;
}

.radar-warning-icon::before {
  content: '⚠';
  font-size: 3.5rem;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: var(--danger-neon);
  text-shadow: 0 0 20px rgba(255, 0, 85, 0.5);
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.05); opacity: 0.9; }
}

.profile-section {
  text-align: left;
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.info-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.info-row.full-width {
  grid-column: 1 / -1;
}

.info-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.info-value {
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 600;
}

.danger-text {
  color: var(--danger-neon);
  font-size: 1.3rem;
}

.risk-high {
  color: var(--danger-neon);
}

.risk-medium {
  color: #facc15;
}

.risk-low {
  color: var(--safe-color);
}

.tags-display {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.tag-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: rgba(0, 240, 255, 0.1);
  border: 1px solid rgba(0, 240, 255, 0.3);
  border-radius: 12px;
  font-size: 0.8rem;
  color: var(--accent-neon);
}

.time-info {
  margin-top: 1.5rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.time-info p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin: 0.25rem 0;
}

.appeal-box {
  margin-top: 1.5rem;
  padding: 1.5rem;
  background: rgba(33, 150, 243, 0.1);
  border: 1px solid rgba(33, 150, 243, 0.3);
  border-radius: var(--radius-md);
}

.appeal-text {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.6;
  margin-bottom: 1rem;
}

.vote-display {
  display: flex;
  gap: 2rem;
  margin-bottom: 1rem;
}

.vote-stat {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.vote-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.vote-number {
  font-size: 1.5rem;
  font-weight: 700;
  color: #2196f3;
}

.vote-buttons {
  display: flex;
  gap: 0.75rem;
}

.btn-vote {
  flex: 1;
  padding: 0.75rem;
  background: rgba(33, 150, 243, 0.2);
  border: 1px solid rgba(33, 150, 243, 0.5);
  color: #2196f3;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-vote:hover:not(:disabled) {
  background: rgba(33, 150, 243, 0.3);
  box-shadow: 0 0 10px rgba(33, 150, 243, 0.3);
}

.btn-vote:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.records-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.record-item {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.record-reporter {
  font-weight: 600;
  color: var(--text-primary);
}

.record-date {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.record-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 0.75rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.record-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.record-desc {
  color: var(--text-secondary);
  line-height: 1.6;
  margin-bottom: 1rem;
}

.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 0.75rem;
}

.evidence-img {
  width: 100%;
  height: 100px;
  object-fit: cover;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: transform 0.2s;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.evidence-img:hover {
  transform: scale(1.05);
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.3);
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }

  .vote-display {
    flex-direction: column;
    gap: 1rem;
  }

  .vote-buttons {
    flex-direction: column;
  }

  .record-meta {
    flex-direction: column;
    gap: 0.5rem;
  }
}
</style>

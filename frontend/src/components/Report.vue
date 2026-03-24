<template>
  <div class="report-container glass-panel">
    <h2 style="margin-bottom: 1rem; font-weight: 600;">{{ $t('report.title') }}</h2>
    <p v-if="step !== 4" style="color: var(--text-secondary); margin-bottom: 2rem; font-size: 0.95rem;">
      {{ $t('report.subtitle') }}
    </p>

    <p v-if="errorMsg && step < 4" class="error-msg" @click="errorMsg = ''">{{ errorMsg }}</p>

    <!-- Progress Tracker -->
    <div v-if="step < 4" class="progress-bar">
      <div :class="['step', { active: step >= 1 }]">{{ $t('report.step1') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 2 }]">{{ $t('report.step2') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 3 }]">{{ $t('report.step3') }}</div>
    </div>

    <!-- Step 1: Reporter Info -->
    <div v-if="step === 1" class="step-content">
      <div class="form-group">
        <label>{{ $t('report.step1_label') }}</label>
        <PhoneInput
          v-model="form.reporter_phone"
          v-model:countryCode="reporterCC"
          @validity="v => reporterPhoneValid = v"
          :placeholder="$t('report.step1_placeholder')"
        />
      </div>
      
      <div class="action-buttons" style="justify-content: flex-end;">
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!reporterPhoneValid">
          {{ $t('report.next') }}
        </button>
      </div>
    </div>

    <!-- Step 2: Target Info & Evidence -->
    <div v-if="step === 2" class="step-content">
      <div class="form-group">
        <label>{{ $t('report.target_phone_label') }}</label>
        <PhoneInput
          v-model="form.target_phone"
          v-model:countryCode="targetCC"
          @validity="v => targetPhoneValid = v"
          :placeholder="$t('report.target_phone_placeholder')"
        />
      </div>
      <div class="form-group">
        <label>{{ $t('report.target_name_label') }}</label>
        <input v-model="form.target_name" type="text" class="input-premium" :placeholder="$t('report.target_name_placeholder')" required />
      </div>

      <div class="form-group">
        <label>{{ $t('report.city_label') }}</label>
        <input v-model="form.location_city" type="text" class="input-premium" :placeholder="$t('report.city_placeholder')" required />
      </div>

      <div class="form-group">
        <label>{{ $t('report.tags_label') }}</label>
        <div class="tags-grouped">
          <div class="tag-category" v-for="cat in tagCategories" :key="cat.key">
            <div class="tag-category-title">{{ $t('report.cat_' + cat.key) }}</div>
            <div class="tags-selector">
              <button
                type="button"
                v-for="key in cat.tags"
                :key="key"
                :class="['tag-btn', { selected: form.selectedTags.includes(key) }]"
                @click="toggleTag(key)"
              >
                {{ $t('report.tag_' + key) }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="form-group">
        <label>{{ $t('report.description_label') }}</label>
        <textarea
          v-model="form.description"
          class="input-premium textarea-premium"
          :placeholder="$t('report.description_placeholder')"
          rows="4"
          maxlength="2000"
        ></textarea>
      </div>

      <div class="form-group">
        <label>{{ $t('report.evidence_label') }}</label>
        
        <div class="evidence-grid" v-if="form.evidenceFiles.length > 0">
          <div class="evidence-item" v-for="(file, index) in form.evidenceFiles" :key="index">
            <span class="file-name">{{ file.name }}</span>
            <button type="button" class="remove-btn" @click.stop="removeFile(index)">&times;</button>
          </div>
        </div>

        <div class="upload-area" @click="triggerFileInput" v-if="form.evidenceFiles.length < 9">
          <span class="upload-hint">{{ $t('report.evidence_hint', { n: form.evidenceFiles.length }) }}</span>
        </div>
        
        <input type="file" ref="fileInput" @change="handleFileChange" accept="image/*" multiple style="display: none;" />
      </div>

      <div class="action-buttons">
        <button class="btn-text" @click="step = 1">{{ $t('report.prev') }}</button>
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!targetPhoneValid || !form.target_name || form.selectedTags.length === 0">
          {{ $t('report.to_confirm') }}
        </button>
      </div>
    </div>

    <!-- Step 3: Confirmation Layer -->
    <div v-if="step === 3" class="step-content">
      <div class="confirm-box">
        <h3 style="color: var(--danger-neon); margin-bottom: 1rem;">{{ $t('report.confirm_title') }}</h3>
        <p style="color: var(--text-secondary); font-size: 0.95rem; line-height: 1.6;">
          {{ $t('report.confirm_desc') }}
        </p>
        <p style="color: var(--text-secondary); font-size: 0.95rem; line-height: 1.6; margin-top: 15px;">
          <span style="color:var(--text-primary)">{{ $t('report.confirm_pledge') }}</span>
        </p>
      </div>

      <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>

      <div class="action-buttons">
        <button class="btn-text" @click="step = 2" :disabled="submitting">{{ $t('report.withdraw') }}</button>
        <button class="btn-premium danger-btn" @click="submitReport" :disabled="submitting">
          <span v-if="!submitting">{{ $t('report.submit') }}</span>
          <span v-else><span class="spinner"></span> {{ $t('report.submitting') }}</span>
        </button>
      </div>
    </div>

    <!-- Step 4: Success Result -->
    <div v-if="step === 4" class="step-content success-view">
      <div class="success-icon"></div>
      <h3 style="color: var(--safe-neon); font-size: 1.5rem; margin-bottom: 1rem;">{{ $t('report.success_title') }}</h3>
      <p style="color: var(--text-secondary); font-size: 1rem; line-height: 1.6;">
        {{ $t('report.success_msg') }}
      </p>
      <p style="color: var(--text-secondary); font-size: 0.9rem; margin-top: 15px;">
        {{ $t('report.success_note') }}
      </p>
      <button class="btn-premium" @click="resetForm" style="margin-top: 2rem; width: 100%;">
        {{ $t('report.reset') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PhoneInput from './PhoneInput.vue'
import { getDefaultCountry, getDialCode } from '../data/countryCodes.js'

const { t, locale } = useI18n()

const step = ref(1)

const fileInput = ref(null)
const submitting = ref(false)
const errorMsg = ref('')

const reporterCC = ref(getDefaultCountry(locale.value))
const targetCC = ref(getDefaultCountry(locale.value))
const reporterPhoneValid = ref(false)
const targetPhoneValid = ref(false)

const tagCategories = [
  {
    key: 'integrity',
    tags: ['identity_fraud', 'habitual_lying', 'hidden_marriage', 'child_concealment', 'criminal_record']
  },
  {
    key: 'relational',
    tags: ['cheating', 'multi_timing', 'emotional_abuse', 'pua', 'ghosting', 'exploitation']
  },
  {
    key: 'financial',
    tags: ['financial_dispute', 'property_fraud', 'romance_scam', 'debt_concealment', 'info_theft', 'gambling', 'drug_abuse']
  },
  {
    key: 'safety',
    tags: ['violent_tendency', 'stalking', 'privacy_threat', 'verbal_abuse', 'cyber_bullying', 'over_control']
  },
  {
    key: 'background',
    tags: ['hidden_sex_work', 'hidden_nightlife', 'hidden_disease', 'hidden_marriages', 'hidden_illegal_habits', 'career_fabrication']
  }
]

const form = ref({
  reporter_phone: '',
  target_phone: '',
  target_name: '',
  location_city: '',
  description: '',
  selectedTags: [],
  evidenceFiles: []
})

const nextStep = () => {
  errorMsg.value = ''
  if (step.value === 1 && !reporterPhoneValid.value) {
    errorMsg.value = t('report.alert_phone_required')
    return
  }
  if (step.value === 2 && form.value.selectedTags.length === 0) {
    errorMsg.value = t('report.alert_tags_required')
    return
  }
  step.value++
}

const toggleTag = (tag) => {
  const idx = form.value.selectedTags.indexOf(tag)
  if (idx > -1) {
    form.value.selectedTags.splice(idx, 1)
  } else {
    form.value.selectedTags.push(tag)
  }
}

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileChange = (e) => {
  const files = Array.from(e.target.files)
  if (files.length === 0) return
  
  if (form.value.evidenceFiles.length + files.length > 9) {
    errorMsg.value = t('report.alert_max_files')
    return
  }
  form.value.evidenceFiles.push(...files)
  // reset file input value in case same files selected again
  e.target.value = ''
}

const removeFile = (index) => {
  form.value.evidenceFiles.splice(index, 1)
}

const resetForm = () => {
  form.value = { reporter_phone:'', target_phone:'', target_name:'', location_city:'', description:'', selectedTags:[], evidenceFiles:[] }
  reporterCC.value = getDefaultCountry(locale.value)
  targetCC.value = getDefaultCountry(locale.value)
  step.value = 1
}

const submitReport = async () => {
  submitting.value = true
  errorMsg.value = ''
  
  const formData = new FormData()
  formData.append('reporter_phone', getDialCode(reporterCC.value) + form.value.reporter_phone)
  formData.append('reporter_phone_local', form.value.reporter_phone)
  formData.append('target_phone', getDialCode(targetCC.value) + form.value.target_phone)
  formData.append('target_phone_local', form.value.target_phone)
  formData.append('target_name', form.value.target_name)
  formData.append('location_city', form.value.location_city)
  formData.append('description', form.value.description)
  formData.append('tags', JSON.stringify(form.value.selectedTags))
  
  for (let file of form.value.evidenceFiles) {
    formData.append('evidence_files[]', file)
  }

  try {
    const res = await fetch('/api/v1/report', {
      method: 'POST',
      body: formData
    })
    const data = await res.json()
    
    if (res.ok) {
      step.value = 4 // Success Screen Jump
    } else {
      errorMsg.value = t('report.error_submit_failed')
    }
  } catch(e) {
    errorMsg.value = t('report.error_network')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.textarea-premium {
  width: 100%;
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
  font-size: 0.95rem;
  line-height: 1.5;
}
.progress-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2.5rem;
}
.step {
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.3s;
}
.step.active {
  color: var(--accent-neon);
  text-shadow: 0 0 8px rgba(0, 240, 255, 0.4);
}
.line {
  flex: 1;
  height: 2px;
  background: rgba(255,255,255,0.05);
  margin: 0 10px;
}

.step-content {
  animation: fadeIn 0.4s ease;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-group {
  margin-bottom: 1.5rem;
  text-align: left;
}
.form-group label {
  display: block;
  font-size: 0.95rem;
  color: var(--text-primary);
  margin-bottom: 0.8rem;
  font-weight: 500;
}
.row-group {
  display: flex;
  gap: 1rem;
}
.half {
  flex: 1;
}

.tags-grouped {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}
.tag-category-title {
  font-size: 0.8rem;
  color: var(--accent-neon);
  margin-bottom: 0.4rem;
  letter-spacing: 0.5px;
  opacity: 0.85;
  border-left: 2px solid var(--accent-neon);
  padding-left: 8px;
}
.tags-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.tag-btn {
  background: var(--silicon-bg);
  border: 1px solid rgba(0, 240, 255, 0.05);
  color: var(--text-secondary);
  padding: 10px 18px;
  border-radius: var(--radius-xl);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  box-shadow: var(--shadow-in);
  font-weight: 500;
}
.tag-btn.selected {
  background: var(--silicon-surface);
  color: var(--accent-neon);
  box-shadow: var(--shadow-out), 0 0 10px rgba(0, 240, 255, 0.2);
  border-color: rgba(0, 240, 255, 0.4);
}

.upload-area {
  background: var(--silicon-bg);
  box-shadow: var(--shadow-in);
  border-radius: var(--radius-lg);
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px dashed rgba(255, 255, 255, 0.1);
  margin-top: 10px;
}
.upload-area:hover {
  box-shadow: var(--shadow-out);
  color: var(--accent-neon);
  border-color: rgba(0, 240, 255, 0.3);
}
.upload-hint {
  color: var(--text-secondary);
  font-size: 0.95rem;
  font-weight: 500;
  letter-spacing: 0.5px;
}

/* 9-Image Grid List layout */
.evidence-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}
.evidence-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.03);
  padding: 10px 15px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
.file-name {
  color: var(--accent-neon);
  font-size: 0.85rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80%;
}
.remove-btn {
  background: transparent;
  border: none;
  color: var(--danger-neon);
  cursor: pointer;
  font-size: 1rem;
  font-weight: bold;
}
.remove-btn:hover {
  text-shadow: 0 0 5px var(--danger-neon);
  transform: scale(1.1);
}
.highlight {
  color: var(--accent-neon);
  font-weight: 500;
}

.action-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 2rem;
}
.next-btn {
  min-width: 150px;
}
.btn-text {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1rem;
  cursor: pointer;
}
.btn-text:hover {
  color: var(--text-primary);
}

.danger-btn {
  background: var(--silicon-bg);
  color: var(--danger-neon);
  border-color: rgba(255, 0, 85, 0.3);
}
.danger-btn:hover {
  border-color: var(--danger-neon);
  text-shadow: 0 0 8px var(--danger-neon);
  box-shadow: var(--shadow-out), 0 0 15px rgba(255, 0, 85, 0.3);
}

.confirm-box {
  background: var(--silicon-bg);
  padding: 24px;
  border-radius: var(--radius-lg);
  border: 1px dashed rgba(255, 0, 85, 0.3);
  box-shadow: var(--shadow-in);
  text-align: center;
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

.error-msg {
  color: var(--danger-neon);
  font-size: 0.85rem;
  text-align: center;
  margin-bottom: 1rem;
  padding: 10px;
  background: rgba(255, 0, 85, 0.08);
  border: 1px solid rgba(255, 0, 85, 0.2);
  border-radius: 8px;
  cursor: pointer;
}

/* Success View */
.success-view {
  text-align: center;
  padding: 30px 10px;
}
.success-icon {
  font-size: 4.5rem;
  color: var(--safe-neon);
  margin-bottom: 2rem;
  text-shadow: 0 0 20px rgba(0, 255, 102, 0.4);
}

@media (max-width: 480px) {
  .progress-bar { margin-bottom: 1.5rem; }
  .step { font-size: 0.75rem; }
  .line { margin: 0 6px; }

  .tag-btn { padding: 7px 12px; font-size: 0.8rem; }
  .tags-selector { gap: 6px; }
  .tag-category-title { font-size: 0.72rem; }

  .form-group { margin-bottom: 1rem; }
  .form-group label { font-size: 0.85rem; margin-bottom: 0.5rem; }

  .upload-area { padding: 16px; }
  .upload-hint { font-size: 0.85rem; }

  .action-buttons { margin-top: 1.2rem; }
  .next-btn { min-width: 120px; }

  .confirm-box { padding: 16px; }
  .success-view { padding: 20px 8px; }
  .success-icon { font-size: 3.5rem; margin-bottom: 1.5rem; }
}
</style>

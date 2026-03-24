<template>
  <div class="appeal-container glass-panel">
    <h2 style="margin-bottom: 1rem; font-weight: 600; color: #facc15;">{{ $t('appeal.title') }}</h2>
    <p style="color: var(--text-secondary); margin-bottom: 1.5rem; font-size: 0.9rem;">
      {{ $t('appeal.subtitle') }}
    </p>

    <!-- Legal Disclaimer Notice -->
    <div class="legal-notice">
      <div class="legal-title">{{ $t('appeal.legal_title') }}</div>
      <p>{{ $t('appeal.legal_text') }}</p>
    </div>

    <p v-if="errorMsg && step < 4" class="error-msg" @click="errorMsg = ''">{{ errorMsg }}</p>

    <!-- Progress Tracker -->
    <div v-if="step < 4" class="progress-bar appeal-progress">
      <div :class="['step', { active: step >= 1 }]">{{ $t('appeal.step1') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 2 }]">{{ $t('appeal.step2') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 3 }]">{{ $t('appeal.step3') }}</div>
    </div>

    <!-- Step 1 -->
    <div v-if="step === 1" class="step-content">
      <div class="form-group">
        <label>{{ $t('appeal.contact_label') }}</label>
        <PhoneInput
          v-model="form.contact_phone"
          v-model:countryCode="contactCC"
          @validity="v => contactPhoneValid = v"
          :placeholder="$t('appeal.contact_placeholder')"
        />
      </div>

      <div class="action-buttons" style="justify-content: flex-end;">
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!contactPhoneValid">
          {{ $t('appeal.next') }}
        </button>
      </div>
    </div>

    <!-- Step 2 -->
    <div v-if="step === 2" class="step-content">
      <div class="form-group row-group">
        <div class="half">
          <label>{{ $t('appeal.target_label') }}</label>
          <PhoneInput
            v-model="form.target_phone"
            v-model:countryCode="targetCC"
            @validity="v => targetPhoneValid = v"
            :placeholder="$t('appeal.target_placeholder')"
          />
        </div>
      </div>

      <div class="form-group">
        <label>{{ $t('appeal.reason_label') }}</label>
        <textarea v-model="form.reason" class="input-premium" rows="4" :placeholder="$t('appeal.reason_placeholder')" required></textarea>
      </div>

      <div class="form-group">
        <label>{{ $t('appeal.evidence_label') }}</label>
        <div class="evidence-grid" v-if="form.evidenceFiles.length > 0">
          <div class="evidence-item" v-for="(file, index) in form.evidenceFiles" :key="index">
            <span class="file-name">{{ file.name }}</span>
            <button type="button" class="remove-btn" @click.stop="removeFile(index)">×</button>
          </div>
        </div>
        <div class="upload-area" @click="triggerFileInput" v-if="form.evidenceFiles.length < 9">
          <span class="upload-hint">{{ $t('appeal.evidence_hint', { n: form.evidenceFiles.length }) }}</span>
        </div>
        <input type="file" ref="fileInput" @change="handleFileChange" accept="image/*" multiple style="display: none;" />
      </div>

      <div class="action-buttons">
        <button class="btn-text" @click="step = 1">{{ $t('appeal.prev_step1') }}</button>
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!form.target_phone || !form.reason">
          {{ $t('appeal.to_confirm') }}
        </button>
      </div>
    </div>

    <!-- Step 3: Confirmation -->
    <div v-if="step === 3" class="step-content">
      <div class="confirm-box" style="border-color: rgba(250, 204, 21, 0.4);">
        <h3 style="color: #facc15; margin-bottom: 1rem;">{{ $t('appeal.confirm_title') }}</h3>
        <p style="color: var(--text-secondary); font-size: 0.95rem; line-height: 1.6;">
          <strong style="color:var(--text-primary)">{{ $t('appeal.confirm_desc') }}</strong>
        </p>
        <p style="color: var(--text-secondary); font-size: 0.95rem; line-height: 1.6; margin-top: 15px;">
          <span style="color:var(--text-primary)">{{ $t('appeal.confirm_pledge') }}</span>
        </p>
      </div>

      <div class="action-buttons">
        <button class="btn-text" @click="step = 2" :disabled="submitting">{{ $t('appeal.add_more') }}</button>
        <button class="btn-premium submit-btn" style="margin-top:0;" @click="submitAppeal" :disabled="submitting">
          <span v-if="!submitting">{{ $t('appeal.submit') }}</span>
          <span v-else><span class="spinner"></span> {{ $t('appeal.submitting') }}</span>
        </button>
      </div>
    </div>

    <!-- Step 4: Success View -->
    <div v-if="step === 4" class="step-content success-view">
      <div class="success-icon" style="color: #facc15;"></div>
      <h3 style="color: #facc15; font-size: 1.5rem; margin-bottom: 1rem;">{{ $t('appeal.success_title') }}</h3>
      <p style="color: var(--text-secondary); font-size: 1rem; line-height: 1.6;">
        {{ $t('appeal.success_msg') }}
      </p>
      <p style="color: var(--text-secondary); font-size: 0.9rem; margin-top: 15px;">
        {{ $t('appeal.success_note') }}
      </p>
      <button class="btn-text" @click="resetForm" style="margin-top: 2rem;">
        {{ $t('appeal.back') }}
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

const contactCC = ref(getDefaultCountry(locale.value))
const targetCC = ref(getDefaultCountry(locale.value))
const contactPhoneValid = ref(false)
const targetPhoneValid = ref(false)

const form = ref({
  contact_phone: '',
  target_phone: '',
  reason: '',
  evidenceFiles: []
})

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileChange = (e) => {
  const files = Array.from(e.target.files)
  if (files.length === 0) return
  
  if (form.value.evidenceFiles.length + files.length > 9) {
    errorMsg.value = t('appeal.alert_max_files')
    return
  }
  form.value.evidenceFiles.push(...files)
  e.target.value = ''
}

const removeFile = (index) => {
  form.value.evidenceFiles.splice(index, 1)
}

const nextStep = () => {
  errorMsg.value = ''
  if (step.value === 1 && !contactPhoneValid.value) {
    errorMsg.value = t('appeal.alert_phone_required')
    return
  }
  if (step.value === 2 && (!form.value.target_phone || !form.value.reason)) {
    errorMsg.value = t('appeal.alert_fields_required')
    return
  }
  step.value++
}

const resetForm = () => {
  form.value = { contact_phone:'', target_phone:'', reason:'', evidenceFiles:[] }
  contactCC.value = getDefaultCountry(locale.value)
  targetCC.value = getDefaultCountry(locale.value)
  step.value = 1
}

const submitAppeal = async () => {
  submitting.value = true
  const formData = new FormData()
  formData.append('contact_phone', getDialCode(contactCC.value) + form.value.contact_phone)
  formData.append('target_phone', getDialCode(targetCC.value) + form.value.target_phone)
  formData.append('target_phone_local', form.value.target_phone)
  formData.append('reason', form.value.reason)
  for (const file of form.value.evidenceFiles) {
    formData.append('evidence_files[]', file)
  }
  try {
    const res = await fetch('/api/v1/appeal', { method: 'POST', body: formData })
    if (res.ok) {
      step.value = 4
    } else {
      errorMsg.value = t('appeal.alert_submit_failed')
    }
  } catch (e) {
    errorMsg.value = t('appeal.alert_network_error')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.legal-notice {
  background: rgba(255, 255, 255, 0.02);
  border-left: 3px solid #8b949e;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 2rem;
}
.legal-title {
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 0.5rem;
  font-size: 0.85rem;
}
.legal-notice p {
  color: var(--text-secondary);
  font-size: 0.8rem;
  line-height: 1.5;
}

/* Progress Tracker */
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
  color: #facc15;
  text-shadow: 0 0 8px rgba(250, 204, 21, 0.4);
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
  color: #facc15;
  border-color: rgba(250, 204, 21, 0.3);
}
.upload-hint {
  color: var(--text-secondary);
  font-size: 0.95rem;
  font-weight: 500;
  letter-spacing: 0.5px;
}

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
  color: #facc15;
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

.submit-btn {
  width: 100%;
  margin-top: 1rem;
  background: var(--silicon-bg);
  color: #facc15;
  border: 1px solid rgba(250, 204, 21, 0.2);
}
.submit-btn:hover {
  text-shadow: 0 0 8px #facc15;
  border-color: #facc15;
  box-shadow: var(--shadow-out), 0 0 15px rgba(250, 204, 21, 0.3);
}

.action-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 2rem;
}
.next-btn {
  min-width: 150px;
  background: var(--silicon-bg);
  color: #facc15;
  border-color: rgba(250, 204, 21, 0.2);
}
.next-btn:hover {
  text-shadow: 0 0 8px #facc15;
  border-color: #facc15;
  box-shadow: var(--shadow-out), 0 0 15px rgba(250, 204, 21, 0.3);
}
.btn-text {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 0.95rem;
  cursor: pointer;
}
.btn-text:hover {
  color: var(--text-primary);
}
.confirm-box {
  background: rgba(250, 204, 21, 0.05);
  padding: 24px;
  border-radius: var(--radius-lg);
  border: 1px dashed rgba(250, 204, 21, 0.3);
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

/* Success View */
.success-view {
  text-align: center;
  padding: 30px 10px;
}
.success-icon {
  font-size: 4rem;
  margin-bottom: 1.5rem;
}
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

@media (max-width: 480px) {
  .progress-bar { margin-bottom: 1.5rem; }
  .step { font-size: 0.75rem; }
  .line { margin: 0 6px; }
  .legal-notice { padding: 0.7rem; margin-bottom: 1.2rem; }
  .legal-title { font-size: 0.78rem; }
  .legal-notice p { font-size: 0.72rem; }
  .form-group { margin-bottom: 1rem; }
  .form-group label { font-size: 0.85rem; margin-bottom: 0.5rem; }
  .upload-area { padding: 16px; }
  .action-buttons { margin-top: 1.2rem; }
  .next-btn { min-width: 120px; }
  .confirm-box { padding: 16px; }
  .success-view { padding: 20px 8px; }
  .success-icon { font-size: 3rem; }
}
</style>

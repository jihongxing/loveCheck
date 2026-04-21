<template>
  <div class="company-appeal-container glass-panel">
    <h2 class="section-title">{{ $t('companyAppeal.title') }}</h2>
    <p v-if="step !== 4" class="section-subtitle">{{ $t('companyAppeal.subtitle') }}</p>

    <div class="legal-notice">
      <div class="legal-title">{{ $t('companyAppeal.legalTitle') }}</div>
      <p>{{ $t('companyAppeal.legalText') }}</p>
    </div>

    <p v-if="errorMsg && step < 4" class="error-msg" @click="errorMsg = ''">{{ errorMsg }}</p>

    <div v-if="step < 4" class="progress-bar">
      <div :class="['step', { active: step >= 1 }]">{{ $t('companyAppeal.step1') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 2 }]">{{ $t('companyAppeal.step2') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 3 }]">{{ $t('companyAppeal.step3') }}</div>
    </div>

    <div v-if="step === 1" class="step-content">
      <div class="form-group">
        <label>{{ $t('companyAppeal.contactPhone') }} <span class="required">*</span></label>
        <PhoneInput
          v-model="form.contactPhone"
          v-model:countryCode="contactCC"
          @validity="contactPhoneValid = $event"
          :placeholder="$t('companyAppeal.contactPhonePlaceholder')"
        />
        <small class="hint">{{ $t('companyAppeal.contactPhoneHint') }}</small>
      </div>

      <div class="action-buttons single-end">
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!contactPhoneValid">
          {{ $t('companyAppeal.next') }}
        </button>
      </div>
    </div>

    <div v-if="step === 2" class="step-content">
      <div class="form-group">
        <label>{{ $t('companyAppeal.registrationNo') }} <span class="required">*</span></label>
        <input
          v-model.trim="form.registrationNo"
          type="text"
          class="input-premium"
          :placeholder="$t('companyAppeal.registrationNoPlaceholder')"
        />
        <small class="hint">{{ $t('companyAppeal.registrationNoHint') }}</small>
      </div>

      <div class="form-group">
        <label>{{ $t('companyAppeal.reason') }} <span class="required">*</span></label>
        <textarea
          v-model.trim="form.reason"
          class="input-premium textarea-premium"
          rows="6"
          maxlength="5000"
          :placeholder="$t('companyAppeal.reasonPlaceholder')"
        ></textarea>
        <small class="hint">{{ form.reason.length }}/5000</small>
      </div>

      <div class="form-group">
        <label>{{ $t('companyAppeal.evidence') }}</label>
        <div class="evidence-grid" v-if="selectedFiles.length > 0">
          <div class="evidence-item" v-for="(file, index) in selectedFiles" :key="index">
            <span class="file-name">{{ file.name }}</span>
            <button type="button" class="remove-btn" @click="removeFile(index)">&times;</button>
          </div>
        </div>

        <div class="upload-area" @click="triggerFileInput" v-if="selectedFiles.length < 9">
          <span class="upload-text">{{ $t('companyAppeal.evidenceHint') }}</span>
        </div>
        <input
          ref="fileInput"
          type="file"
          multiple
          accept="image/*,application/pdf"
          @change="handleFileUpload"
          style="display: none;"
        />
      </div>

      <div class="action-buttons">
        <button class="btn-secondary" @click="prevStep">{{ $t('companyAppeal.back') }}</button>
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!form.registrationNo || !form.reason">
          {{ $t('companyAppeal.toConfirm') }}
        </button>
      </div>
    </div>

    <div v-if="step === 3" class="step-content">
      <div class="confirm-box">
        <h3 class="confirm-title">{{ $t('companyAppeal.confirmTitle') }}</h3>
        <p class="confirm-copy">{{ $t('companyAppeal.confirmDesc') }}</p>
        <p class="confirm-copy">{{ $t('companyAppeal.confirmPledge') }}</p>
      </div>

      <div class="action-buttons">
        <button class="btn-secondary" @click="prevStep" :disabled="isSubmitting">{{ $t('companyAppeal.back') }}</button>
        <button class="btn-premium submit-btn" @click="submitAppeal" :disabled="isSubmitting">
          {{ isSubmitting ? $t('companyAppeal.submitting') : $t('companyAppeal.submit') }}
        </button>
      </div>
    </div>

    <div v-if="step === 4" class="success-screen">
      <div class="success-icon">✓</div>
      <h3 class="success-title">{{ $t('companyAppeal.successTitle') }}</h3>
      <p class="success-message">{{ $t('companyAppeal.successMessage') }}</p>
      <button class="btn-premium" @click="resetForm">{{ $t('companyAppeal.submitAnother') }}</button>
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
const errorMsg = ref('')
const isSubmitting = ref(false)
const fileInput = ref(null)
const contactCC = ref(getDefaultCountry(locale.value))
const contactPhoneValid = ref(false)

const form = ref({
  contactPhone: '',
  registrationNo: '',
  reason: ''
})

const selectedFiles = ref([])

const triggerFileInput = () => {
  fileInput.value?.click()
}

const nextStep = () => {
  errorMsg.value = ''

  if (step.value === 1 && !contactPhoneValid.value) {
    errorMsg.value = t('companyAppeal.contactPhoneRequired')
    return
  }

  if (step.value === 2 && (!form.value.registrationNo || !form.value.reason)) {
    errorMsg.value = t('companyAppeal.fillRequired')
    return
  }

  step.value += 1
}

const prevStep = () => {
  errorMsg.value = ''
  step.value -= 1
}

const handleFileUpload = (event) => {
  const files = Array.from(event.target.files || [])
  if (selectedFiles.value.length + files.length > 9) {
    errorMsg.value = t('companyAppeal.tooManyFiles')
    event.target.value = ''
    return
  }

  selectedFiles.value.push(...files)
  event.target.value = ''
}

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1)
}

const resetForm = () => {
  step.value = 1
  errorMsg.value = ''
  isSubmitting.value = false
  contactPhoneValid.value = false
  contactCC.value = getDefaultCountry(locale.value)
  form.value = {
    contactPhone: '',
    registrationNo: '',
    reason: ''
  }
  selectedFiles.value = []
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const submitAppeal = async () => {
  isSubmitting.value = true
  errorMsg.value = ''

  const formData = new FormData()
  formData.append('contact_phone', getDialCode(contactCC.value) + form.value.contactPhone)
  formData.append('registration_no', form.value.registrationNo)
  formData.append('reason', form.value.reason)
  selectedFiles.value.forEach((file) => {
    formData.append('evidence_files[]', file)
  })

  try {
    const response = await fetch('/api/v1/company/appeal', {
      method: 'POST',
      body: formData
    })

    const data = await response.json().catch(() => ({}))
    if (response.ok) {
      step.value = 4
      return
    }

    errorMsg.value = data.error || t('companyAppeal.submitError')
  } catch {
    errorMsg.value = t('companyAppeal.networkError')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.company-appeal-container {
  max-width: 800px;
  margin: 0 auto;
}

.section-title {
  margin-bottom: 1rem;
  font-weight: 600;
  color: #facc15;
}

.section-subtitle {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

.legal-notice {
  background: rgba(255, 255, 255, 0.02);
  border-left: 3px solid #facc15;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
}

.legal-title {
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.legal-notice p,
.hint,
.confirm-copy,
.success-message {
  color: var(--text-secondary);
}

.progress-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
  padding: 0 1rem;
}

.step {
  flex: 1;
  text-align: center;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 600;
}

.step.active {
  background: rgba(250, 204, 21, 0.1);
  border-color: rgba(250, 204, 21, 0.5);
  color: #facc15;
  box-shadow: 0 0 15px rgba(250, 204, 21, 0.15);
}

.line {
  width: 40px;
  height: 2px;
  background: rgba(255, 255, 255, 0.1);
  margin: 0 0.5rem;
}

.step-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  color: var(--text-primary);
  font-size: 0.95rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.required {
  color: var(--danger-neon);
}

.hint {
  display: block;
  margin-top: 0.5rem;
  font-size: 0.85rem;
}

.error-msg {
  padding: 0.75rem;
  background: rgba(255, 0, 85, 0.1);
  border: 1px solid var(--danger-neon);
  border-radius: var(--radius-md);
  color: var(--danger-neon);
  font-size: 0.9rem;
  margin-bottom: 1.5rem;
  cursor: pointer;
}

.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.evidence-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
}

.file-name {
  flex: 1;
  color: var(--text-secondary);
  font-size: 0.85rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 0.5rem;
}

.remove-btn {
  background: rgba(255, 0, 85, 0.2);
  border: 1px solid var(--danger-neon);
  color: var(--danger-neon);
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.1rem;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.upload-area {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: rgba(255, 255, 255, 0.03);
  border: 2px dashed rgba(255, 255, 255, 0.2);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.upload-area:hover {
  border-color: rgba(250, 204, 21, 0.45);
  background: rgba(255, 255, 255, 0.05);
}

.upload-text {
  color: var(--text-secondary);
  text-align: center;
}

.confirm-box {
  background: rgba(250, 204, 21, 0.05);
  padding: 1.5rem;
  border-radius: var(--radius-lg);
  border: 1px dashed rgba(250, 204, 21, 0.35);
  text-align: center;
}

.confirm-title,
.success-title {
  color: #facc15;
}

.confirm-copy + .confirm-copy {
  margin-top: 1rem;
}

.action-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  justify-content: space-between;
}

.single-end {
  justify-content: flex-end;
}

.btn-secondary {
  padding: 0.75rem 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 1rem;
  font-weight: 600;
}

.next-btn,
.submit-btn {
  flex: 1;
}

.success-screen {
  text-align: center;
  padding: 3rem 2rem;
}

.success-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  background: radial-gradient(circle, rgba(250, 204, 21, 0.2) 0%, transparent 70%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 4rem;
  color: #facc15;
}

@media (max-width: 768px) {
  .progress-bar {
    padding: 0;
  }

  .progress-bar .step {
    font-size: 0.75rem;
    padding: 0.5rem;
  }

  .line {
    width: 20px;
    margin: 0 0.25rem;
  }

  .action-buttons {
    flex-direction: column;
  }

  .btn-secondary {
    order: 2;
  }

  .next-btn,
  .submit-btn {
    order: 1;
  }
}
</style>

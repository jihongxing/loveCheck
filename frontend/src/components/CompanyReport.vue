<template>
  <div class="company-report-container glass-panel">
    <h2 style="margin-bottom: 1rem; font-weight: 600;">{{ $t('companyReport.title') }}</h2>
    <p v-if="step !== 4" style="color: var(--text-secondary); margin-bottom: 2rem; font-size: 0.95rem;">
      {{ $t('companyReport.subtitle') }}
    </p>

    <p v-if="errorMsg && step < 4" class="error-msg" @click="errorMsg = ''">{{ errorMsg }}</p>

    <!-- Progress Tracker -->
    <div v-if="step < 4" class="progress-bar">
      <div :class="['step', { active: step >= 1 }]">{{ $t('companyReport.step1') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 2 }]">{{ $t('companyReport.step2') }}</div>
      <div class="line"></div>
      <div :class="['step', { active: step >= 3 }]">{{ $t('companyReport.step3') }}</div>
    </div>

    <!-- Step 1: Reporter Info -->
    <div v-if="step === 1" class="step-content">
      <div class="form-group">
        <label>{{ $t('companyReport.reporterPhone') }} <span class="required">*</span></label>
        <PhoneInput
          v-model="form.reporterPhone"
          v-model:countryCode="reporterCC"
          @validity="reporterPhoneValid = $event"
          :placeholder="$t('companyReport.reporterPhonePlaceholder')"
        />
        <small class="hint">{{ $t('companyReport.reporterPhoneHint') }}</small>
      </div>

      <div class="action-buttons" style="justify-content: flex-end;">
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!reporterPhoneValid">
          {{ $t('companyReport.next') }}
        </button>
      </div>
    </div>

    <!-- Step 2: Company Info -->
    <div v-if="step === 2" class="step-content">
      <div class="form-group">
        <label>{{ $t('companyReport.registrationNo') }} <span class="required">*</span></label>
        <input
          v-model="form.registrationNo"
          type="text"
          class="input-premium"
          :placeholder="$t('companyReport.registrationNoPlaceholder')"
        />
        <small class="hint">{{ $t('companyReport.registrationNoHint') }}</small>
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.companyName') }}</label>
        <input
          v-model="form.companyName"
          type="text"
          class="input-premium"
          :placeholder="$t('companyReport.companyNamePlaceholder')"
        />
        <small class="hint">{{ $t('companyReport.companyNameHint') }}</small>
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.industry') }}</label>
        <select v-model="form.industry" class="input-premium">
          <option value="">{{ $t('companyReport.selectIndustry') }}</option>
          <option value="互联网/IT">{{ $t('companyReport.industries.tech') }}</option>
          <option value="金融">{{ $t('companyReport.industries.finance') }}</option>
          <option value="教育培训">{{ $t('companyReport.industries.education') }}</option>
          <option value="房地产">{{ $t('companyReport.industries.realEstate') }}</option>
          <option value="制造业">{{ $t('companyReport.industries.manufacturing') }}</option>
          <option value="零售/电商">{{ $t('companyReport.industries.retail') }}</option>
          <option value="餐饮/服务">{{ $t('companyReport.industries.hospitality') }}</option>
          <option value="医疗健康">{{ $t('companyReport.industries.healthcare') }}</option>
          <option value="其他">{{ $t('companyReport.industries.other') }}</option>
        </select>
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.locationCity') }}</label>
        <input
          v-model="form.locationCity"
          type="text"
          class="input-premium"
          :placeholder="$t('companyReport.locationCityPlaceholder')"
        />
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.employmentPeriod') }}</label>
        <input
          v-model="form.employmentPeriod"
          type="text"
          class="input-premium"
          :placeholder="$t('companyReport.employmentPeriodPlaceholder')"
        />
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.position') }}</label>
        <input
          v-model="form.position"
          type="text"
          class="input-premium"
          :placeholder="$t('companyReport.positionPlaceholder')"
        />
      </div>

      <div class="action-buttons">
        <button class="btn-secondary" @click="prevStep">{{ $t('companyReport.back') }}</button>
        <button class="btn-premium next-btn" @click="nextStep" :disabled="!form.registrationNo">
          {{ $t('companyReport.next') }}
        </button>
      </div>
    </div>

    <!-- Step 3: Details & Evidence -->
    <div v-if="step === 3" class="step-content">
      <div class="form-group">
        <label>{{ $t('companyReport.tags') }} <span class="required">*</span></label>
        <div class="tags-selector">
          <button
            type="button"
            v-for="tag in availableTags"
            :key="tag.value"
            :class="['tag-btn', { selected: selectedTags.includes(tag.value) }]"
            @click="toggleTag(tag.value)"
          >
            {{ tag.label }}
          </button>
        </div>
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.description') }} <span class="required">*</span></label>
        <textarea
          v-model="form.description"
          class="input-premium textarea-premium"
          :placeholder="$t('companyReport.descriptionPlaceholder')"
          rows="6"
          maxlength="5000"
        ></textarea>
        <small class="hint">{{ form.description.length }}/5000</small>
      </div>

      <div class="form-group">
        <label>{{ $t('companyReport.evidence') }}</label>

        <div class="evidence-grid" v-if="selectedFiles.length > 0">
          <div class="evidence-item" v-for="(file, index) in selectedFiles" :key="index">
            <span class="file-name">{{ file.name }}</span>
            <button type="button" class="remove-btn" @click="removeFile(index)">&times;</button>
          </div>
        </div>

        <div class="upload-area" @click="triggerFileInput" v-if="selectedFiles.length < 9">
          <span class="upload-icon">📎</span>
          <span class="upload-text">{{ $t('companyReport.evidenceHint') }}</span>
        </div>
        <input
          type="file"
          multiple
          accept="image/*,application/pdf"
          @change="handleFileUpload"
          ref="fileInput"
          style="display: none;"
        />
      </div>

      <div class="action-buttons">
        <button class="btn-secondary" @click="prevStep">{{ $t('companyReport.back') }}</button>
        <button class="btn-premium submit-btn" @click="submitReport" :disabled="isSubmitting || !canSubmit">
          {{ isSubmitting ? $t('companyReport.submitting') : $t('companyReport.submit') }}
        </button>
      </div>
    </div>

    <!-- Step 4: Success -->
    <div v-if="step === 4" class="success-screen">
      <div class="success-icon">✓</div>
      <h3 style="color: var(--safe-color); margin-bottom: 1rem;">{{ $t('companyReport.successTitle') }}</h3>
      <p style="color: var(--text-secondary); margin-bottom: 2rem;">{{ $t('companyReport.successMessage') }}</p>
      <button class="btn-premium" @click="resetForm">{{ $t('companyReport.submitAnother') }}</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PhoneInput from './PhoneInput.vue'
import { getDialCode } from '../data/countryCodes.js'

const { t, locale } = useI18n()

// Language to country code mapping
const langToCountry = {
  zh: 'CN',
  en: 'US',
  ja: 'JP',
  ko: 'KR',
  th: 'TH',
  vi: 'VN',
  es: 'ES',
  hi: 'IN'
}

const step = ref(1)
const errorMsg = ref('')
const isSubmitting = ref(false)

const form = ref({
  reporterPhone: '',
  companyName: '',
  registrationNo: '',
  industry: '',
  locationCity: '',
  employmentPeriod: '',
  position: '',
  description: ''
})

const reporterCC = ref('CN')
const reporterPhoneValid = ref(false)

// Initialize country code based on current locale
onMounted(() => {
  reporterCC.value = langToCountry[locale.value] || 'CN'
})

// Watch locale changes and update country code
watch(locale, (newLocale) => {
  reporterCC.value = langToCountry[newLocale] || 'CN'
})

const selectedTags = ref([])
const selectedFiles = ref([])
const fileInput = ref(null)

const availableTags = computed(() => [
  { value: '拖欠工资', label: t('companyReport.tagOptions.unpaidWages') },
  { value: '强制加班', label: t('companyReport.tagOptions.forcedOvertime') },
  { value: '无偿加班', label: t('companyReport.tagOptions.unpaidOvertime') },
  { value: 'PUA管理', label: t('companyReport.tagOptions.toxicManagement') },
  { value: '违法裁员', label: t('companyReport.tagOptions.illegalLayoff') },
  { value: '不缴社保', label: t('companyReport.tagOptions.noInsurance') },
  { value: '职场霸凌', label: t('companyReport.tagOptions.workplaceBullying') },
  { value: '性骚扰', label: t('companyReport.tagOptions.sexualHarassment') },
  { value: '虚假招聘', label: t('companyReport.tagOptions.fraudulentHiring') },
  { value: '克扣奖金', label: t('companyReport.tagOptions.withheldBonus') },
  { value: '劳动合同欺诈', label: t('companyReport.tagOptions.contractFraud') },
  { value: '其他违法行为', label: t('companyReport.tagOptions.other') }
])

const canSubmit = computed(() => {
  return form.value.reporterPhone &&
         form.value.registrationNo &&
         selectedTags.value.length > 0 &&
         form.value.description.trim().length > 0
})

const nextStep = () => {
  if (step.value === 1 && !form.value.reporterPhone) {
    errorMsg.value = t('companyReport.reporterPhoneRequired')
    return
  }
  if (step.value === 2 && !form.value.registrationNo) {
    errorMsg.value = t('companyReport.registrationNoRequired')
    return
  }
  errorMsg.value = ''
  step.value++
}

const prevStep = () => {
  errorMsg.value = ''
  step.value--
}

const toggleTag = (tag) => {
  const index = selectedTags.value.indexOf(tag)
  if (index > -1) {
    selectedTags.value.splice(index, 1)
  } else {
    selectedTags.value.push(tag)
  }
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const handleFileUpload = (event) => {
  const files = Array.from(event.target.files)
  if (selectedFiles.value.length + files.length > 9) {
    errorMsg.value = t('companyReport.tooManyFiles')
    return
  }
  selectedFiles.value.push(...files)
}

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1)
}

const submitReport = async () => {
  if (!canSubmit.value) {
    errorMsg.value = t('companyReport.fillRequired')
    return
  }

  if (selectedTags.value.length === 0) {
    errorMsg.value = t('companyReport.selectTagsError')
    return
  }

  isSubmitting.value = true
  errorMsg.value = ''

  const formData = new FormData()
  formData.append('reporter_phone', getDialCode(reporterCC.value) + form.value.reporterPhone)
  formData.append('company_name', form.value.companyName)
  formData.append('registration_no', form.value.registrationNo)
  formData.append('industry', form.value.industry)
  formData.append('location_city', form.value.locationCity)
  formData.append('employment_period', form.value.employmentPeriod)
  formData.append('position', form.value.position)
  formData.append('tags', JSON.stringify(selectedTags.value))
  formData.append('description', form.value.description)

  selectedFiles.value.forEach(file => {
    formData.append('evidence_files[]', file)
  })

  try {
    const response = await fetch('/api/v1/company/report', {
      method: 'POST',
      body: formData
    })

    const data = await response.json()

    if (response.ok) {
      step.value = 4
    } else {
      errorMsg.value = data.error || t('companyReport.submitError')
    }
  } catch (error) {
    errorMsg.value = t('companyReport.networkError')
  } finally {
    isSubmitting.value = false
  }
}

const resetForm = () => {
  step.value = 1
  form.value = {
    reporterPhone: '',
    companyName: '',
    registrationNo: '',
    industry: '',
    locationCity: '',
    employmentPeriod: '',
    position: '',
    description: ''
  }
  selectedTags.value = []
  selectedFiles.value = []
  errorMsg.value = ''
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}
</script>

<style scoped>
.company-report-container {
  max-width: 800px;
  margin: 0 auto;
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
  transition: all 0.2s;
}

.error-msg:hover {
  background: rgba(255, 0, 85, 0.15);
}

/* Progress Bar */
.progress-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2.5rem;
  padding: 0 1rem;
}

.progress-bar .step {
  flex: 1;
  text-align: center;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 600;
  transition: all 0.3s;
}

.progress-bar .step.active {
  background: rgba(0, 240, 255, 0.1);
  border-color: var(--accent-neon);
  color: var(--accent-neon);
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.2);
}

.progress-bar .line {
  width: 40px;
  height: 2px;
  background: rgba(255, 255, 255, 0.1);
  margin: 0 0.5rem;
}

/* Step Content */
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
  color: var(--text-secondary);
  font-size: 0.85rem;
}

select.input-premium {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%2300f0ff' d='M6 9L1 4h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 1rem center;
  padding-right: 2.5rem;
}

/* Tags Selector */
.tags-selector {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 0.75rem;
}

.tag-btn {
  padding: 0.75rem 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.tag-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.tag-btn.selected {
  background: rgba(0, 240, 255, 0.15);
  border-color: var(--accent-neon);
  color: var(--accent-neon);
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
}

/* Evidence Upload */
.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
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
  font-size: 1.2rem;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.remove-btn:hover {
  background: rgba(255, 0, 85, 0.3);
  box-shadow: 0 0 10px rgba(255, 0, 85, 0.3);
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.03);
  border: 2px dashed rgba(255, 255, 255, 0.2);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.upload-area:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: var(--accent-neon);
}

.upload-icon {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.upload-text {
  color: var(--text-secondary);
  font-size: 0.9rem;
  text-align: center;
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  justify-content: space-between;
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
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}

.next-btn,
.submit-btn {
  flex: 1;
}

.submit-btn:disabled,
.next-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Success Screen */
.success-screen {
  text-align: center;
  padding: 3rem 2rem;
  animation: fadeIn 0.5s ease;
}

.success-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  background: radial-gradient(circle, rgba(76, 175, 80, 0.2) 0%, transparent 70%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 4rem;
  color: var(--safe-color);
  text-shadow: 0 0 20px rgba(76, 175, 80, 0.5);
  animation: scaleIn 0.5s ease;
}

@keyframes scaleIn {
  from { transform: scale(0); }
  to { transform: scale(1); }
}

@media (max-width: 768px) {
  .progress-bar {
    padding: 0;
  }

  .progress-bar .step {
    font-size: 0.75rem;
    padding: 0.5rem;
  }

  .progress-bar .line {
    width: 20px;
    margin: 0 0.25rem;
  }

  .tags-selector {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.5rem;
  }

  .tag-btn {
    padding: 0.6rem 0.75rem;
    font-size: 0.85rem;
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

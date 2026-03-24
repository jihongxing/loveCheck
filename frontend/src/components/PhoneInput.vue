<template>
  <div class="phone-input-wrapper">
    <div class="country-select" @click="showDropdown = !showDropdown" ref="selectRef">
      <span class="country-flag">{{ selected.flag }}</span>
      <span class="country-dial">{{ selected.dialCode }}</span>
      <span class="dropdown-arrow">&#9662;</span>
    </div>
    <input
      :value="modelValue"
      @input="onInput"
      type="tel"
      class="input-premium phone-number-input"
      :placeholder="placeholder || $t('phone_input.placeholder')"
      :maxlength="selected.maxLen"
    />
    <transition name="dropdown-fade">
      <div v-if="showDropdown" class="country-dropdown" ref="dropdownRef">
        <input
          v-model="search"
          type="text"
          class="country-search"
          :placeholder="$t('phone_input.search_country')"
          @click.stop
          ref="searchInputRef"
        />
        <div class="country-list">
          <div
            v-for="c in filteredCountries"
            :key="c.code + c.dialCode"
            class="country-option"
            :class="{ active: c.code === selected.code && c.dialCode === selected.dialCode }"
            @click="selectCountry(c)"
          >
            <span class="option-flag">{{ c.flag }}</span>
            <span class="option-name">{{ $t('country.' + c.code) }}</span>
            <span class="option-dial">{{ c.dialCode }}</span>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import countryCodes from '../data/countryCodes.js'

const { t, locale } = useI18n()

const props = defineProps({
  modelValue: { type: String, default: '' },
  countryCode: { type: String, default: 'CN' },
  placeholder: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue', 'update:countryCode', 'validity'])

const showDropdown = ref(false)
const search = ref('')
const selectRef = ref(null)
const dropdownRef = ref(null)
const searchInputRef = ref(null)

const selected = computed(() =>
  countryCodes.find(c => c.code === props.countryCode) || countryCodes[0]
)

const filteredCountries = computed(() => {
  if (!search.value) return countryCodes
  const q = search.value.toLowerCase()
  return countryCodes.filter(c =>
    c.dialCode.includes(q) ||
    c.code.toLowerCase().includes(q) ||
    t('country.' + c.code).toLowerCase().includes(q)
  )
})

const isValid = computed(() => {
  const len = props.modelValue.replace(/\D/g, '').length
  return len >= selected.value.minLen
})

watch(isValid, (v) => emit('validity', v), { immediate: true })
watch(() => props.modelValue, () => emit('validity', isValid.value))

const onInput = (e) => {
  const digits = e.target.value.replace(/\D/g, '')
  emit('update:modelValue', digits)
}

const selectCountry = (c) => {
  emit('update:countryCode', c.code)
  emit('update:modelValue', '')
  showDropdown.value = false
  search.value = ''
}

watch(showDropdown, async (v) => {
  if (v) {
    await nextTick()
    searchInputRef.value?.focus()
  }
})

const onClickOutside = (e) => {
  if (
    showDropdown.value &&
    selectRef.value && !selectRef.value.contains(e.target) &&
    dropdownRef.value && !dropdownRef.value.contains(e.target)
  ) {
    showDropdown.value = false
    search.value = ''
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.phone-input-wrapper {
  display: flex;
  align-items: stretch;
  position: relative;
  gap: 0;
}

.country-select {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--silicon-bg);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-right: none;
  border-radius: var(--radius-md) 0 0 var(--radius-md);
  padding: 0 10px;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
  white-space: nowrap;
  min-width: 95px;
}
.country-select:hover {
  border-color: rgba(0, 240, 255, 0.3);
}
.country-flag { font-size: 1.1rem; }
.country-dial {
  color: var(--text-primary);
  font-size: 0.85rem;
  font-weight: 500;
}
.dropdown-arrow {
  color: var(--text-secondary);
  font-size: 0.65rem;
  margin-left: 2px;
}

.phone-number-input {
  border-radius: 0 var(--radius-md) var(--radius-md) 0 !important;
  flex: 1;
  min-width: 0;
}

/* Dropdown */
.country-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 100%;
  max-height: 280px;
  background: var(--silicon-surface);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.country-search {
  background: var(--silicon-bg);
  border: none;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
  padding: 10px 12px;
  font-size: 0.85rem;
  outline: none;
}
.country-search::placeholder { color: var(--text-secondary); }

.country-list {
  overflow-y: auto;
  flex: 1;
}
.country-list::-webkit-scrollbar { width: 4px; }
.country-list::-webkit-scrollbar-thumb {
  background: rgba(0, 240, 255, 0.2);
  border-radius: 4px;
}

.country-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 0.85rem;
}
.country-option:hover { background: rgba(255, 255, 255, 0.05); }
.country-option.active {
  background: rgba(0, 240, 255, 0.08);
  color: var(--accent-neon);
}
.option-flag { font-size: 1rem; }
.option-name {
  flex: 1;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.country-option.active .option-name { color: var(--accent-neon); }
.option-dial {
  color: var(--text-secondary);
  font-size: 0.8rem;
  font-family: monospace;
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}
.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@media (max-width: 480px) {
  .country-select { min-width: 80px; padding: 0 6px; }
  .country-flag { font-size: 1rem; }
  .country-dial { font-size: 0.78rem; }
  .dropdown-arrow { font-size: 0.55rem; }
  .country-dropdown { max-height: 240px; }
  .country-search { padding: 8px 10px; font-size: 0.8rem; }
  .country-option { padding: 7px 10px; font-size: 0.8rem; }
}
</style>

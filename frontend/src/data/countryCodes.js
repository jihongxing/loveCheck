// Country dial codes for the phone input selector
// Ordered by relevance to LoverTrust's supported locales, then alphabetically

const localeToCountry = {
  zh: 'CN', en: 'US', ja: 'JP', ko: 'KR',
  th: 'TH', vi: 'VN', es: 'ES', hi: 'IN',
}

export function getDefaultCountry(locale) {
  return localeToCountry[locale] || 'CN'
}

export function getDialCode(countryIso) {
  const entry = countryCodes.find(c => c.code === countryIso)
  return entry ? entry.dialCode : '+86'
}

const countryCodes = [
  { code: 'CN', dialCode: '+86',  flag: '🇨🇳', minLen: 11, maxLen: 11 },
  { code: 'US', dialCode: '+1',   flag: '🇺🇸', minLen: 10, maxLen: 10 },
  { code: 'JP', dialCode: '+81',  flag: '🇯🇵', minLen: 10, maxLen: 11 },
  { code: 'KR', dialCode: '+82',  flag: '🇰🇷', minLen: 10, maxLen: 11 },
  { code: 'TH', dialCode: '+66',  flag: '🇹🇭', minLen: 9,  maxLen: 10 },
  { code: 'VN', dialCode: '+84',  flag: '🇻🇳', minLen: 9,  maxLen: 10 },
  { code: 'ES', dialCode: '+34',  flag: '🇪🇸', minLen: 9,  maxLen: 9  },
  { code: 'IN', dialCode: '+91',  flag: '🇮🇳', minLen: 10, maxLen: 10 },
  { code: 'GB', dialCode: '+44',  flag: '🇬🇧', minLen: 10, maxLen: 11 },
  { code: 'TW', dialCode: '+886', flag: '🇹🇼', minLen: 9,  maxLen: 10 },
  { code: 'HK', dialCode: '+852', flag: '🇭🇰', minLen: 8,  maxLen: 8  },
  { code: 'MO', dialCode: '+853', flag: '🇲🇴', minLen: 8,  maxLen: 8  },
  { code: 'SG', dialCode: '+65',  flag: '🇸🇬', minLen: 8,  maxLen: 8  },
  { code: 'MY', dialCode: '+60',  flag: '🇲🇾', minLen: 9,  maxLen: 11 },
  { code: 'ID', dialCode: '+62',  flag: '🇮🇩', minLen: 10, maxLen: 12 },
  { code: 'PH', dialCode: '+63',  flag: '🇵🇭', minLen: 10, maxLen: 10 },
  { code: 'AU', dialCode: '+61',  flag: '🇦🇺', minLen: 9,  maxLen: 9  },
  { code: 'CA', dialCode: '+1',   flag: '🇨🇦', minLen: 10, maxLen: 10 },
  { code: 'DE', dialCode: '+49',  flag: '🇩🇪', minLen: 10, maxLen: 12 },
  { code: 'FR', dialCode: '+33',  flag: '🇫🇷', minLen: 9,  maxLen: 9  },
  { code: 'MX', dialCode: '+52',  flag: '🇲🇽', minLen: 10, maxLen: 10 },
  { code: 'BR', dialCode: '+55',  flag: '🇧🇷', minLen: 10, maxLen: 11 },
  { code: 'RU', dialCode: '+7',   flag: '🇷🇺', minLen: 10, maxLen: 10 },
  { code: 'AE', dialCode: '+971', flag: '🇦🇪', minLen: 9,  maxLen: 9  },
  { code: 'SA', dialCode: '+966', flag: '🇸🇦', minLen: 9,  maxLen: 9  },
  { code: 'NZ', dialCode: '+64',  flag: '🇳🇿', minLen: 8,  maxLen: 10 },
  { code: 'IT', dialCode: '+39',  flag: '🇮🇹', minLen: 9,  maxLen: 11 },
  { code: 'PT', dialCode: '+351', flag: '🇵🇹', minLen: 9,  maxLen: 9  },
  { code: 'AR', dialCode: '+54',  flag: '🇦🇷', minLen: 10, maxLen: 10 },
  { code: 'CL', dialCode: '+56',  flag: '🇨🇱', minLen: 9,  maxLen: 9  },
  { code: 'CO', dialCode: '+57',  flag: '🇨🇴', minLen: 10, maxLen: 10 },
  { code: 'MM', dialCode: '+95',  flag: '🇲🇲', minLen: 8,  maxLen: 10 },
  { code: 'KH', dialCode: '+855', flag: '🇰🇭', minLen: 8,  maxLen: 9  },
  { code: 'LA', dialCode: '+856', flag: '🇱🇦', minLen: 8,  maxLen: 10 },
  { code: 'BD', dialCode: '+880', flag: '🇧🇩', minLen: 10, maxLen: 10 },
  { code: 'PK', dialCode: '+92',  flag: '🇵🇰', minLen: 10, maxLen: 10 },
  { code: 'LK', dialCode: '+94',  flag: '🇱🇰', minLen: 9,  maxLen: 9  },
  { code: 'NP', dialCode: '+977', flag: '🇳🇵', minLen: 10, maxLen: 10 },
]

export default countryCodes

# 公司举报功能 - 快速集成指南

## 🚀 5分钟快速上线

### 步骤1: 数据库自动迁移
```bash
# 启动服务，自动创建新表和索引
cd backend
go run cmd/lovecheck/main.go
```

### 步骤2: 前端路由集成
编辑 `frontend/src/App.vue`，添加路由：

```vue
<template>
  <div id="app">
    <nav>
      <button @click="currentView = 'Search'">查询个人</button>
      <button @click="currentView = 'Report'">举报个人</button>
      <button @click="currentView = 'CompanySearch'">查询公司</button>
      <button @click="currentView = 'CompanyReport'">举报公司</button>
      <button @click="currentView = 'Admin'">管理后台</button>
    </nav>

    <component :is="currentView"></component>
  </div>
</template>

<script>
import Search from './components/Search.vue'
import Report from './components/Report.vue'
import CompanySearch from './components/CompanySearch.vue'
import CompanyReport from './components/CompanyReport.vue'
import Admin from './components/Admin.vue'

export default {
  components: {
    Search,
    Report,
    CompanySearch,
    CompanyReport,
    Admin
  },
  data() {
    return {
      currentView: 'Search'
    }
  }
}
</script>
```

### 步骤3: 添加国际化文本
编辑 `frontend/src/i18n.js`（如果不存在则创建）：

```javascript
export default {
  zh: {
    companyReport: {
      title: '举报公司',
      subtitle: '曝光黑心企业，保护劳动者权益',
      reporterPhone: '您的手机号',
      reporterPhonePlaceholder: '用于身份验证，不会公开',
      reporterPhoneHint: '仅用于防止恶意举报，前台完全匿名',
      companyName: '公司名称',
      companyNamePlaceholder: '请输入完整公司名称',
      registrationNo: '工商注册号',
      registrationNoPlaceholder: '选填，提高匹配精度',
      registrationNoHint: '可在企查查/天眼查查询',
      industry: '所属行业',
      selectIndustry: '请选择行业',
      industries: {
        tech: '互联网/IT',
        finance: '金融',
        education: '教育培训',
        realEstate: '房地产',
        manufacturing: '制造业',
        retail: '零售/电商',
        hospitality: '餐饮/服务',
        healthcare: '医疗健康',
        other: '其他'
      },
      locationCity: '公司所在城市',
      locationCityPlaceholder: '如：北京',
      employmentPeriod: '就职时间段',
      employmentPeriodPlaceholder: '如：2023.01-2024.06',
      position: '职位',
      positionPlaceholder: '如：软件工程师（可选）',
      tags: '违法行为类型',
      tagOptions: {
        unpaidWages: '拖欠工资',
        forcedOvertime: '强制加班',
        unpaidOvertime: '无偿加班',
        toxicManagement: 'PUA管理',
        illegalLayoff: '违法裁员',
        noInsurance: '不缴社保',
        workplaceBullying: '职场霸凌',
        sexualHarassment: '性骚扰',
        fraudulentHiring: '虚假招聘',
        withheldBonus: '克扣奖金',
        contractFraud: '劳动合同欺诈',
        other: '其他违法行为'
      },
      description: '详细描述',
      descriptionPlaceholder: '请详细描述您的经历，包括时间、地点、具体事件...',
      evidence: '证据材料',
      evidenceHint: '支持图片和PDF，最多9个文件，每个最大10MB',
      submit: '提交举报',
      submitting: '提交中...',
      successMessage: '举报已提交！感谢您的勇敢发声',
      tooManyFiles: '最多只能上传9个文件',
      selectTagsError: '请至少选择一个违法行为类型',
      submitError: '提交失败，请稍后重试',
      networkError: '网络错误，请检查连接'
    },
    companySearch: {
      title: '查询公司',
      subtitle: '求职前先查一查，避开黑心企业',
      companyNamePlaceholder: '输入公司名称',
      registrationNoPlaceholder: '工商注册号（可选）',
      search: '查询',
      searching: '查询中...',
      searchingMessage: '正在搜索数据库...',
      cleanTitle: '暂无举报记录',
      companyProfile: '公司风险档案',
      companyName: '公司名称',
      totalReports: '举报总数',
      riskScore: '风险评分',
      riskLevel: '风险等级',
      riskHigh: '高风险',
      riskMedium: '中风险',
      riskLow: '低风险',
      locations: '涉及城市',
      tags: '违法行为',
      firstReport: '首次举报',
      latestReport: '最新举报',
      companyAppeal: '公司申诉',
      reporterVotes: '支持举报人',
      companyVotes: '支持公司',
      supportReporter: '支持举报人',
      supportCompany: '支持公司',
      individualReports: '详细举报记录',
      position: '职位',
      period: '就职时间',
      location: '地点',
      voteSuccess: '投票成功',
      voteError: '投票失败',
      companyNameRequired: '请输入公司名称',
      searchError: '查询失败',
      networkError: '网络错误'
    }
  },
  en: {
    companyReport: {
      title: 'Report Company',
      subtitle: 'Expose toxic employers, protect workers\' rights',
      reporterPhone: 'Your Phone Number',
      reporterPhonePlaceholder: 'For verification, will not be disclosed',
      reporterPhoneHint: 'Only for preventing malicious reports, fully anonymous',
      companyName: 'Company Name',
      companyNamePlaceholder: 'Enter full company name',
      registrationNo: 'Registration Number',
      registrationNoPlaceholder: 'Optional, improves accuracy',
      registrationNoHint: 'Can be found on business registry',
      industry: 'Industry',
      selectIndustry: 'Select industry',
      industries: {
        tech: 'Internet/IT',
        finance: 'Finance',
        education: 'Education',
        realEstate: 'Real Estate',
        manufacturing: 'Manufacturing',
        retail: 'Retail/E-commerce',
        hospitality: 'Hospitality',
        healthcare: 'Healthcare',
        other: 'Other'
      },
      locationCity: 'City',
      locationCityPlaceholder: 'e.g., Beijing',
      employmentPeriod: 'Employment Period',
      employmentPeriodPlaceholder: 'e.g., 2023.01-2024.06',
      position: 'Position',
      positionPlaceholder: 'e.g., Software Engineer (optional)',
      tags: 'Abuse Types',
      tagOptions: {
        unpaidWages: 'Unpaid Wages',
        forcedOvertime: 'Forced Overtime',
        unpaidOvertime: 'Unpaid Overtime',
        toxicManagement: 'Toxic Management',
        illegalLayoff: 'Illegal Layoff',
        noInsurance: 'No Insurance',
        workplaceBullying: 'Workplace Bullying',
        sexualHarassment: 'Sexual Harassment',
        fraudulentHiring: 'Fraudulent Hiring',
        withheldBonus: 'Withheld Bonus',
        contractFraud: 'Contract Fraud',
        other: 'Other Violations'
      },
      description: 'Detailed Description',
      descriptionPlaceholder: 'Please describe your experience in detail...',
      evidence: 'Evidence',
      evidenceHint: 'Images and PDF supported, max 9 files, 10MB each',
      submit: 'Submit Report',
      submitting: 'Submitting...',
      successMessage: 'Report submitted! Thank you for speaking up',
      tooManyFiles: 'Maximum 9 files allowed',
      selectTagsError: 'Please select at least one abuse type',
      submitError: 'Submission failed, please try again',
      networkError: 'Network error, please check connection'
    },
    companySearch: {
      title: 'Search Company',
      subtitle: 'Check before you apply, avoid toxic employers',
      companyNamePlaceholder: 'Enter company name',
      registrationNoPlaceholder: 'Registration number (optional)',
      search: 'Search',
      searching: 'Searching...',
      searchingMessage: 'Searching database...',
      cleanTitle: 'No Reports Found',
      companyProfile: 'Company Risk Profile',
      companyName: 'Company Name',
      totalReports: 'Total Reports',
      riskScore: 'Risk Score',
      riskLevel: 'Risk Level',
      riskHigh: 'High Risk',
      riskMedium: 'Medium Risk',
      riskLow: 'Low Risk',
      locations: 'Locations',
      tags: 'Violations',
      firstReport: 'First Report',
      latestReport: 'Latest Report',
      companyAppeal: 'Company Appeal',
      reporterVotes: 'Support Reporters',
      companyVotes: 'Support Company',
      supportReporter: 'Support Reporters',
      supportCompany: 'Support Company',
      individualReports: 'Individual Reports',
      position: 'Position',
      period: 'Period',
      location: 'Location',
      voteSuccess: 'Vote recorded',
      voteError: 'Vote failed',
      companyNameRequired: 'Company name required',
      searchError: 'Search failed',
      networkError: 'Network error'
    }
  }
}

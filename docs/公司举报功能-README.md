# 公司举报功能 - 快速开始

## 🎯 功能简介

在原有的**恋爱信用预警平台**基础上，新增**公司举报功能**，允许用户：
- 📝 举报存在劳动违法行为的公司
- 🔍 查询公司的举报记录和风险评分
- ⚖️ 公司可提交申诉，公众可投票裁决
- 📊 查看行业黑名单统计数据

---

## 📦 新增文件清单

### 后端 (Go)
```
backend/
├── internal/
│   ├── model/
│   │   └── company.go              # 数据模型（3个表）
│   └── handler/
│       └── company.go              # API处理器（5个接口）
├── internal/db/db.go               # 已更新：添加表迁移和索引
└── cmd/lovecheck/main.go           # 已更新：添加路由配置
```

### 前端 (Vue 3)
```
frontend/
└── src/
    ├── components/
    │   ├── CompanyReport.vue       # 公司举报表单
    │   └── CompanySearch.vue       # 公司查询界面
    └── i18n-company.js             # 国际化文本（中英文）
```

### 文档与脚本
```
docs/
├── 公司举报功能文档.md              # 完整技术文档
└── 公司举报功能-实施清单.md         # 实施清单

scripts/
├── test-company-api.sh             # API测试脚本（Linux/Mac）
└── test-company-api.bat            # API测试脚本（Windows）
```

---

## 🚀 快速启动（3步）

### 步骤1: 启动后端（自动创建数据库表）
```bash
cd backend
go run cmd/lovecheck/main.go
```

后端会自动：
- 创建3张新表：`company_records`, `company_stats`, `company_watchlists`
- 创建7个优化索引（Hash索引 + Partial索引 + GIN索引）

### 步骤2: 集成前端组件
编辑 `frontend/src/App.vue`，添加：

```vue
<script>
import CompanySearch from './components/CompanySearch.vue'
import CompanyReport from './components/CompanyReport.vue'

export default {
  components: {
    Search,
    Report,
    CompanySearch,    // 新增
    CompanyReport,    // 新增
    Admin
  }
}
</script>

<template>
  <nav>
    <button @click="currentView = 'Search'">查询个人</button>
    <button @click="currentView = 'Report'">举报个人</button>
    <button @click="currentView = 'CompanySearch'">查询公司</button>
    <button @click="currentView = 'CompanyReport'">举报公司</button>
  </nav>
  <component :is="currentView"></component>
</template>
```

### 步骤3: 测试功能
```bash
# Linux/Mac
chmod +x scripts/test-company-api.sh
./scripts/test-company-api.sh

# Windows
scripts\test-company-api.bat
```

---

## 📡 API接口一览

| 方法 | 路径 | 功能 | 限流 |
|------|------|------|------|
| POST | `/api/v1/company/report` | 提交公司举报 | 10次/分钟 |
| GET | `/api/v1/company/query` | 查询公司记录 | 60次/分钟 |
| POST | `/api/v1/company/appeal` | 公司申诉 | 10次/分钟 |
| POST | `/api/v1/company/vote` | 陪审团投票 | 20次/天 |
| GET | `/api/v1/company/stats` | 统计数据 | 60次/分钟 |

---

## 🗄️ 数据库设计

### company_records（公司举报记录表）
```sql
- company_hash: 公司名+注册号的HMAC哈希（零知识查询）
- company_name: 完整公司名
- display_name: 脱敏公司名（如：北京字**科技有限公司）
- registration_no: 脱敏注册号（如：9111****234567）
- industry: 行业分类
- tags: 违法行为标签（JSONB数组）
- evidence_mask_url: 证据文件URL（JSON数组）
- employment_period: 就职时间段
- position: 职位
- status: active/hidden/cleansed_by_jury
```

### company_stats（申诉与投票统计表）
```sql
- company_hash: 公司哈希（主键）
- appeal_reason: 申诉理由
- appeal_evidence: 申诉证据URL
- reporter_votes: 支持举报人的票数
- company_votes: 支持公司的票数
```

---

## 🔒 隐私与安全

### 零知识查询
- 公司名 + 注册号通过 `HMAC-SHA256` 加盐哈希存储
- 数据库仅存储哈希值，无法反向推导
- 查询时实时计算哈希进行匹配

### 举报人保护
- **后台实名**：手机号哈希存储，防止恶意举报
- **前台匿名**：显示随机昵称（如：举报者#A7B3）

### 防刷机制
- **Bloom Filter**：快速过滤不存在的公司
- **Redis限流**：IP + 设备指纹双重防护
- **投票去重**：30天TTL，防止刷票

---

## 🎨 前端功能

### CompanyReport.vue（举报表单）
- ✅ 12种违法行为标签（拖欠工资、强制加班、PUA管理等）
- ✅ 文件上传预览（最多9个，支持图片和PDF）
- ✅ 实时字数统计（最多5000字）
- ✅ 表单验证和错误提示

### CompanySearch.vue（查询界面）
- ✅ 风险评分可视化（0-100分）
- ✅ 风险等级标识（低/中/高）
- ✅ 举报记录时间轴
- ✅ 证据图片画廊
- ✅ 陪审团投票功能
- ✅ 申诉信息展示

---

## 📊 风险评分算法

### 评分维度
1. **举报数量**：1-2条(20分) → 3-5条(40分) → 6-10条(60分) → 10+条(80分)
2. **标签严重性**：高危(+15分) / 中危(+10分) / 低危(+5分)
3. **证据完整性**：有证据(+10分) / 有法律文书(+20分)
4. **陪审团投票**：按投票比例调整评分

### 风险等级
- 🟢 **低风险**：0-40分
- 🟡 **中风险**：41-70分
- 🔴 **高风险**：71-100分

---

## 🧪 测试示例

### 1. 提交举报
```bash
curl -X POST "http://localhost:8080/api/v1/company/report" \
  -F "reporter_phone=13800138000" \
  -F "company_name=测试科技有限公司" \
  -F "registration_no=91110108MA01234567" \
  -F "industry=互联网/IT" \
  -F "location_city=北京" \
  -F 'tags=["拖欠工资","强制加班"]' \
  -F "description=该公司长期拖欠工资..."
```

### 2. 查询公司
```bash
curl "http://localhost:8080/api/v1/company/query?company_name=测试科技有限公司&registration_no=91110108MA01234567"
```

### 3. 陪审团投票
```bash
curl -X POST "http://localhost:8080/api/v1/company/vote" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "测试科技有限公司",
    "registration_no": "91110108MA01234567",
    "side": "reporter",
    "fingerprint": "device_fingerprint"
  }'
```

---

## 📈 未来扩展方向

### 功能扩展
- [ ] 公司关注功能（新举报时推送通知）
- [ ] 行业黑名单排行榜
- [ ] 地区风险热力图
- [ ] 导出举报报告（PDF）
- [ ] 企业信用评级API

### 数据增强
- [ ] 对接企查查/天眼查API
- [ ] OCR识别劳动合同
- [ ] NLP分析举报描述
- [ ] 关联分析（同一法人的多家公司）

### 商业化
- [ ] 企业背调服务（B端付费）
- [ ] 猎头/HR会员订阅
- [ ] 数据报告定制

---

## 🐛 故障排查

### 常见问题

**Q: 数据库表没有创建？**  
A: 检查后端日志，确认AutoMigrate是否执行成功。手动检查：
```sql
SELECT table_name FROM information_schema.tables 
WHERE table_name LIKE 'company%';
```

**Q: 前端组件404错误？**  
A: 确认组件已正确导入到 `App.vue`，检查路由配置。

**Q: API返回429错误？**  
A: 触发限流保护，等待1分钟后重试，或检查Redis连接。

**Q: 文件上传失败？**  
A: 检查MinIO是否启动，Bucket是否创建，文件大小是否超过10MB。

**Q: 投票失败（already_voted）？**  
A: 该IP或设备已投过票（30天内），这是正常的防刷机制。

---

## 📞 技术支持

### 查看日志
```bash
# 后端日志
tail -f backend/logs/app.log

# Docker日志
docker-compose logs -f backend
```

### 数据库检查
```sql
-- 查看举报记录数
SELECT COUNT(*) FROM company_records WHERE status = 'active';

-- 查看涉及公司数
SELECT COUNT(DISTINCT company_hash) FROM company_records WHERE status = 'active';

-- 查看行业分布
SELECT industry, COUNT(*) as count 
FROM company_records 
WHERE status = 'active' 
GROUP BY industry 
ORDER BY count DESC;
```

---

## ✅ 完成清单

- [x] 后端数据模型设计
- [x] 后端API实现
- [x] 数据库表和索引创建
- [x] 前端举报表单
- [x] 前端查询界面
- [x] 国际化支持（中英文）
- [x] API测试脚本
- [x] 完整技术文档
- [ ] 前端路由集成（需要你手动完成）
- [ ] 功能测试验证
- [ ] 生产环境部署

---

## 🎉 总结

这次功能扩展为你的平台增加了**公司举报**能力，与原有的**个人举报**功能形成互补：

| 功能 | 个人举报 | 公司举报 |
|------|---------|---------|
| 查询方式 | 手机号 | 公司名+注册号 |
| 举报标签 | 情感欺诈 | 劳动违法 |
| 目标用户 | 恋爱人群 | 求职人群 |
| 商业价值 | C端付费查询 | B端背调服务 |

**核心优势**：
- 🔒 零知识隐私设计，数据库无法反推
- ⚡ Bloom Filter + Hash索引，查询性能优异
- 🛡️ 多层防刷机制，防止恶意攻击
- ⚖️ 陪审团投票，公平公正
- 📱 响应式设计，移动端友好

**下一步**：
1. 完成前端路由集成（5分钟）
2. 运行测试脚本验证功能
3. 部署到生产环境
4. 开始收集真实数据

祝你的平台越做越好！🚀

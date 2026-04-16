#!/bin/bash

# 公司举报功能 API 测试脚本
# 使用方法: chmod +x test-company-api.sh && ./test-company-api.sh

BASE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "公司举报功能 API 测试"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试1: 提交公司举报
echo -e "${YELLOW}测试1: 提交公司举报${NC}"
REPORT_RESPONSE=$(curl -s -X POST "$BASE_URL/company/report" \
  -F "reporter_phone=13800138000" \
  -F "company_name=测试科技有限公司" \
  -F "registration_no=91110108MA01234567" \
  -F "industry=互联网/IT" \
  -F "location_city=北京" \
  -F "employment_period=2023.01-2024.06" \
  -F "position=软件工程师" \
  -F 'tags=["拖欠工资","强制加班"]' \
  -F "description=该公司长期拖欠工资，强制员工无偿加班，且不缴纳社保。多次沟通无果，已向劳动仲裁部门投诉。")

echo "$REPORT_RESPONSE" | jq .
if echo "$REPORT_RESPONSE" | grep -q "company_report_submitted"; then
  echo -e "${GREEN}✓ 举报提交成功${NC}"
else
  echo -e "${RED}✗ 举报提交失败${NC}"
fi
echo ""

# 等待数据写入
sleep 1

# 测试2: 查询公司记录（存在记录）
echo -e "${YELLOW}测试2: 查询公司记录（存在记录）${NC}"
QUERY_RESPONSE=$(curl -s -X GET "$BASE_URL/company/query?registration_no=91110108MA01234567")

echo "$QUERY_RESPONSE" | jq .
if echo "$QUERY_RESPONSE" | grep -q "warning"; then
  echo -e "${GREEN}✓ 查询成功，找到举报记录${NC}"
  COMPANY_HASH=$(echo "$QUERY_RESPONSE" | jq -r '.query_token')
  echo "Company Hash: $COMPANY_HASH"
else
  echo -e "${RED}✗ 查询失败或无记录${NC}"
fi
echo ""

# 测试3: 查询公司记录（不存在记录）
echo -e "${YELLOW}测试3: 查询公司记录（不存在记录）${NC}"
CLEAN_RESPONSE=$(curl -s -X GET "$BASE_URL/company/query?registration_no=99999999999999999")

echo "$CLEAN_RESPONSE" | jq .
if echo "$CLEAN_RESPONSE" | grep -q "clean"; then
  echo -e "${GREEN}✓ 查询成功，无举报记录${NC}"
else
  echo -e "${RED}✗ 查询失败${NC}"
fi
echo ""

# 测试4: 公司提交申诉
echo -e "${YELLOW}测试4: 公司提交申诉${NC}"
APPEAL_RESPONSE=$(curl -s -X POST "$BASE_URL/company/appeal" \
  -F "contact_phone=13900139000" \
  -F "registration_no=91110108MA01234567" \
  -F "reason=举报内容不实，我司从未拖欠工资，所有员工均按时发放薪资并缴纳五险一金。该举报人因违反公司规章制度被依法解除劳动合同，现恶意报复。")

echo "$APPEAL_RESPONSE" | jq .
if echo "$APPEAL_RESPONSE" | grep -q "company_appeal_submitted"; then
  echo -e "${GREEN}✓ 申诉提交成功${NC}"
else
  echo -e "${RED}✗ 申诉提交失败${NC}"
fi
echo ""

# 测试5: 陪审团投票（支持举报人）
echo -e "${YELLOW}测试5: 陪审团投票（支持举报人）${NC}"
VOTE_REPORTER_RESPONSE=$(curl -s -X POST "$BASE_URL/company/vote" \
  -H "Content-Type: application/json" \
  -d '{
    "registration_no": "91110108MA01234567",
    "side": "reporter",
    "fingerprint": "test_fingerprint_001"
  }')

echo "$VOTE_REPORTER_RESPONSE" | jq .
if echo "$VOTE_REPORTER_RESPONSE" | grep -q "vote_recorded"; then
  echo -e "${GREEN}✓ 投票成功（支持举报人）${NC}"
else
  echo -e "${RED}✗ 投票失败${NC}"
fi
echo ""

# 测试6: 陪审团投票（支持公司）- 使用不同IP模拟
echo -e "${YELLOW}测试6: 陪审团投票（支持公司）${NC}"
VOTE_COMPANY_RESPONSE=$(curl -s -X POST "$BASE_URL/company/vote" \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 192.168.1.100" \
  -d '{
    "registration_no": "91110108MA01234567",
    "side": "company",
    "fingerprint": "test_fingerprint_002"
  }')

echo "$VOTE_COMPANY_RESPONSE" | jq .
if echo "$VOTE_COMPANY_RESPONSE" | grep -q "vote_recorded"; then
  echo -e "${GREEN}✓ 投票成功（支持公司）${NC}"
else
  echo -e "${RED}✗ 投票失败${NC}"
fi
echo ""

# 测试7: 重复投票测试（应该被拦截）
echo -e "${YELLOW}测试7: 重复投票测试（应该被拦截）${NC}"
DUPLICATE_VOTE_RESPONSE=$(curl -s -X POST "$BASE_URL/company/vote" \
  -H "Content-Type: application/json" \
  -d '{
    "registration_no": "91110108MA01234567",
    "side": "reporter",
    "fingerprint": "test_fingerprint_001"
  }')

echo "$DUPLICATE_VOTE_RESPONSE" | jq .
if echo "$DUPLICATE_VOTE_RESPONSE" | grep -q "already_voted"; then
  echo -e "${GREEN}✓ 重复投票被正确拦截${NC}"
else
  echo -e "${RED}✗ 重复投票拦截失败${NC}"
fi
echo ""

# 测试8: 查询公司统计数据
echo -e "${YELLOW}测试8: 查询公司统计数据${NC}"
STATS_RESPONSE=$(curl -s -X GET "$BASE_URL/company/stats")

echo "$STATS_RESPONSE" | jq .
if echo "$STATS_RESPONSE" | grep -q "total_reports"; then
  echo -e "${GREEN}✓ 统计数据查询成功${NC}"
  echo "总举报数: $(echo "$STATS_RESPONSE" | jq -r '.total_reports')"
  echo "涉及公司数: $(echo "$STATS_RESPONSE" | jq -r '.total_companies')"
else
  echo -e "${RED}✗ 统计数据查询失败${NC}"
fi
echo ""

# 测试9: 再次查询公司记录（验证投票后的状态）
echo -e "${YELLOW}测试9: 再次查询公司记录（验证投票后的状态）${NC}"
FINAL_QUERY_RESPONSE=$(curl -s -X GET "$BASE_URL/company/query?registration_no=91110108MA01234567")

echo "$FINAL_QUERY_RESPONSE" | jq '.aggregated_profile | {reporter_votes, company_votes, consensus_risk_score}'
REPORTER_VOTES=$(echo "$FINAL_QUERY_RESPONSE" | jq -r '.aggregated_profile.reporter_votes')
COMPANY_VOTES=$(echo "$FINAL_QUERY_RESPONSE" | jq -r '.aggregated_profile.company_votes')
echo -e "${GREEN}✓ 举报人票数: $REPORTER_VOTES, 公司票数: $COMPANY_VOTES${NC}"
echo ""

# 测试10: 限流测试（快速连续请求）
echo -e "${YELLOW}测试10: 限流测试（快速连续请求）${NC}"
echo "发送11次连续查询请求（限制为10次/分钟）..."
for i in {1..11}; do
  RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/company/query?company_name=测试公司$i")
  if [ "$RESPONSE" == "429" ]; then
    echo -e "${GREEN}✓ 第${i}次请求被限流拦截（HTTP 429）${NC}"
    break
  elif [ "$i" -eq 11 ]; then
    echo -e "${YELLOW}⚠ 未触发限流（可能限流配置较宽松）${NC}"
  fi
done
echo ""

echo "========================================="
echo "测试完成！"
echo "========================================="
echo ""
echo "总结:"
echo "- 如果所有测试都显示 ✓，说明功能正常"
echo "- 如果有 ✗ 出现，请检查后端日志"
echo "- 可以访问前端界面进行可视化测试"
echo ""

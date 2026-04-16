@echo off
chcp 65001 >nul
echo =========================================
echo 公司举报功能 API 测试 (Windows)
echo =========================================
echo.

set BASE_URL=http://localhost:8080/api/v1

echo [测试1] 提交公司举报
curl -X POST "%BASE_URL%/company/report" ^
  -F "reporter_phone=13800138000" ^
  -F "company_name=测试科技有限公司" ^
  -F "registration_no=91110108MA01234567" ^
  -F "industry=互联网/IT" ^
  -F "location_city=北京" ^
  -F "employment_period=2023.01-2024.06" ^
  -F "position=软件工程师" ^
  -F "tags=[\"拖欠工资\",\"强制加班\"]" ^
  -F "description=该公司长期拖欠工资，强制员工无偿加班。"
echo.
echo.

timeout /t 2 >nul

echo [测试2] 查询公司记录（存在记录）
curl -X GET "%BASE_URL%/company/query?registration_no=91110108MA01234567"
echo.
echo.

echo [测试3] 查询公司记录（不存在记录）
curl -X GET "%BASE_URL%/company/query?registration_no=99999999999999999"
echo.
echo.

echo [测试4] 公司提交申诉
curl -X POST "%BASE_URL%/company/appeal" ^
  -F "contact_phone=13900139000" ^
  -F "registration_no=91110108MA01234567" ^
  -F "reason=举报内容不实，我司从未拖欠工资。"
echo.
echo.

echo [测试5] 陪审团投票（支持举报人）
curl -X POST "%BASE_URL%/company/vote" ^
  -H "Content-Type: application/json" ^
  -d "{\"registration_no\":\"91110108MA01234567\",\"side\":\"reporter\",\"fingerprint\":\"test_001\"}"
echo.
echo.

echo [测试6] 查询公司统计数据
curl -X GET "%BASE_URL%/company/stats"
echo.
echo.

echo =========================================
echo 测试完成！
echo =========================================
echo.
echo 提示：
echo 1. 确保后端服务已启动（端口8080）
echo 2. 确保PostgreSQL、Redis、MinIO已运行
echo 3. 可以访问 http://localhost:8080 查看前端界面
echo.
pause

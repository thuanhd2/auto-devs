#!/bin/bash

# Script curl đơn giản nhất để kiểm tra GitHub rate limit
# Thay YOUR_TOKEN bằng GitHub token thực của bạn

echo "🔍 Kiểm tra GitHub API rate limit..."

# Thay YOUR_TOKEN bằng token thực
GITHUB_TOKEN="YOUR_TOKEN"

# Kiểm tra rate limit
curl -s \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: curl-rate-limit-checker" \
    "https://api.github.com/rate_limit"

echo ""
echo "📝 Để sử dụng:"
echo "1. Thay YOUR_TOKEN bằng GitHub Personal Access Token thực"
echo "2. Chạy: bash check-rate-limit.sh"
echo ""
echo "📚 Hoặc chạy trực tiếp với curl:"
echo "curl -H \"Authorization: token YOUR_TOKEN\" https://api.github.com/rate_limit"

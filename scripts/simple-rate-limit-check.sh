#!/bin/bash

# Script đơn giản để kiểm tra GitHub API rate limit
# Sử dụng: ./simple-rate-limit-check.sh [GITHUB_TOKEN]

GITHUB_TOKEN="${1:-$GITHUB_TOKEN}"

if [[ -z "$GITHUB_TOKEN" ]]; then
    echo "❌ Cần cung cấp GitHub token"
    echo "Sử dụng: $0 YOUR_GITHUB_TOKEN"
    exit 1
fi

echo "🔍 Kiểm tra GitHub API rate limit..."
echo ""

# Kiểm tra rate limit
curl -s \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: auto-devs-checker" \
    "https://api.github.com/rate_limit" | \
    jq -r '
        "📊 Rate Limit Status:",
        "Core API: \(.rate.remaining)/\(.rate.limit) requests remaining",
        "Reset at: \(.rate.reset | strftime("%Y-%m-%d %H:%M:%S UTC"))",
        "",
        if .resources.search then
            "Search API: \(.resources.search.remaining)/\(.resources.search.limit) requests remaining",
            "Search reset at: \(.resources.search.reset | strftime("%Y-%m-%d %H:%M:%S UTC"))",
            ""
        else empty end,
        if .resources.graphql then
            "GraphQL API: \(.resources.graphql.remaining)/\(.resources.graphql.limit) points remaining",
            "GraphQL reset at: \(.resources.graphql.reset | strftime("%Y-%m-%d %H:%M:%S UTC"))"
        else empty end
    ' 2>/dev/null || echo "❌ Cần cài đặt 'jq' để hiển thị kết quả đẹp"

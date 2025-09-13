#!/bin/bash

# Script để kiểm tra GitHub API rate limit
# Sử dụng: ./check-github-rate-limit.sh [GITHUB_TOKEN] [REPO_OWNER/REPO_NAME] [PR_NUMBER]

set -e

# Màu sắc cho output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Hàm hiển thị help
show_help() {
    echo "Script kiểm tra GitHub API rate limit"
    echo ""
    echo "Cách sử dụng:"
    echo "  $0 [GITHUB_TOKEN] [REPO_OWNER/REPO_NAME] [PR_NUMBER]"
    echo ""
    echo "Tham số:"
    echo "  GITHUB_TOKEN      - GitHub Personal Access Token (bắt buộc)"
    echo "  REPO_OWNER/REPO_NAME - Repository dạng owner/repo (tùy chọn)"
    echo "  PR_NUMBER         - Số Pull Request để test (tùy chọn)"
    echo ""
    echo "Ví dụ:"
    echo "  $0 ghp_xxxxxxxxxxxxxxx"
    echo "  $0 ghp_xxxxxxxxxxxxxxx facebook/react 12345"
    echo ""
    echo "Biến môi trường:"
    echo "  GITHUB_TOKEN      - Có thể set token qua biến môi trường"
    echo ""
    echo "Rate limit headers được kiểm tra:"
    echo "  X-RateLimit-Limit     - Tổng số request được phép"
    echo "  X-RateLimit-Remaining - Số request còn lại"
    echo "  X-RateLimit-Reset     - Timestamp khi rate limit reset"
    echo "  X-RateLimit-Used      - Số request đã sử dụng"
}

# Kiểm tra tham số
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Lấy token từ tham số hoặc biến môi trường
GITHUB_TOKEN="${1:-$GITHUB_TOKEN}"
if [[ -z "$GITHUB_TOKEN" ]]; then
    echo -e "${RED}Lỗi: Cần cung cấp GitHub token${NC}"
    echo "Sử dụng: $0 -h để xem hướng dẫn"
    exit 1
fi

# Repository và PR number (tùy chọn)
REPO="${2:-}"
PR_NUMBER="${3:-}"

# Base URL cho GitHub API
BASE_URL="https://api.github.com"

echo -e "${BLUE}=== GitHub API Rate Limit Checker ===${NC}"
echo ""

# Hàm format timestamp
format_timestamp() {
    local timestamp=$1
    if command -v date >/dev/null 2>&1; then
        date -d "@$timestamp" 2>/dev/null || date -r "$timestamp" 2>/dev/null || echo "$timestamp"
    else
        echo "$timestamp"
    fi
}

# Hàm tính thời gian còn lại
calculate_time_remaining() {
    local reset_timestamp=$1
    local current_timestamp=$(date +%s)
    local remaining_seconds=$((reset_timestamp - current_timestamp))

    if [[ $remaining_seconds -gt 0 ]]; then
        local hours=$((remaining_seconds / 3600))
        local minutes=$(((remaining_seconds % 3600) / 60))
        local seconds=$((remaining_seconds % 60))

        if [[ $hours -gt 0 ]]; then
            echo "${hours}h ${minutes}m ${seconds}s"
        elif [[ $minutes -gt 0 ]]; then
            echo "${minutes}m ${seconds}s"
        else
            echo "${seconds}s"
        fi
    else
        echo "Đã reset"
    fi
}

# Hàm kiểm tra rate limit từ response headers
check_rate_limit_from_headers() {
    local response_file=$1
    local endpoint_name=$2

    echo -e "${YELLOW}--- Rate Limit Info cho $endpoint_name ---${NC}"

    # Đọc headers từ file response
    local limit=$(grep -i "x-ratelimit-limit:" "$response_file" | cut -d' ' -f2- | tr -d '\r' || echo "N/A")
    local remaining=$(grep -i "x-ratelimit-remaining:" "$response_file" | cut -d' ' -f2- | tr -d '\r' || echo "N/A")
    local reset=$(grep -i "x-ratelimit-reset:" "$response_file" | cut -d' ' -f2- | tr -d '\r' || echo "N/A")
    local used=$(grep -i "x-ratelimit-used:" "$response_file" | cut -d' ' -f2- | tr -d '\r' || echo "N/A")

    # Hiển thị thông tin rate limit
    echo "  Limit:     $limit"
    echo "  Remaining: $remaining"
    echo "  Used:      $used"

    if [[ "$reset" != "N/A" && "$reset" =~ ^[0-9]+$ ]]; then
        local reset_time=$(format_timestamp "$reset")
        local time_remaining=$(calculate_time_remaining "$reset")
        echo "  Reset at:  $reset_time"
        echo "  Time left: $time_remaining"

        # Cảnh báo nếu rate limit thấp
        if [[ "$remaining" =~ ^[0-9]+$ ]] && [[ $remaining -lt 100 ]]; then
            echo -e "  ${RED}⚠️  Cảnh báo: Rate limit còn lại thấp!${NC}"
        fi

        # Cảnh báo nếu đã hết rate limit
        if [[ "$remaining" =~ ^[0-9]+$ ]] && [[ $remaining -eq 0 ]]; then
            echo -e "  ${RED}🚫 Rate limit đã hết! Cần chờ reset.${NC}"
        fi
    fi

    echo ""
}

# Test 1: Kiểm tra rate limit từ endpoint /rate_limit
echo -e "${GREEN}1. Kiểm tra rate limit từ endpoint /rate_limit${NC}"
response_file=$(mktemp)
http_code=$(curl -s -w "%{http_code}" \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: auto-devs-rate-limit-checker" \
    -o "$response_file" \
    "$BASE_URL/rate_limit")

if [[ "$http_code" == "200" ]]; then
    echo "✅ Kết nối thành công"

    # Parse JSON response để lấy thông tin chi tiết
    if command -v jq >/dev/null 2>&1; then
        echo ""
        echo "Chi tiết rate limit:"
        jq -r '.rate | "Core API: \(.limit) requests/hour, còn lại: \(.remaining), reset lúc: \(.reset | strftime("%Y-%m-%d %H:%M:%S UTC"))"' "$response_file"

        # Kiểm tra search rate limit nếu có
        search_limit=$(jq -r '.resources.search.limit // empty' "$response_file")
        if [[ -n "$search_limit" ]]; then
            search_remaining=$(jq -r '.resources.search.remaining // empty' "$response_file")
            search_reset=$(jq -r '.resources.search.reset // empty' "$response_file")
            echo "Search API: $search_limit requests/hour, còn lại: $search_remaining, reset lúc: $(date -d "@$search_reset" 2>/dev/null || echo $search_reset)"
        fi

        # Kiểm tra graphql rate limit nếu có
        graphql_limit=$(jq -r '.resources.graphql.limit // empty' "$response_file")
        if [[ -n "$graphql_limit" ]]; then
            graphql_remaining=$(jq -r '.resources.graphql.remaining // empty' "$response_file")
            graphql_reset=$(jq -r '.resources.graphql.reset // empty' "$response_file")
            echo "GraphQL API: $graphql_limit requests/hour, còn lại: $graphql_remaining, reset lúc: $(date -d "@$graphql_reset" 2>/dev/null || echo $graphql_reset)"
        fi
    else
        echo "⚠️  Cài đặt 'jq' để xem thông tin chi tiết hơn"
        cat "$response_file"
    fi
else
    echo -e "${RED}❌ Lỗi HTTP $http_code${NC}"
    cat "$response_file"
fi

rm -f "$response_file"
echo ""

# Test 2: Kiểm tra rate limit từ endpoint /user (authenticated)
echo -e "${GREEN}2. Kiểm tra rate limit từ endpoint /user${NC}"
response_file=$(mktemp)
http_code=$(curl -s -w "%{http_code}" \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: auto-devs-rate-limit-checker" \
    -D "$response_file" \
    "$BASE_URL/user" >/dev/null)

if [[ "$http_code" == "200" ]]; then
    echo "✅ Xác thực thành công"
    check_rate_limit_from_headers "$response_file" "/user endpoint"
else
    echo -e "${RED}❌ Lỗi xác thực HTTP $http_code${NC}"
fi

rm -f "$response_file"

# Test 3: Kiểm tra rate limit từ Pull Request endpoint (nếu có repo và PR)
if [[ -n "$REPO" && -n "$PR_NUMBER" ]]; then
    echo -e "${GREEN}3. Kiểm tra rate limit từ Pull Request endpoint${NC}"
    response_file=$(mktemp)
    http_code=$(curl -s -w "%{http_code}" \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        -H "User-Agent: auto-devs-rate-limit-checker" \
        -D "$response_file" \
        "$BASE_URL/repos/$REPO/pulls/$PR_NUMBER" >/dev/null)

    if [[ "$http_code" == "200" ]]; then
        echo "✅ Lấy Pull Request thành công"
        check_rate_limit_from_headers "$response_file" "Pull Request endpoint"
    else
        echo -e "${RED}❌ Lỗi lấy Pull Request HTTP $http_code${NC}"
        echo "Kiểm tra lại repository và PR number"
    fi

    rm -f "$response_file"
elif [[ -n "$REPO" ]]; then
    echo -e "${GREEN}3. Kiểm tra rate limit từ Repository endpoint${NC}"
    response_file=$(mktemp)
    http_code=$(curl -s -w "%{http_code}" \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        -H "User-Agent: auto-devs-rate-limit-checker" \
        -D "$response_file" \
        "$BASE_URL/repos/$REPO" >/dev/null)

    if [[ "$http_code" == "200" ]]; then
        echo "✅ Truy cập repository thành công"
        check_rate_limit_from_headers "$response_file" "Repository endpoint"
    else
        echo -e "${RED}❌ Lỗi truy cập repository HTTP $http_code${NC}"
        echo "Kiểm tra lại repository name"
    fi

    rm -f "$response_file"
fi

# Test 4: Kiểm tra rate limit từ Search endpoint
echo -e "${GREEN}4. Kiểm tra rate limit từ Search endpoint${NC}"
response_file=$(mktemp)
http_code=$(curl -s -w "%{http_code}" \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: auto-devs-rate-limit-checker" \
    -D "$response_file" \
    "$BASE_URL/search/repositories?q=github" >/dev/null)

if [[ "$http_code" == "200" ]]; then
    echo "✅ Search API hoạt động"
    check_rate_limit_from_headers "$response_file" "Search endpoint"
else
    echo -e "${RED}❌ Lỗi Search API HTTP $http_code${NC}"
fi

rm -f "$response_file"

echo -e "${BLUE}=== Kết thúc kiểm tra ===${NC}"
echo ""
echo "💡 Mẹo:"
echo "  - Rate limit reset mỗi giờ"
echo "  - Authenticated users: 5,000 requests/hour"
echo "  - Unauthenticated users: 60 requests/hour"
echo "  - Search API: 30 requests/minute"
echo "  - GraphQL API: 5,000 points/hour"
echo ""
echo "📚 Tài liệu: https://docs.github.com/en/rest/rate-limit"

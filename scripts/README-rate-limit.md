# GitHub API Rate Limit Checker Scripts

Bộ script để kiểm tra rate limit của GitHub API, đặc biệt cho các hàm Pull Request.

## Các Script Có Sẵn

### 1. `check-github-rate-limit.sh` - Script Đầy Đủ

Script chi tiết nhất với nhiều tính năng:

```bash
# Cách sử dụng cơ bản
./scripts/check-github-rate-limit.sh YOUR_GITHUB_TOKEN

# Với repository và PR cụ thể
./scripts/check-github-rate-limit.sh YOUR_GITHUB_TOKEN facebook/react 12345

# Sử dụng biến môi trường
export GITHUB_TOKEN="your_token_here"
./scripts/check-github-rate-limit.sh
```

**Tính năng:**

- ✅ Kiểm tra rate limit từ endpoint `/rate_limit`
- ✅ Kiểm tra rate limit từ endpoint `/user` (xác thực)
- ✅ Kiểm tra rate limit từ Pull Request endpoint
- ✅ Kiểm tra rate limit từ Search endpoint
- ✅ Hiển thị thời gian reset và cảnh báo
- ✅ Hỗ trợ màu sắc và format đẹp
- ✅ Tính toán thời gian còn lại

### 2. `simple-rate-limit-check.sh` - Script Đơn Giản

Script ngắn gọn với jq để format JSON:

```bash
./scripts/simple-rate-limit-check.sh YOUR_GITHUB_TOKEN
```

**Tính năng:**

- ✅ Kiểm tra rate limit cơ bản
- ✅ Format JSON đẹp với jq
- ✅ Hiển thị thông tin Core API, Search API, GraphQL API

### 3. `check-rate-limit.sh` - Script Cơ Bản

Script đơn giản nhất, chỉ cần chỉnh sửa token:

```bash
# Chỉnh sửa YOUR_TOKEN trong file, sau đó chạy:
./scripts/check-rate-limit.sh
```

## Cách Lấy GitHub Token

1. Truy cập GitHub Settings: https://github.com/settings/tokens
2. Click "Generate new token" → "Generate new token (classic)"
3. Chọn scopes cần thiết:
   - `repo` - Full control of private repositories
   - `public_repo` - Access public repositories
4. Copy token và sử dụng trong script

## Rate Limit GitHub API

### Core API

- **Authenticated users**: 5,000 requests/hour
- **Unauthenticated users**: 60 requests/hour

### Search API

- **Authenticated users**: 30 requests/minute
- **Unauthenticated users**: 10 requests/minute

### GraphQL API

- **Authenticated users**: 5,000 points/hour

## Headers Quan Trọng

GitHub API trả về các header sau trong response:

```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1640995200
X-RateLimit-Used: 1
```

## Ví Dụ Sử Dụng

### Kiểm tra rate limit cơ bản:

```bash
curl -H "Authorization: token YOUR_TOKEN" \
     -H "Accept: application/vnd.github.v3+json" \
     https://api.github.com/rate_limit
```

### Kiểm tra rate limit từ Pull Request:

```bash
curl -H "Authorization: token YOUR_TOKEN" \
     -H "Accept: application/vnd.github.v3+json" \
     https://api.github.com/repos/facebook/react/pulls/12345
```

### Kiểm tra headers rate limit:

```bash
curl -I -H "Authorization: token YOUR_TOKEN" \
     -H "Accept: application/vnd.github.v3+json" \
     https://api.github.com/user
```

## Troubleshooting

### Lỗi 401 Unauthorized

- Kiểm tra token có đúng không
- Kiểm tra token có hết hạn không
- Kiểm tra token có đủ quyền không

### Lỗi 403 Forbidden (Rate Limited)

- Đã hết rate limit
- Chờ đến thời gian reset
- Sử dụng token khác

### Lỗi 404 Not Found

- Repository không tồn tại
- Pull Request không tồn tại
- Kiểm tra lại tên repository và số PR

## Tài Liệu Tham Khảo

- [GitHub API Rate Limiting](https://docs.github.com/en/rest/rate-limit)
- [GitHub API Authentication](https://docs.github.com/en/rest/authentication)
- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)

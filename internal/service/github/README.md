# GitHub Service

Package này cung cấp các service để tương tác với GitHub API.

## Services

### GitHubServiceV2

`GitHubServiceV2` là service mới sử dụng thư viện `go-github` thay vì tự tạo HTTP client. Service này implement đầy đủ `GitHubServiceInterface`.

#### Tính năng

- **Tạo Pull Request**: Tạo PR mới với title, body, base branch và head branch
- **Lấy Pull Request**: Lấy thông tin chi tiết của PR theo số PR
- **Cập nhật Pull Request**: Cập nhật title, body, assignees, labels, reviewers
- **Merge Pull Request**: Merge PR với các method khác nhau (merge, squash, rebase)
- **Validate Token**: Kiểm tra tính hợp lệ của GitHub token
- **Rate Limiting**: Tự động xử lý GitHub API rate limits

#### Cách sử dụng

```go
package main

import (
    "context"
    "log"

    "github.com/auto-devs/auto-devs/internal/service/github"
)

func main() {
    // Tạo config
    config := &github.GitHubConfig{
        Token:     "your-github-token",
        BaseURL:   "https://api.github.com", // hoặc GitHub Enterprise URL
        UserAgent: "my-app/1.0",
        Timeout:   30,
    }

    // Tạo service
    service := github.NewGitHubServiceV2(config)

    // Sử dụng service
    ctx := context.Background()

    // Tạo PR
    pr, err := service.CreatePullRequest(ctx, "owner/repo", "main", "feature-branch", "Title", "Description")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created PR: %s", pr.Title)
}
```

#### Cấu hình

- **Token**: GitHub Personal Access Token hoặc OAuth token
- **BaseURL**: GitHub API base URL (mặc định: https://api.github.com)
- **UserAgent**: User agent string cho API requests
- **Timeout**: Timeout cho HTTP requests (giây)

#### Rate Limiting

Service tự động xử lý GitHub API rate limits:

- Theo dõi rate limit headers từ GitHub response
- Cập nhật thông tin rate limit real-time
- Cung cấp thông tin về limit, remaining và reset time

#### Error Handling

Service trả về các error có ý nghĩa:

- Validation errors cho repository format
- GitHub API errors với context
- Rate limit errors khi cần thiết

#### Testing

Service có đầy đủ unit tests:

```bash
go test ./internal/service/github/ -run "TestGitHubServiceV2" -v
```

## Migration từ GitHubService cũ

Nếu bạn đang sử dụng `GitHubService` cũ, bạn có thể dễ dàng migrate sang `GitHubServiceV2`:

1. Thay đổi import và khởi tạo service
2. Các method calls giữ nguyên interface
3. Service mới có performance tốt hơn và error handling tốt hơn

## Dependencies

- `github.com/google/go-github/v74/github`: GitHub API client library
- `golang.org/x/oauth2`: OAuth2 authentication
- `github.com/auto-devs/auto-devs/internal/entity`: Entity definitions

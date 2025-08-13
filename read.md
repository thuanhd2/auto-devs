# Tóm tắt nội dung README.md

File README.md đã được đọc thành công và có nội dung như sau:

## Thông tin dự án
**Auto-Devs API Core** - Hệ thống API cho dự án quản lý Auto-Devs

## Yêu cầu hệ thống
- Go 1.24+
- PostgreSQL  
- Make (tùy chọn)

## Tính năng chính
- API RESTful với validation toàn diện
- Tài liệu OpenAPI/Swagger
- Cấu hình CORS cho frontend
- Rate limiting và security headers
- Request logging và error handling
- Database migrations với GORM
- Kiến trúc Clean Architecture

## API Documentation
- Swagger UI: http://localhost:8098/swagger/index.html
- Swagger JSON: http://localhost:8098/swagger.json
- Swagger YAML: http://localhost:8098/swagger.yaml

## Cấu trúc kiến trúc (4 tầng)
1. **DTO Layer** - Request/response models
2. **Handler Layer** - HTTP request handlers  
3. **Usecase Layer** - Business logic
4. **Repository Layer** - Data access

## API Endpoints chính
- **Health Check**: GET /api/v1/health
- **Projects**: CRUD operations cho projects
- **Tasks**: CRUD operations cho tasks với filtering

## Lệnh phát triển
- `make build` - Build ứng dụng
- `make run` - Chạy ứng dụng
- `make test` - Chạy tests
- `make swagger` - Tạo Swagger documentation
- `make migrate-up/down/reset` - Quản lý database migrations

## Cấu trúc thư mục
Dự án được tổ chức theo Clean Architecture với các thư mục chính: cmd/, config/, docs/, internal/, migrations/, pkg/, scripts/

---
*Tóm tắt được tạo bởi Serena MCP từ file README.md có sẵn trong dự án.*
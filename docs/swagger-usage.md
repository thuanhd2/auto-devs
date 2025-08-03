# Swagger Documentation

## Truy cập Swagger UI

Sau khi khởi động server, bạn có thể truy cập Swagger UI qua các cách sau:

### 1. Trực tiếp qua browser

- **Swagger UI**: http://localhost:8098/swagger/index.html
- **Root redirect**: http://localhost:8098/ (sẽ tự động redirect đến Swagger UI)

### 2. API Documentation Files

- **Swagger JSON**: http://localhost:8098/swagger.json
- **Swagger YAML**: http://localhost:8098/swagger.yaml

## Các Endpoint có sẵn

### Health Check

- `GET /api/v1/health` - Kiểm tra trạng thái server và database

### Projects

- `POST /api/v1/projects` - Tạo project mới
- `GET /api/v1/projects` - Lấy danh sách tất cả projects
- `GET /api/v1/projects/{id}` - Lấy project theo ID
- `PUT /api/v1/projects/{id}` - Cập nhật project
- `DELETE /api/v1/projects/{id}` - Xóa project
- `GET /api/v1/projects/{id}/tasks` - Lấy project với danh sách tasks

### Tasks

- `POST /api/v1/tasks` - Tạo task mới
- `GET /api/v1/tasks` - Lấy danh sách tasks với filtering
- `GET /api/v1/tasks/{id}` - Lấy task theo ID
- `PUT /api/v1/tasks/{id}` - Cập nhật task
- `DELETE /api/v1/tasks/{id}` - Xóa task
- `PATCH /api/v1/tasks/{id}/status` - Cập nhật trạng thái task
- `GET /api/v1/tasks/{id}/project` - Lấy task với thông tin project

## Cách sử dụng Swagger UI

1. **Mở Swagger UI**: Truy cập http://localhost:8098/swagger/index.html
2. **Chọn endpoint**: Click vào endpoint bạn muốn test
3. **Click "Try it out"**: Để mở form input
4. **Nhập parameters**: Điền các thông tin cần thiết
5. **Execute**: Click "Execute" để gửi request
6. **Xem response**: Kết quả sẽ hiển thị bên dưới

## Ví dụ sử dụng

### Tạo Project

```bash
curl -X POST http://localhost:8098/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "description": "A sample project",
    "repo_url": "https://github.com/user/repo"
  }'
```

### Tạo Task

```bash
curl -X POST http://localhost:8098/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Implement feature",
    "description": "Add new functionality"
  }'
```

## Lưu ý

- Server chạy trên port 8098 mặc định
- Tất cả API endpoints đều có prefix `/api/v1`
- Swagger UI hỗ trợ test trực tiếp các API endpoints
- Có thể export OpenAPI specification để sử dụng với các công cụ khác

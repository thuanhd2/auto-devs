# Quick Start: Job Worker

Hướng dẫn nhanh để chạy job worker trong 5 phút.

## Bước 1: Cài đặt Redis

### macOS

```bash
brew install redis
brew services start redis
```

### Ubuntu/Debian

```bash
sudo apt-get install redis-server
sudo systemctl start redis-server
```

### Docker

```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

## Bước 2: Kiểm tra Redis

```bash
redis-cli ping
# Kết quả: PONG
```

## Bước 3: Cấu hình Database

Đảm bảo PostgreSQL đang chạy và có database:

```bash
# Tạo database nếu chưa có
createdb autodevs_dev

# Chạy migrations
make migrate-up
```

## Bước 4: Chạy Worker

### Cách 1: Sử dụng Makefile (Khuyến nghị)

```bash
# Chạy worker với cấu hình mặc định
make run-worker

# Chạy với verbose logging
make run-worker-verbose

# Chạy với tên tùy chỉnh
make run-worker-named name=my-worker
```

### Cách 2: Sử dụng Script

```bash
# Chạy worker
./scripts/run-worker.sh

# Chạy với verbose
./scripts/run-worker.sh -v

# Chạy với tên tùy chỉnh
./scripts/run-worker.sh -n my-worker
```

### Cách 3: Chạy trực tiếp

```bash
# Build worker
go build -o worker ./cmd/worker

# Chạy worker
./worker -worker=my-worker -verbose
```

## Bước 5: Test Worker

### Tạo test job

```bash
# Chạy test example
go run examples/test-worker.go
```

### Hoặc sử dụng Redis CLI

```bash
# Kiểm tra queue
redis-cli LLEN asynq:queues:planning

# Xem jobs trong queue
redis-cli LRANGE asynq:queues:planning 0 -1
```

## Bước 6: Monitor Logs

Worker sẽ output logs như sau:

```
time=2024-01-01T12:00:00.000Z level=INFO msg="Starting job worker" worker_name=my-worker
time=2024-01-01T12:00:00.001Z level=INFO msg="Starting job server" redis_addr=localhost:6379
time=2024-01-01T12:00:01.000Z level=INFO msg="Processing task planning job" task_id=123e4567-e89b-12d3-a456-426614174000
```

## Troubleshooting

### Worker không start

```bash
# Kiểm tra Redis
redis-cli ping

# Kiểm tra database
psql -d autodevs_dev -c "SELECT 1"

# Chạy với verbose để xem lỗi
make run-worker-verbose
```

### Jobs không được xử lý

```bash
# Kiểm tra queue
redis-cli LLEN asynq:queues:planning

# Restart worker
pkill worker
make run-worker
```

## Next Steps

1. **Chạy nhiều workers**: Mở terminal mới và chạy `make run-worker-named name=worker-2`
2. **Production setup**: Xem `docs/job-worker-guide.md`
3. **Monitoring**: Setup monitoring với Prometheus/Grafana
4. **Scaling**: Sử dụng Supervisor hoặc Systemd

## Help

```bash
# Xem help
make worker-help

# Hoặc
./scripts/run-worker.sh -h
```

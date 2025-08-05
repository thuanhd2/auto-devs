# Job Worker Guide

Hướng dẫn chi tiết về cách chạy và quản lý background job processor.

## Tổng quan

Job Worker là một process riêng biệt chạy song song với main application để xử lý các background jobs như:

- Task planning jobs
- Git worktree operations
- AI execution tasks

## Yêu cầu hệ thống

### 1. Redis Server

Job worker cần Redis để lưu trữ job queue:

```bash
# Cài đặt Redis (Ubuntu/Debian)
sudo apt-get install redis-server

# Cài đặt Redis (macOS)
brew install redis

# Cài đặt Redis (Docker)
docker run -d --name redis -p 6379:6379 redis:alpine
```

### 2. Go Environment

Đảm bảo Go đã được cài đặt và cấu hình đúng.

### 3. Database

PostgreSQL database phải được cấu hình và chạy.

## Cấu hình

### Environment Variables

Tạo file `.env` hoặc set environment variables:

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
DB_NAME=autodevs
DB_SSLMODE=disable

# Worktree Configuration
WORKTREE_BASE_DIR=/tmp/worktrees
WORKTREE_MAX_PATH_LENGTH=4096
WORKTREE_MIN_DISK_SPACE=104857600
WORKTREE_CLEANUP_INTERVAL=24h
WORKTREE_ENABLE_LOGGING=true
```

## Cách chạy Job Worker

### 1. Sử dụng Script (Khuyến nghị)

```bash
# Chạy với cấu hình mặc định
./scripts/run-worker.sh

# Chạy với tên worker tùy chỉnh
./scripts/run-worker.sh -n planning-worker-1

# Chạy với verbose logging
./scripts/run-worker.sh -v

# Chạy với Redis host tùy chỉnh
REDIS_HOST=redis.example.com ./scripts/run-worker.sh

# Xem help
./scripts/run-worker.sh -h
```

### 2. Chạy trực tiếp

```bash
# Build worker binary
go build -o worker ./cmd/worker

# Chạy worker
./worker -worker=planning-worker-1 -verbose
```

### 3. Chạy trong Docker

```bash
# Build Docker image
docker build -t autodevs-worker .

# Chạy container
docker run -d \
  --name autodevs-worker \
  --network host \
  -e REDIS_HOST=localhost \
  -e DB_HOST=localhost \
  autodevs-worker
```

## Monitoring và Debugging

### 1. Logs

Worker sẽ output logs với format:

```
time=2024-01-01T12:00:00.000Z level=INFO msg="Starting job worker" worker_name=planning-worker-1
time=2024-01-01T12:00:00.001Z level=INFO msg="Starting job server" redis_addr=localhost:6379
time=2024-01-01T12:00:01.000Z level=INFO msg="Processing task planning job" task_id=123e4567-e89b-12d3-a456-426614174000
```

### 2. Redis Queue Monitoring

```bash
# Kiểm tra queue status
redis-cli LLEN asynq:queues:planning

# Xem jobs trong queue
redis-cli LRANGE asynq:queues:planning 0 -1

# Xem failed jobs
redis-cli LLEN asynq:failed

# Xem processing jobs
redis-cli LLEN asynq:processing
```

### 3. Health Check

```bash
# Kiểm tra Redis connection
redis-cli ping

# Kiểm tra worker process
ps aux | grep worker

# Kiểm tra logs
tail -f worker.log
```

## Scaling

### 1. Chạy nhiều workers

```bash
# Terminal 1
./scripts/run-worker.sh -n worker-1

# Terminal 2
./scripts/run-worker.sh -n worker-2

# Terminal 3
./scripts/run-worker.sh -n worker-3
```

### 2. Process Management với Supervisor

Tạo file `/etc/supervisor/conf.d/autodevs-worker.conf`:

```ini
[program:autodevs-worker]
command=/path/to/your/project/worker -worker=%(program_name)s
directory=/path/to/your/project
user=www-data
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/autodevs-worker.log
environment=REDIS_HOST="localhost",DB_HOST="localhost"
```

### 3. Systemd Service

Tạo file `/etc/systemd/system/autodevs-worker.service`:

```ini
[Unit]
Description=AutoDevs Job Worker
After=network.target redis.service

[Service]
Type=simple
User=autodevs
WorkingDirectory=/path/to/your/project
ExecStart=/path/to/your/project/worker -worker=systemd-worker
Restart=always
RestartSec=5
Environment=REDIS_HOST=localhost
Environment=DB_HOST=localhost

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

### 1. Worker không start

```bash
# Kiểm tra Redis connection
redis-cli ping

# Kiểm tra database connection
psql -h localhost -U postgres -d autodevs -c "SELECT 1"

# Kiểm tra logs
./worker -verbose
```

### 2. Jobs không được xử lý

```bash
# Kiểm tra queue status
redis-cli LLEN asynq:queues:planning

# Kiểm tra failed jobs
redis-cli LRANGE asynq:failed 0 -1

# Restart worker
pkill worker
./scripts/run-worker.sh
```

### 3. Performance Issues

```bash
# Tăng concurrency
# Edit internal/jobs/server.go
Concurrency: 8, // Tăng từ 4 lên 8

# Tăng queue priority
Queues: map[string]int{
    "critical": 10,
    "planning": 6,
    "default":  2,
}
```

## Development

### 1. Local Development

```bash
# Start Redis
redis-server

# Start PostgreSQL
sudo service postgresql start

# Run worker in development mode
./scripts/run-worker.sh -v -n dev-worker
```

### 2. Testing

```bash
# Run unit tests
go test ./internal/jobs -v

# Run integration tests
go test ./internal/jobs -tags=integration -v

# Test with real Redis
REDIS_HOST=localhost go test ./internal/jobs -v
```

### 3. Debug Mode

```bash
# Run with debug logging
./worker -verbose -worker=debug-worker

# Use Delve debugger
dlv debug ./cmd/worker -- -worker=debug-worker
```

## Production Deployment

### 1. Environment Setup

```bash
# Production environment variables
export REDIS_HOST=redis.production.com
export REDIS_PASSWORD=your_redis_password
export DB_HOST=db.production.com
export DB_PASSWORD=your_db_password
```

### 2. Security

```bash
# Run worker với user không có quyền root
sudo useradd -r -s /bin/false autodevs
sudo chown autodevs:autodevs /path/to/worker
sudo -u autodevs ./worker
```

### 3. Monitoring

```bash
# Setup monitoring với Prometheus
# Add metrics endpoint to worker

# Setup alerting với Grafana
# Monitor queue length, failed jobs, processing time
```

## Best Practices

1. **Always run multiple workers** để đảm bảo high availability
2. **Monitor queue length** để scale workers khi cần
3. **Set up proper logging** để debug issues
4. **Use health checks** để detect worker failures
5. **Implement retry logic** cho failed jobs
6. **Monitor resource usage** (CPU, memory, disk)
7. **Backup Redis data** định kỳ
8. **Use proper error handling** trong job processing

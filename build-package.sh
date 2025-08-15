#!/bin/bash

set -e  # Exit on any error

CLI_DIR="cmd/npx"
TARGET_DIR="cmd/npx/dist"

echo "🧹 Cleaning up the previous build..."

rm -rf $TARGET_DIR

echo "🏗️ Building frontend..."

cd frontend && VITE_API_BASE_URL=/api/v1 VITE_WS_BASE_URL=ws://base_host/ws npm run build

cd ..

echo "🏗️ Building server & copy fake-cli..."

go build -o $TARGET_DIR/server cmd/server/main.go
cp -r fake-cli $TARGET_DIR/fake-cli

echo "🏗️ Building worker..."

go build -o $TARGET_DIR/worker cmd/worker/main.go

echo "✅ Build complete!"

echo "Copy .env.example to .env if not exists"

if [ ! -f $CLI_DIR/.env ]; then
    cp .env.example $CLI_DIR/.env
fi

echo "Copy frontend build to dist/public"

mkdir -p $TARGET_DIR/public
cp -r frontend/dist/* $TARGET_DIR/public

echo "✅ Package built successfully!"

echo "🎉 Done!"
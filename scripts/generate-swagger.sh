#!/bin/bash

# Generate Swagger Documentation Script
# This script generates Swagger documentation from Go code annotations

echo "🔄 Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "❌ Error: swag CLI is not installed"
    echo "Please install it with: go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

# Generate documentation
echo "📝 Running swag init..."
swag init -g cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    echo "📁 Files created:"
    echo "   - docs/docs.go"
    echo "   - docs/swagger.json"
    echo "   - docs/swagger.yaml"
    echo ""
    echo "🌐 Swagger UI available at:"
    echo "   - http://localhost:8098/swagger/index.html"
    echo "   - http://localhost:8098/ (redirects to Swagger UI)"
    echo ""
    echo "📊 Generated endpoints: $(curl -s http://localhost:8098/swagger.json | grep -o '"summary":' | wc -l | tr -d ' ')"
    echo "📋 Generated schemas: $(curl -s http://localhost:8098/swagger.json | grep -o 'dto\.[A-Za-z]*Response\|dto\.[A-Za-z]*Request' | sort | uniq | wc -l | tr -d ' ')"
else
    echo "❌ Error: Failed to generate Swagger documentation"
    exit 1
fi 
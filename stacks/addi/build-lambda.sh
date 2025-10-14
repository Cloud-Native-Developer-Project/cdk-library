#!/bin/bash

set -e

echo "🔨 Building Lambda: webhook-notifier"
echo "======================================"

# Change to Lambda directory
cd "$(dirname "$0")/lambda/webhook-notifier"

echo "📦 Downloading Go dependencies..."
go mod download

echo "🏗️  Building for ARM64 (Graviton2)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o main main.go

echo "📊 Build info:"
ls -lh main
file main

echo ""
echo "✅ Lambda built successfully!"
echo ""
echo "Next steps:"
echo "  1. Deploy: cdk deploy AddiS3ToSFTPStack"
echo "  2. Test: aws s3 cp test.pdf s3://addi-landing-zone-prod/uploads/"

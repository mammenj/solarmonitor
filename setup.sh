#!/usr/bin/env bash

set -e

echo "🚀 Creating directory structure..."

# Create core project folders
mkdir -p cmd/server
mkdir -p internal/domain
mkdir -p internal/storage
mkdir -p internal/web/templates
mkdir -p internal/web/static/css
mkdir -p internal/web/static/js

echo "📁 Directory structure created."

echo "⬇️ Downloading HTMX and Alpine.js locally..."

# Download HTMX (v1.9.12)
curl -sL "https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" -o internal/web/static/js/htmx.min.js

# Download Alpine.js (v3.13.10)
curl -sL "https://unpkg.com/alpinejs@3.13.10/dist/cdn.min.js" -o internal/web/static/js/alpine.min.js

# Create dummy main.css file
touch internal/web/static/css/main.css

echo "✅ Assets downloaded to internal/web/static/js/"
echo "✨ Project layout initialized successfully!"
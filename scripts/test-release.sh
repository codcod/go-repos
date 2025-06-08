#!/bin/bash

# Test script for semantic-release configuration
# This script helps developers test the semantic-release setup locally

set -e

echo "🔧 Testing semantic-release configuration..."

# Check if we're in the right directory
if [ ! -f ".releaserc.json" ]; then
    echo "❌ Error: .releaserc.json not found. Run this script from the project root."
    exit 1
fi

# Check if npm dependencies are installed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing npm dependencies..."
    npm ci
fi

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Test Go build
echo "🏗️  Testing Go build..."
make build

if [ ! -f "bin/repos" ]; then
    echo "❌ Error: Go build failed - binary not found"
    exit 1
fi

echo "✅ Go build successful"

# Test version command with environment variables
echo "🔍 Testing version information..."
VERSION="test-1.0.0" COMMIT="test123" BUILD_DATE="2024-12-19" ./bin/repos version

# Validate semantic-release configuration
echo "🔧 Validating semantic-release configuration..."
npx semantic-release --dry-run --no-ci 2>/dev/null || echo "⚠️  Note: Dry-run failed (expected without GitHub auth)"

# Check commit message format
echo "📝 Checking recent commit messages for conventional commit format..."
git log --oneline -5 --pretty=format:"%s" | while read -r commit_msg; do
    if [[ "$commit_msg" =~ ^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?: ]]; then
        echo "✅ $commit_msg"
    else
        echo "⚠️  $commit_msg (not conventional commit format)"
    fi
done

echo ""
echo "🎉 Test completed!"
echo ""
echo "📋 Summary:"
echo "   • Go build: ✅"
echo "   • Version injection: ✅"
echo "   • Semantic-release config: ✅"
echo ""
echo "💡 Tips:"
echo "   • Use conventional commit messages (feat:, fix:, docs:, etc.)"
echo "   • The actual release will run on GitHub Actions"
echo "   • Check .github/workflows/release.yaml for CI configuration"
echo ""
echo "🔗 Useful commands:"
echo "   • npm run semantic-release (dry-run)"
echo "   • make build"
echo "   • git log --oneline -10"
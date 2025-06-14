#!/bin/bash

# Setup script for Git hooks and commitlint

echo "Setting up Git hooks and commitlint..."

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Create commit-msg hook for commitlint
cat > .git/hooks/commit-msg << 'EOF'
#!/bin/sh
# Commitlint hook to validate commit messages

# Check if npx is available
if command -v npx >/dev/null 2>&1; then
    npx commitlint --edit "$1"
else
    echo "Warning: npx not found. Skipping commit message validation."
    echo "Please install Node.js and npm to enable commit message validation."
fi
EOF

# Make the hook executable
chmod +x .git/hooks/commit-msg

# Set up commit message template
git config commit.template .gitmessage

echo "✅ Git hooks and commitlint setup complete!"
echo ""
echo "📝 Commit message format:"
echo "   <type>(<scope>): <subject>"
echo ""
echo "🔧 Available types:"
echo "   feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
echo ""
echo "💡 Examples:"
echo "   feat(api): add user authentication endpoint"
echo "   fix(ui): resolve navigation menu overflow issue"  
echo "   docs: update installation instructions"
echo ""
echo "🚀 To install dependencies and enable validation:"
echo "   npm install"

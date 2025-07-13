#!/bin/bash

# Setup script for Sting Ray environment configuration
echo "Setting up Sting Ray environment configuration..."
echo "================================================"

# Check if .env file already exists
if [ -f .env ]; then
    echo "Warning: .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. Your existing .env file was preserved."
        exit 0
    fi
fi

# Copy env.example to .env
if [ -f env.example ]; then
    cp env.example .env
    echo "✓ Created .env file from env.example"
    echo ""
    echo "Please review and modify the .env file as needed:"
    echo "  - Update database credentials if different from defaults"
    echo "  - Change test passwords for security"
    echo ""
    echo "You can now run the test script: ./test_user_system.sh"
else
    echo "✗ Error: env.example file not found!"
    exit 1
fi 
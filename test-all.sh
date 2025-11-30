#!/bin/bash

# Master Test Script - Runs all tests

set -e

echo "=========================================="
echo "Yukti Complete Test Suite"
echo "=========================================="
echo ""

# Step 1: Setup
echo "Step 1: Running setup..."
./test-setup.sh
if [ $? -ne 0 ]; then
    echo "Setup failed. Exiting."
    exit 1
fi

# Step 2: Check if backend is running
echo ""
echo "Step 2: Checking backend..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Backend is running"
else
    echo "✗ Backend is not running"
    echo ""
    echo "Please start the backend in another terminal:"
    echo "  ./yukti-api"
    echo ""
    echo "Then run this script again."
    exit 1
fi

# Step 3: Run backend tests
echo ""
echo "Step 3: Running backend tests..."
./test-backend.sh
BACKEND_EXIT=$?

# Summary
echo ""
echo "=========================================="
echo "Complete Test Suite Summary"
echo "=========================================="

if [ $BACKEND_EXIT -eq 0 ]; then
    echo "✓ Backend tests: PASSED"
else
    echo "✗ Backend tests: FAILED"
fi

echo ""
if [ $BACKEND_EXIT -eq 0 ]; then
    echo "✓ All tests passed!"
    echo ""
    echo "Next steps:"
    echo "  1. Test frontend: cd frontend && npm start"
    echo "  2. Review: cat HIGH_PRIORITY_COMPLETE.md"
    echo "  3. Deploy to AWS when ready"
    exit 0
else
    echo "✗ Some tests failed. Review output above."
    exit 1
fi

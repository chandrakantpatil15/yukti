#!/bin/bash

# Yukti FinOps - Cleanup Obsolete Files
# This script removes outdated documentation and backup files
# Run with: bash scripts/cleanup-obsolete-files.sh

set -e

echo "🧹 Yukti FinOps - Cleanup Obsolete Files"
echo "========================================"
echo ""

# Create archive directory for backup
ARCHIVE_DIR="archive/obsolete-$(date +%Y%m%d)"
mkdir -p "$ARCHIVE_DIR"

echo "📦 Archiving obsolete files to: $ARCHIVE_DIR"
echo ""

# Function to archive file
archive_file() {
    if [ -f "$1" ]; then
        echo "  ✓ Archiving: $1"
        mv "$1" "$ARCHIVE_DIR/"
    fi
}

# Duplicate/Old Documentation
echo "📄 Removing duplicate documentation..."
archive_file "PROJECT_HISTORY_COMPACT.md"
archive_file "PROJECT_HISTORY_PHASE10.md"
archive_file "CONVERSATION_PROGRESS.md"
archive_file "PROGRESS_SUMMARY.md"
archive_file "CHANGES_SUMMARY.md"
archive_file "IMPLEMENTATION_SUMMARY.md"
archive_file "IMPLEMENTATION_SUMMARY_RBAC.md"

# Old Session Summaries
echo "📄 Removing old session summaries..."
archive_file "SESSION_13_SUMMARY.md"
archive_file "SESSION_15_SUMMARY.md"
archive_file "DASHBOARD_FIX_COMPLETE.md"
archive_file "EXTERNAL_ID_REMOVED.md"
archive_file "ONBOARDING_FLOW_FIXED.md"
archive_file "SES_FIX_APPLIED.md"

# Old Implementation Docs
echo "📄 Removing old implementation docs..."
archive_file "WEEK2_DAY3-4_COMPLETE.md"
archive_file "WEEK4_DAY1-4_COMPLETE.md"
archive_file "WEEK5_DAY1_COMPLETE.md"
archive_file "WEEK5_DAY2_COMPLETE.md"
archive_file "WEEK5_DAY3_COMPLETE.md"

# Old Testing Docs
echo "📄 Removing old testing docs..."
archive_file "QUICK_TEST.md"
archive_file "QUICK_TEST_GUIDE.md"
archive_file "TEST_GUIDE.md"
archive_file "UI_TESTING_GUIDE.md"
archive_file "LOCAL_TESTING_GUIDE.md"

# Old Architecture Docs
echo "📄 Removing old architecture docs..."
archive_file "GO_ONLY_ARCHITECTURE.md"
archive_file "enterprise-saas-architecture.md"
archive_file "ULTIMATE_MULTICLOUD_ARCHITECTURE.md"

# Old Port Docs
echo "📄 Removing old port docs..."
archive_file "PORT_ALLOCATION.md"
archive_file "PORT_MANAGEMENT.md"
archive_file "PORT_FLOW_DIAGRAM.md"
archive_file "PORTS_EXPLAINED.md"
archive_file "HOW_TO_CHANGE_PORTS.md"

# Old Quick Start Docs
echo "📄 Removing old quick start docs..."
archive_file "QUICK_START.md"
archive_file "QUICK_REFERENCE.md"
archive_file "E2E_QUICK_START.md"
archive_file "DOCKER_QUICK_REFERENCE.md"

# Old Security Docs
echo "📄 Removing old security docs..."
archive_file "SECURITY_STATUS.md"
archive_file "CRITICAL_FIXES_APPLIED.md"

# Old Deployment Docs
echo "📄 Removing old deployment docs..."
archive_file "DEPLOYMENT_SUMMARY.md"

# Old Feature Docs
echo "📄 Removing old feature docs..."
archive_file "SMOOTH_ONBOARDING.md"
archive_file "FRONTEND_DYNAMIC_PAGES.md"
archive_file "MARKETING_SITE_COMPLETE.md"

# Old Roadmap Docs
echo "📄 Removing old roadmap docs..."
archive_file "FEATURE_ROADMAP_2025.md"
archive_file "NEXT_STEPS_COMPLETE.md"

# Old Backup Files
echo "🗄️  Removing backup files..."
archive_file "docker-compose.yml.bak"
archive_file "prometheus.yml.bak"
archive_file "cmd/full-aws-pricing-sync.go.bak"
archive_file "cmd/full-aws-pricing-sync.go.old"

# Old Log Files
echo "📋 Removing old log files..."
archive_file "automated_changes_20251107_171246.log"
archive_file "automated_changes_20251107_171634.log"
archive_file "server.log"

# Old PID Files
echo "🔧 Removing old PID files..."
rm -f .api.pid .health.pid .multicloud.pid .ui.pid

# Old Test Files
echo "🧪 Removing old test files..."
rm -f tests.test health-monitor

echo ""
echo "✅ Cleanup complete!"
echo ""
echo "📊 Summary:"
echo "  - Archived files: $(ls -1 $ARCHIVE_DIR | wc -l)"
echo "  - Archive location: $ARCHIVE_DIR"
echo ""
echo "💡 To restore files: mv $ARCHIVE_DIR/* ."
echo ""

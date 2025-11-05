#!/bin/bash

# Health Check and Kill Switch Control Script

HEALTH_URL="http://localhost:8081"

echo "🏥 YUKTI FINOPS - HEALTH & KILL SWITCH CONTROL"
echo "=============================================="

# Function to check health
check_health() {
    echo "🔍 Checking system health..."
    curl -s "$HEALTH_URL/health" | jq '.' || echo "❌ Health check failed"
}

# Function to get kill switch status
get_kill_switch_status() {
    echo "🔍 Checking kill switch status..."
    curl -s "$HEALTH_URL/kill-switch" | jq '.' || echo "❌ Kill switch check failed"
}

# Function to enable kill switch
enable_kill_switch() {
    local reason="${1:-Manual activation}"
    echo "🚨 Enabling kill switch: $reason"
    curl -s -X POST "$HEALTH_URL/kill-switch" \
        -H "Content-Type: application/json" \
        -d "{\"enable\": true, \"reason\": \"$reason\"}" | jq '.'
}

# Function to disable kill switch
disable_kill_switch() {
    echo "✅ Disabling kill switch..."
    curl -s -X POST "$HEALTH_URL/kill-switch" \
        -H "Content-Type: application/json" \
        -d "{\"enable\": false, \"reason\": \"Manual deactivation\"}" | jq '.'
}

# Function to trigger emergency stop
emergency_stop() {
    echo "🚨 TRIGGERING EMERGENCY STOP..."
    curl -s -X POST "$HEALTH_URL/emergency-stop" | jq '.'
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  health              - Check system health"
    echo "  kill-status         - Check kill switch status"
    echo "  kill-enable [reason] - Enable kill switch"
    echo "  kill-disable        - Disable kill switch"
    echo "  emergency-stop      - Emergency stop all instances"
    echo "  monitor             - Continuous health monitoring"
    echo ""
    echo "Examples:"
    echo "  $0 health"
    echo "  $0 kill-enable \"Cost limit exceeded\""
    echo "  $0 emergency-stop"
}

# Function for continuous monitoring
monitor() {
    echo "📊 Starting continuous health monitoring (Ctrl+C to stop)..."
    while true; do
        echo ""
        echo "$(date '+%Y-%m-%d %H:%M:%S') - Health Check"
        echo "----------------------------------------"
        check_health
        echo ""
        sleep 30
    done
}

# Main command handling
case "${1:-health}" in
    "health")
        check_health
        ;;
    "kill-status")
        get_kill_switch_status
        ;;
    "kill-enable")
        enable_kill_switch "$2"
        ;;
    "kill-disable")
        disable_kill_switch
        ;;
    "emergency-stop")
        emergency_stop
        ;;
    "monitor")
        monitor
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo "❌ Unknown command: $1"
        show_usage
        exit 1
        ;;
esac
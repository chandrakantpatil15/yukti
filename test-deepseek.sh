#!/bin/bash
echo "=== STOPPING JENKINS ==="

# Kill any process on port 8080
sudo kill -9 $(sudo lsof -t -i:8080) 2>/dev/null || true

# Stop Jenkins services
sudo launchctl unload /Library/LaunchDaemons/org.jenkins-ci.plist 2>/dev/null || true
brew services stop jenkins 2>/dev/null || true
pkill -f jenkins 2>/dev/null || true

echo "=== REMOVING JENKINS FILES ==="

# Remove installation files
sudo rm -rf /Applications/Jenkins 2>/dev/null || true
sudo rm -rf /usr/local/var/lib/jenkins 2>/dev/null || true
sudo rm -rf /var/lib/jenkins 2>/dev/null || true
sudo rm -rf /var/log/jenkins 2>/dev/null || true

# Remove config files
rm -rf ~/.jenkins 2>/dev/null || true
sudo rm -f /Library/Preferences/org.jenkins-ci.plist 2>/dev/null || true

# Remove from Docker
docker stop jenkins 2>/dev/null || true
docker rm jenkins 2>/dev/null || true
docker rmi jenkins/jenkins:lts 2>/dev/null || true

echo "=== VERIFYING PORT 8080 IS FREE ==="
sleep 2
if sudo lsof -i :8080 > /dev/null; then
    echo "❌ Port 8080 is still in use:"
    sudo lsof -i :8080
    echo "Run this to kill: sudo kill -9 \$(sudo lsof -t -i:8080)"
else
    echo "✅ Port 8080 is now free!"
fi

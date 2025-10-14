#!/bin/bash
export JAVA_HOME=$(/usr/libexec/java_home -v 17)
echo "Starting Spring Boot application..."
mvn spring-boot:run &
APP_PID=$!
sleep 5
echo "Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Health check failed"
kill $APP_PID
echo "Test completed"
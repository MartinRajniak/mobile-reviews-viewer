#!/bin/bash
# Script to run the application with proper shutdown logging
# Gradle's run task cannot show shutdown logs properly due to daemon process management

set -e

echo "Building application..."
./gradlew build -q

echo "Starting application (Press Ctrl+C to stop)..."
java -jar build/libs/ktor-sample-all.jar
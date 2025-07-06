#!/bin/bash

set -e  # Exit on any error

# Runs docker compose test commands from the project root using the shared bin script.
# Uses a subshell so the working directory change doesn't affect the parent script.
docker_compose() {
    (cd ../../ && ./bin/docker-compose-test.sh "$@")
}

# Cleanup function that will be called on script exit or interrupt
cleanup() {
    echo ""
    echo "═══════════════════════════════════════"
    echo "🧹 CLEANUP (Force quit)"
    echo "═══════════════════════════════════════"
    echo ""
    docker_compose down || true
    echo ""
    echo "✅ Cleanup completed"
    exit 0
}

# Set up signal traps to catch Ctrl+C (SIGINT) and SIGTERM
trap cleanup SIGINT SIGTERM

# Clean environment and start docker compose with test configuration
echo ""
echo "═══════════════════════════════════════"
echo "🐳 DOCKER SETUP"
echo "═══════════════════════════════════════"
docker_compose up -d --build

# Wait for postgres service to be ready
echo ""
echo "⏳ Waiting for postgres to be healthy..."
while ! docker_compose ps postgres --format "table {{.Service}}\t{{.Status}}" | grep -q "healthy"; do
    echo "   Still waiting..."
    sleep 2
done
echo "✅ Postgres is ready!"

# Run database migrations
echo ""
echo "═══════════════════════════════════════"
echo "📄 DATABASE MIGRATIONS"
echo "═══════════════════════════════════════"
echo ""
make ENV_FILE=".env.test" pg-goose up

echo ""
echo "═══════════════════════════════════════"
echo "🧪 RUNNING TESTS"
echo "═══════════════════════════════════════"
echo ""
set +e  # Don't exit on test failure, we want to clean up first
go test ./tests/web -v
test_status=$?
set -e

echo ""
echo "═══════════════════════════════════════"
echo "🧹 CLEANUP"
echo "═══════════════════════════════════════"
docker_compose down
echo ""

# Clear the trap since we're doing normal cleanup
trap - SIGINT SIGTERM

echo ""
echo "═══════════════════════════════════════"
# Exit with the test status
if [ $test_status -eq 0 ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "💥 Tests failed!"
    exit 1
fi

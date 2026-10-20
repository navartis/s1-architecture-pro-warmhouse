#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

if docker compose version > /dev/null 2>&1; then
  COMPOSE="docker compose"
elif command -v docker-compose > /dev/null 2>&1; then
  COMPOSE="docker-compose"
else
  echo "Error: Docker Compose не найден (ни 'docker compose', ни 'docker-compose')."
  exit 1
fi

echo "Starting the Smart Home stack..."
$COMPOSE up -d --build

wait_for() {
  local name="$1"
  local url="$2"
  local attempts=60

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$url" > /dev/null 2>&1; then
      echo "  $name ready"
      return 0
    fi
    sleep 1
  done

  echo "  $name did not respond within ${attempts}s: $url"
  return 1
}

echo "Waiting for services to respond..."
failed=0
wait_for "monolith        (8080)" http://localhost:8080/health || failed=1
wait_for "temperature-api (8081)" http://localhost:8081/health || failed=1
wait_for "device-service  (8082)" http://localhost:8082/health || failed=1
wait_for "telemetry-service (8083)" http://localhost:8083/health || failed=1

if [ "$failed" -ne 0 ]; then
  echo ""
  echo "Some services are not responding. Check the logs:"
  echo "  $COMPOSE logs -f"
  exit 1
fi

echo ""
echo "All services are up:"
echo "  monolith           http://localhost:8080"
echo "  temperature-api    http://localhost:8081"
echo "  device-service     http://localhost:8082"
echo "  telemetry-service  http://localhost:8083"
echo ""
echo "Examples:"
echo "  curl localhost:8080/api/v1/sensors"
echo "  curl localhost:8080/api/v1/devices"
echo "  curl 'localhost:8080/api/v1/telemetry?deviceId=1'"

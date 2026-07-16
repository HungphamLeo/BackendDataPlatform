#!/usr/bin/env bash
# kafka-topics.sh — create all data-platform Kafka topics
# Usage: ./kafka-topics.sh [KAFKA_BOOTSTRAP_SERVER]

set -euo pipefail

BOOTSTRAP="${1:-localhost:9092}"
PARTITIONS="${PARTITIONS:-4}"
REPLICATION="${REPLICATION:-1}"

TOPICS=(
  "market.ticker"
  "market.orderbook"
  "market.kline.1m"
  "market.kline.5m"
  "market.kline.1h"
)

for topic in "${TOPICS[@]}"; do
  echo "Creating topic: $topic"
  kafka-topics.sh \
    --bootstrap-server "$BOOTSTRAP" \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions "$PARTITIONS" \
    --replication-factor "$REPLICATION"
done

echo "All topics created."

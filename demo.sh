#!/bin/sh

# Sample images
IMAGE_A="example_a.webp"
IMAGE_B="example_b.jfif"
URL="http://localhost:8080/upload"
API_KEY="secret"

# Alternate images
for i in $(seq 1 5); do
  echo "[$(date)] Sending $IMAGE_A ($i/5)"
  curl -s -F "image=@$IMAGE_A" -H "X-API-Key: $API_KEY" "$URL" && echo "✅"
  sleep 5

  echo "[$(date)] Sending $IMAGE_B ($i/5)"
  curl -s -F "image=@$IMAGE_B" -H "X-API-Key: $API_KEY" "$URL" && echo "✅"
  sleep 5
done

echo "✅ Done."

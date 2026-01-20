#!/bin/sh

# Sample images
IMAGE_A="example_a.webp"
IMAGE_B="example_b.jfif"
URL="http://localhost:8080/demo/upload"
ADMIN_PASSWORD="secret"

# Alternate images
for i in $(seq 1 5); do
  echo "[$(date)] Sending $IMAGE_A ($i/5)"
  # using basic auth for admin endpoint
  curl -s -F "image=@$IMAGE_A" -u "admin:$ADMIN_PASSWORD" "$URL" && echo "✅"
  sleep 5

  echo "[$(date)] Sending $IMAGE_B ($i/5)"
  curl -s -F "image=@$IMAGE_B" -u "admin:$ADMIN_PASSWORD" "$URL" && echo "✅"
  sleep 5
done

echo "✅ Done."

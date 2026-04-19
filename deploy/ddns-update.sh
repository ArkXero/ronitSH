#!/bin/bash
# Cloudflare DDNS updater for ronit.sh
# Run every 5 minutes via cron:
#   */5 * * * * /opt/termfolio/ddns-update.sh >> /var/log/termfolio-ddns.log 2>&1
#
# Required environment variables (set in /etc/environment or a sourced file):
#   CF_API_TOKEN  -- Cloudflare API token with DNS:Edit on the zone
#   CF_ZONE_ID    -- zone ID for ronit.sh
#   CF_RECORD_ID  -- DNS record ID for the A record

set -euo pipefail

CF_API="https://api.cloudflare.com/client/v4"
RECORD_NAME="ronit.sh"

if [[ -z "${CF_API_TOKEN:-}" || -z "${CF_ZONE_ID:-}" || -z "${CF_RECORD_ID:-}" ]]; then
  echo "$(date): ERROR: CF_API_TOKEN, CF_ZONE_ID, CF_RECORD_ID must be set"
  exit 1
fi

# Get current public IP.
CURRENT_IP=$(curl -sf https://api.ipify.org)
if [[ -z "$CURRENT_IP" ]]; then
  echo "$(date): ERROR: could not fetch public IP"
  exit 1
fi

# Get the IP currently in DNS.
RECORD_IP=$(curl -sf \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  "$CF_API/zones/$CF_ZONE_ID/dns_records/$CF_RECORD_ID" \
  | grep -o '"content":"[^"]*"' | cut -d'"' -f4)

if [[ "$CURRENT_IP" == "$RECORD_IP" ]]; then
  # No change needed.
  exit 0
fi

# Update the record.
echo "$(date): IP changed $RECORD_IP -> $CURRENT_IP, updating DNS..."
curl -sf -X PUT \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -H "Content-Type: application/json" \
  --data "{\"type\":\"A\",\"name\":\"$RECORD_NAME\",\"content\":\"$CURRENT_IP\",\"ttl\":300,\"proxied\":false}" \
  "$CF_API/zones/$CF_ZONE_ID/dns_records/$CF_RECORD_ID" > /dev/null

echo "$(date): DNS updated to $CURRENT_IP"

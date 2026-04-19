#!/bin/bash

# Скрипт для тестирования REST API
# Требует: запущенный сервер и подключенную зарядную станцию

BASE_URL="http://localhost:8000"
CHARGE_POINT="CP001"
CONNECTOR=1
USER_ID="test-user-123"

echo "═══════════════════════════════════════════════════════"
echo "    OCPP REST API Test Script"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "Prerequisites:"
echo "  1. OCPP server running: go run cmd/server/main.go"
echo "  2. Charge point connected: WS_URL=ws://localhost:9005/CP001 npx tsx index_16.ts"
echo ""
echo "Press Enter to start testing..."
read

echo ""
echo "───────────────────────────────────────────────────────"
echo "1. Health Check"
echo "───────────────────────────────────────────────────────"
curl -s "$BASE_URL/health" | jq
echo ""

echo ""
echo "───────────────────────────────────────────────────────"
echo "2. Start Charging Session"
echo "───────────────────────────────────────────────────────"
echo "Request:"
cat <<EOF | jq
{
  "chargePointId": "$CHARGE_POINT",
  "connectorId": $CONNECTOR,
  "userId": "$USER_ID",
  "amountPaid": 25.00
}
EOF

RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"chargePointId\": \"$CHARGE_POINT\",
    \"connectorId\": $CONNECTOR,
    \"userId\": \"$USER_ID\",
    \"amountPaid\": 25.00
  }")

echo ""
echo "Response:"
echo "$RESPONSE" | jq

# Extract transaction ID
TX_ID=$(echo "$RESPONSE" | jq -r '.transactionId')

if [ "$TX_ID" = "null" ] || [ -z "$TX_ID" ]; then
    echo ""
    echo "❌ Failed to start session. Check if charge point is connected."
    echo "   Run: WS_URL=ws://localhost:9005/$CHARGE_POINT npx tsx index_16.ts"
    exit 1
fi

echo ""
echo "✓ Session started: Transaction ID = $TX_ID"
echo ""

echo "Press Enter to check energy consumption..."
read

echo ""
echo "───────────────────────────────────────────────────────"
echo "3. Get Energy Consumption"
echo "───────────────────────────────────────────────────────"
curl -s "$BASE_URL/sessions/$TX_ID/energy?chargePointId=$CHARGE_POINT" | jq
echo ""

echo "Press Enter to stop the session..."
read

echo ""
echo "───────────────────────────────────────────────────────"
echo "4. Stop Charging Session"
echo "───────────────────────────────────────────────────────"
curl -s -X POST "$BASE_URL/sessions/$TX_ID/stop?chargePointId=$CHARGE_POINT" | jq
echo ""

echo ""
echo "═══════════════════════════════════════════════════════"
echo "    Test Complete!"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "Summary:"
echo "  ✓ Health check"
echo "  ✓ Started session (TX: $TX_ID)"
echo "  ✓ Retrieved energy consumption"
echo "  ✓ Stopped session"
echo ""

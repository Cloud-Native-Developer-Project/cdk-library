#!/bin/bash

# Script to retrieve the ngrok public URL from the API

echo "🔍 Fetching ngrok tunnel URL..."
echo ""

# Wait for ngrok to be ready
sleep 2

# Get the public URL from ngrok API
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"https://[^"]*' | grep -o 'https://[^"]*' | head -1)

if [ -z "$NGROK_URL" ]; then
    echo "❌ Could not retrieve ngrok URL"
    echo ""
    echo "Make sure:"
    echo "  1. docker-compose is running"
    echo "  2. ngrok container is up"
    echo "  3. Wait a few seconds and try again"
    echo ""
    echo "You can also check manually at: http://localhost:4040"
    exit 1
fi

echo "✅ ngrok tunnel is ready!"
echo ""
echo "📍 Public URL: $NGROK_URL"
echo "🌐 Web UI: http://localhost:4040"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Copy this URL for your Lambda configuration:"
echo "   ${NGROK_URL}/webhook/addi-csv"
echo ""
echo "2. Update stacks/addi/AddiStack.go:"
echo "   \"WEBHOOK_URL_OVERRIDE\": jsii.String(\"${NGROK_URL}/webhook/addi-csv\")"
echo ""
echo "3. Deploy Lambda:"
echo "   cdk deploy AddiStack --require-approval never"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

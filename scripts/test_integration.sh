#!/bin/bash

# Integration test for Seg evolution pipeline
# This script should be run from the project root.

# 1. Cleanup old processes
echo "Cleaning up..."
pkill -f "go run cmd/server/main.go" || true

# 2. Start server in background
echo "Starting server..."
go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to build and start
echo "Waiting for engine initialization (10s)..."
sleep 10

# 3. Test 1: Segment an unknown word
# We use a rare string that is likely not in the default dictionary
TEST_WORD="极客之选"
echo "--- 1. Testing unknown term: '$TEST_WORD' ---"
curl -s -X POST -H "Content-Type: application/json" \
     -d "{\"text\": \"这是我们的极客之选频道\", \"algorithm\": \"hybrid\"}" \
     http://localhost:8080/segment | python3 -m json.tool || curl -s -X POST -H "Content-Type: application/json" -d "{\"text\": \"这是我们的极客之选频道\", \"algorithm\": \"hybrid\"}" http://localhost:8080/segment
echo ""

# 4. Feedback: Teach the server the new term
echo "--- 2. Sending Correction Feedback: '$TEST_WORD' ---"
curl -s "http://localhost:8080/feedback?word=$TEST_WORD"
echo ""

# 5. Wait for the background optimization pipeline (Discovery -> Train -> Reload)
echo "--- 3. Waiting 25s for evolution pipeline to complete... ---"
sleep 25

# 6. Test 2: Verify the term is now recognized
echo "--- 4. Verifying evolved model: '$TEST_WORD' ---"
curl -s -X POST -H "Content-Type: application/json" \
     -d "{\"text\": \"这是我们的极客之选频道\", \"algorithm\": \"hybrid\"}" \
     http://localhost:8080/segment | python3 -m json.tool || curl -s -X POST -H "Content-Type: application/json" -d "{\"text\": \"这是我们的极客之选频道\", \"algorithm\": \"hybrid\"}" http://localhost:8080/segment
echo ""

# 7. Final Cleanup
kill $SERVER_PID
echo "Server stopped."
echo "--- Test Complete ---"
echo "Check 'server.log' for detailed pipeline logs."

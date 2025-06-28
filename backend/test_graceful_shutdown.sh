#!/bin/bash

echo "Starting server..."
./server &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
echo "Waiting 3 seconds for server to fully start..."
sleep 3

echo "Making a request to test server is running..."
curl -s http://localhost:8080/api/health

echo ""
echo "Sending SIGTERM to gracefully shutdown server..."
kill -TERM $SERVER_PID

echo "Waiting for server to shutdown gracefully..."
wait $SERVER_PID

echo "Server has been gracefully shutdown!" 
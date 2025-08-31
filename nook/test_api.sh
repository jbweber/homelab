#!/bin/bash

# Test script for Nook Machine API
echo "Testing Nook Machine API..."

# Start the server in background
echo "Starting server..."
./nook &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Testing endpoints..."

# Test 1: List machines (should be empty)
echo "1. GET /api/v0/machines (empty list)"
curl -s -X GET http://localhost:8080/api/v0/machines
echo -e "\n"

# Test 2: Create a machine
echo "2. POST /api/v0/machines (create machine)"
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "test-server", "ipv4": "192.168.1.100"}')
echo $CREATE_RESPONSE

# Extract machine ID from response
MACHINE_ID=$(echo $CREATE_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated machine with ID: $MACHINE_ID\n"

# Test 3: List machines again (should have one)
echo "3. GET /api/v0/machines (should have one machine)"
curl -s -X GET http://localhost:8080/api/v0/machines
echo -e "\n"

# Test 4: Get specific machine
echo "4. GET /api/v0/machines/$MACHINE_ID"
curl -s -X GET http://localhost:8080/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Test 5: Get machine by name
echo "5. GET /api/v0/machines/name/test-server"
curl -s -X GET http://localhost:8080/api/v0/machines/name/test-server
echo -e "\n"

# Test 6: Get machine by IPv4
echo "6. GET /api/v0/machines/ipv4/192.168.1.100"
curl -s -X GET http://localhost:8080/api/v0/machines/ipv4/192.168.1.100
echo -e "\n"

# Test 7: Create another machine
echo "7. POST /api/v0/machines (create another machine)"
curl -s -X POST http://localhost:8080/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "web-server", "ipv4": "192.168.1.101"}'
echo -e "\n"

# Test 8: List all machines
echo "8. GET /api/v0/machines (should have two machines)"
curl -s -X GET http://localhost:8080/api/v0/machines
echo -e "\n"

# Test 9: Delete a machine
echo "9. DELETE /api/v0/machines/$MACHINE_ID"
curl -s -X DELETE http://localhost:8080/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Test 10: List machines after deletion
echo "10. GET /api/v0/machines (should have one machine left)"
curl -s -X GET http://localhost:8080/api/v0/machines
echo -e "\n"

# Test 11: Try to get deleted machine (should 404)
echo "11. GET /api/v0/machines/$MACHINE_ID (should 404)"
curl -s -X GET http://localhost:8080/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Stop the server
echo "Stopping server..."
kill $SERVER_PID

echo "API testing complete!"

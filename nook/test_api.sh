#!/bin/bash

# Test script for Nook Machine API

# Cleanup function to kill the server process and clean up test database
cleanup() {
    echo "Cleaning up..."
    if [ ! -z "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID
        # Wait for process to actually stop
        wait $SERVER_PID 2>/dev/null
        echo "Server stopped."
    else
        echo "Server process not running or already stopped."
    fi
    
    # Clean up test database
    if [ -f "./test_nook.db" ]; then
        echo "Removing test database..."
        rm -f ./test_nook.db
        echo "Test database removed."
    fi
}

# Set up traps for cleanup on exit, interrupt, or termination
trap cleanup EXIT
trap cleanup INT
trap cleanup TERM

echo "Testing Nook Machine API..."

# Start the server in background
echo "Starting server..."
./nook server --db-path ./test_nook.db --port 8081 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Verify server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "ERROR: Server failed to start!"
    exit 1
fi

echo "Server started successfully (PID: $SERVER_PID)"

# Detect the IP address as seen by the API
CLIENT_IP=$(hostname -I | awk '{print $1}')
echo "Detected client IP: $CLIENT_IP"

echo "Testing endpoints..."

# Test 1: List machines (should be empty)
echo "1. GET /api/v0/machines (empty list)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

echo "2. POST /api/v0/machines (create machine)"
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-server\", \"hostname\": \"test-host\", \"ipv4\": \"$CLIENT_IP\"}")
echo $CREATE_RESPONSE

# Extract machine ID from response
MACHINE_ID=$(echo $CREATE_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated machine with ID: $MACHINE_ID\n"

# Test 3: List machines again (should have one)
echo "3. GET /api/v0/machines (should have one machine)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

# Test 4: Get specific machine
echo "4. GET /api/v0/machines/$MACHINE_ID"
curl -s -X GET http://localhost:8081/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Test 5: Get machine by name
echo "5. GET /api/v0/machines/name/test-server"
curl -s -X GET http://localhost:8081/api/v0/machines/name/test-server
echo -e "\n"

# Test 6: Get machine by IPv4
echo "6. GET /api/v0/machines/ipv4/$CLIENT_IP"
curl -s -X GET http://localhost:8081/api/v0/machines/ipv4/$CLIENT_IP
echo -e "\n"

echo "7. POST /api/v0/machines (create another machine)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"web-server\", \"hostname\": \"web-host\", \"ipv4\": \"192.168.1.101\"}"
echo -e "\n"

# Test 12: Metadata endpoints
echo "12. GET /meta-data (should return metadata for test-host)"
curl -s -X GET http://localhost:8081/meta-data --header "X-Forwarded-For: $CLIENT_IP"
echo -e "\n"

echo "13. GET /user-data"
curl -s -X GET http://localhost:8081/user-data
echo -e "\n"

echo "14. GET /vendor-data"
curl -s -X GET http://localhost:8081/vendor-data
echo -e "\n"

echo "15. GET /network-config"
curl -s -X GET http://localhost:8081/network-config
echo -e "\n"

# Test 8: List all machines
echo "8. GET /api/v0/machines (should have two machines)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

# Test 9: Delete a machine
echo "9. DELETE /api/v0/machines/$MACHINE_ID"
curl -s -X DELETE http://localhost:8081/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Test 10: List machines after deletion
echo "10. GET /api/v0/machines (should have one machine left)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"


# Test 16: Create machine with invalid IPv4
echo "16. POST /api/v0/machines (invalid IPv4)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "bad-ip", "hostname": "bad-host", "ipv4": "not-an-ip"}'
echo -e "\n"

# Test 17: Create machine with duplicate IPv4
echo "17. POST /api/v0/machines (duplicate IPv4)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"dup-server\", \"hostname\": \"dup-host\", \"ipv4\": \"$CLIENT_IP\"}"
echo -e "\n"

# Test 18: Create machine with missing fields
echo "18. POST /api/v0/machines (missing fields)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "", "hostname": "", "ipv4": ""}'
echo -e "\n"

# Test 19: GET /meta-data for non-existent IP
echo "19. GET /meta-data (non-existent IP)"
curl -s -X GET http://localhost:8081/meta-data --header "X-Forwarded-For: 203.0.113.99"
echo -e "\n"

# Test 20: GET /api/v0/machines/99999 (invalid ID)"
echo "20. GET /api/v0/machines/99999 (invalid ID)"
curl -s -X GET http://localhost:8081/api/v0/machines/99999
echo -e "\n"

# Test 21: DELETE /api/v0/machines/99999 (invalid ID)"
echo "21. DELETE /api/v0/machines/99999 (invalid ID)"
curl -s -X DELETE http://localhost:8081/api/v0/machines/99999
echo -e "\n"

echo "API testing complete!"
# Cleanup will be handled automatically by the trap

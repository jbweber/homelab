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

# Network endpoint tests
echo "22. GET /api/v0/networks (empty list)"
curl -s -X GET http://localhost:8081/api/v0/networks
echo -e "\n"

echo "23. POST /api/v0/networks (create network)"
NETWORK_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/networks \
  -H "Content-Type: application/json" \
  -d '{"name": "br0", "bridge": "br0", "subnet": "192.168.1.0/24", "gateway": "192.168.1.1", "dns_servers": "8.8.8.8,1.1.1.1", "description": "Main network bridge"}')
echo $NETWORK_RESPONSE

# Extract network ID from response
NETWORK_ID=$(echo $NETWORK_RESPONSE | grep -o '"ID":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated network with ID: $NETWORK_ID\n"

echo "24. GET /api/v0/networks (should have one network)"
curl -s -X GET http://localhost:8081/api/v0/networks
echo -e "\n"

echo "25. GET /api/v0/networks/$NETWORK_ID"
curl -s -X GET http://localhost:8081/api/v0/networks/$NETWORK_ID
echo -e "\n"

echo "26. PATCH /api/v0/networks/$NETWORK_ID (update network)"
curl -s -X PATCH http://localhost:8081/api/v0/networks/$NETWORK_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "br0", "bridge": "br0", "subnet": "192.168.1.0/24", "gateway": "192.168.1.1", "dns_servers": "8.8.8.8,1.1.1.1", "description": "Updated main network bridge"}'
echo -e "\n"

echo "27. POST /api/v0/networks/$NETWORK_ID/dhcp (create DHCP range)"
DHCP_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/networks/$NETWORK_ID/dhcp \
  -H "Content-Type: application/json" \
  -d '{"StartIP": "192.168.1.100", "EndIP": "192.168.1.200", "LeaseTime": "12h"}')
echo $DHCP_RESPONSE

# Extract DHCP range ID from response
DHCP_ID=$(echo $DHCP_RESPONSE | grep -o '"ID":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated DHCP range with ID: $DHCP_ID\n"

echo "28. GET /api/v0/networks/$NETWORK_ID/dhcp (list DHCP ranges)"
curl -s -X GET http://localhost:8081/api/v0/networks/$NETWORK_ID/dhcp
echo -e "\n"

echo "29. DELETE /api/v0/networks/dhcp/$DHCP_ID (delete DHCP range)"
curl -s -X DELETE http://localhost:8081/api/v0/networks/dhcp/$DHCP_ID
echo -e "
"
echo -e "\n"

echo "30. GET /api/v0/networks/$NETWORK_ID/dhcp (should be empty)"
curl -s -X GET http://localhost:8081/api/v0/networks/$NETWORK_ID/dhcp
echo -e "\n"

# Test 31: Metadata endpoints
echo "31. GET /meta-data (should return metadata for test-host)"
curl -s -X GET http://localhost:8081/meta-data --header "X-Forwarded-For: $CLIENT_IP"
echo -e "\n"

echo "32. GET /user-data"
curl -s -X GET http://localhost:8081/user-data
echo -e "\n"

echo "33. GET /vendor-data"
curl -s -X GET http://localhost:8081/vendor-data
echo -e "\n"

echo "34. GET /network-config"
curl -s -X GET http://localhost:8081/network-config
echo -e "\n"

# Test 35: List all machines
echo "35. GET /api/v0/machines (should have two machines)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

# Test 36: Delete a machine
echo "36. DELETE /api/v0/machines/$MACHINE_ID"
curl -s -X DELETE http://localhost:8081/api/v0/machines/$MACHINE_ID
echo -e "\n"

# Test 37: List machines after deletion
echo "37. GET /api/v0/machines (should have one machine left)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"


# Test 38: Create machine with invalid IPv4
echo "38. POST /api/v0/machines (invalid IPv4)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "bad-ip", "hostname": "bad-host", "ipv4": "not-an-ip"}'
echo -e "\n"

# Test 39: Create machine with duplicate IPv4
echo "39. POST /api/v0/machines (duplicate IPv4)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"dup-server\", \"hostname\": \"dup-host\", \"ipv4\": \"$CLIENT_IP\"}"
echo -e "\n"

# Test 40: Create machine with missing fields
echo "40. POST /api/v0/machines (missing fields)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "", "hostname": "", "ipv4": ""}'
echo -e "\n"

# Test 41: GET /meta-data for non-existent IP
echo "41. GET /meta-data (non-existent IP)"
curl -s -X GET http://localhost:8081/meta-data --header "X-Forwarded-For: 203.0.113.99"
echo -e "\n"

# Test 42: GET /api/v0/machines/99999 (invalid ID)"
echo "42. GET /api/v0/machines/99999 (invalid ID)"
curl -s -X GET http://localhost:8081/api/v0/machines/99999
echo -e "\n"

# Test 43: DELETE /api/v0/machines/99999 (invalid ID)"
echo "43. DELETE /api/v0/machines/99999 (invalid ID)"
curl -s -X DELETE http://localhost:8081/api/v0/machines/99999
echo -e "\n"

echo "44. DELETE /api/v0/networks/$NETWORK_ID (delete network)"
curl -s -X DELETE http://localhost:8081/api/v0/networks/$NETWORK_ID
echo -e "\n"

echo "45. GET /api/v0/networks (should be empty)"
curl -s -X GET http://localhost:8081/api/v0/networks
echo -e "\n"

# Recreate network and DHCP range for IP allocation tests
echo "46. POST /api/v0/networks (recreate network for IP allocation tests)"
NETWORK_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/networks \
  -H "Content-Type: application/json" \
  -d '{"name": "br0", "bridge": "br0", "subnet": "192.168.1.0/24", "gateway": "192.168.1.1", "dns_servers": "8.8.8.8,1.1.1.1", "description": "Main network bridge"}')
echo $NETWORK_RESPONSE

# Extract network ID from response
NETWORK_ID=$(echo $NETWORK_RESPONSE | grep -o '"ID":[0-9]*' | cut -d':' -f2)
echo -e "\nRecreated network with ID: $NETWORK_ID\n"

echo "47. POST /api/v0/networks/$NETWORK_ID/dhcp (create DHCP range for IP allocation tests)"
DHCP_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/networks/$NETWORK_ID/dhcp \
  -H "Content-Type: application/json" \
  -d '{"StartIP": "192.168.1.100", "EndIP": "192.168.1.200", "LeaseTime": "12h"}')
echo $DHCP_RESPONSE
echo -e "\n"

echo "48. POST /api/v0/machines (create machine with network-based IP allocation)"
AUTO_IP_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"auto-ip-server\", \"hostname\": \"auto-host\", \"network_id\": $NETWORK_ID}")
echo $AUTO_IP_RESPONSE

# Extract auto-allocated machine ID
AUTO_MACHINE_ID=$(echo $AUTO_IP_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated machine with auto-allocated IP, ID: $AUTO_MACHINE_ID\n"

echo "49. GET /api/v0/machines/$AUTO_MACHINE_ID (verify auto-allocated IP)"
curl -s -X GET http://localhost:8081/api/v0/machines/$AUTO_MACHINE_ID
echo -e "\n"

echo "50. POST /api/v0/machines (create another machine with same network - should get different IP)"
AUTO_IP_RESPONSE2=$(curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"auto-ip-server2\", \"hostname\": \"auto-host2\", \"network_id\": $NETWORK_ID}")
echo $AUTO_IP_RESPONSE2

# Extract second auto-allocated machine ID
AUTO_MACHINE_ID2=$(echo $AUTO_IP_RESPONSE2 | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo -e "\nCreated second machine with auto-allocated IP, ID: $AUTO_MACHINE_ID2\n"

echo "51. GET /api/v0/machines (verify both machines have different IPs)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

echo "52. DELETE /api/v0/machines/$AUTO_MACHINE_ID (delete first auto-IP machine)"
curl -s -X DELETE http://localhost:8081/api/v0/machines/$AUTO_MACHINE_ID
echo -e "\n"

echo "53. POST /api/v0/machines (create machine with invalid network_id)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "bad-network", "hostname": "bad-host", "network_id": 99999}'
echo -e "\n"

echo "54. POST /api/v0/machines (create machine with network_id but no DHCP range)"
# First create a network without DHCP range
EMPTY_NETWORK_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v0/networks \
  -H "Content-Type: application/json" \
  -d '{"name": "empty-net", "bridge": "empty-br", "subnet": "10.0.0.0/24", "gateway": "10.0.0.1", "dns_servers": "8.8.8.8", "description": "Network without DHCP"}')
EMPTY_NETWORK_ID=$(echo $EMPTY_NETWORK_RESPONSE | grep -o '"ID":[0-9]*' | cut -d':' -f2)

curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"no-dhcp-server\", \"hostname\": \"no-dhcp-host\", \"network_id\": $EMPTY_NETWORK_ID}"
echo -e "\n"

echo "55. POST /api/v0/machines (create machine with both ipv4 and network_id - should fail)"
curl -s -X POST http://localhost:8081/api/v0/machines \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"conflict-server\", \"hostname\": \"conflict-host\", \"ipv4\": \"192.168.1.50\", \"network_id\": $NETWORK_ID}"
echo -e "\n"

echo "56. GET /api/v0/machines (final machine list)"
curl -s -X GET http://localhost:8081/api/v0/machines
echo -e "\n"

echo "57. DELETE /api/v0/networks/$EMPTY_NETWORK_ID (cleanup empty network)"
curl -s -X DELETE http://localhost:8081/api/v0/networks/$EMPTY_NETWORK_ID
echo -e "\n"

echo "58. DELETE /api/v0/networks/$NETWORK_ID (cleanup test network)"
curl -s -X DELETE http://localhost:8081/api/v0/networks/$NETWORK_ID
echo -e "\n"

echo "59. GET /api/v0/networks (should be empty)"
curl -s -X GET http://localhost:8081/api/v0/networks
echo -e "\n"

echo "Enhanced API testing complete with comprehensive IP allocation verification!"

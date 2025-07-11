#!/bin/bash
# Way-d E2E API Test Script
# This script demonstrates a full user flow for the Way-d microservices stack.
# Prerequisites: All services running, jq installed for JSON parsing.

set -e

# --- CONFIG ---
# Use Docker host service names for URLs if running from inside a container
AUTH_URL="http://localhost:8080"
PROFILE_URL="http://localhost:8081"
INTERACTIONS_URL="http://localhost:8082/api"


# --- CLEANUP: CLEAR ALL TABLES ---
echo "Cleaning up interactions DB..."
RESPONSE=$(curl -s -X POST $INTERACTIONS_URL/../debug/clear)
echo "$RESPONSE"

# --- DEBUG: PRINT BLOCKS ---
echo "Debug: Blocks table (before test)"
BLOCKS=$(curl -s -X GET $INTERACTIONS_URL/../debug/blocks)
echo "$BLOCKS" | jq

# --- REGISTER USERS ---
echo "Registering users..."
RESPONSE1=$(curl -s -X POST $AUTH_URL/register -H "Content-Type: application/json" -d '{"email":"user1@example.com","password":"password123","first_name":"Alice","last_name":"Doe","birth_date":"2000-01-01","gender":"female"}')
echo "Register user1 response: $RESPONSE1"
RESPONSE2=$(curl -s -X POST $AUTH_URL/register -H "Content-Type: application/json" -d '{"email":"user2@example.com","password":"password123","first_name":"Bob","last_name":"Smith","birth_date":"2000-01-01","gender":"male"}')
echo "Register user2 response: $RESPONSE2"

# --- LOGIN USERS ---
echo "Logging in users..."
USER1_LOGIN=$(curl -s -X POST $AUTH_URL/login -H "Content-Type: application/json" -d '{"email":"user1@example.com","password":"password123"}')
echo "User1 login response: $USER1_LOGIN"
USER2_LOGIN=$(curl -s -X POST $AUTH_URL/login -H "Content-Type: application/json" -d '{"email":"user2@example.com","password":"password123"}')
echo "User2 login response: $USER2_LOGIN"
USER1_JWT=$(echo $USER1_LOGIN | jq -r .access_token)
USER2_JWT=$(echo $USER2_LOGIN | jq -r .access_token)

# --- CREATE/UPDATE PROFILES ---
echo "Creating profiles..."
curl -s -X PUT $PROFILE_URL/profile/me -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"height":170,"profile_photo_url":"http://example.com/photo1.jpg","location":{"lat":5.35,"lng":-4.02},"occupation":"Engineer","trait":"Friendly"}' > /dev/null
curl -s -X PUT $PROFILE_URL/profile/me -H "Authorization: Bearer $USER2_JWT" -H "Content-Type: application/json" -d '{"height":180,"profile_photo_url":"http://example.com/photo2.jpg","location":{"lat":5.36,"lng":-4.03},"occupation":"Designer","trait":"Chill"}' > /dev/null

# --- GET USER UUIDs FROM PROFILES ---
echo "Fetching user UUIDs..."
USER1_UUID=$(curl -s -X GET $PROFILE_URL/profile/me -H "Authorization: Bearer $USER1_JWT" | jq -r .user_id)
USER2_UUID=$(curl -s -X GET $PROFILE_URL/profile/me -H "Authorization: Bearer $USER2_JWT" | jq -r .user_id)

# --- LIKE FLOW ---
echo "User1 likes User2..."
curl -s -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}' > /dev/null
echo "User2 likes User1 (creates match)..."
curl -s -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER2_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER1_UUID'"}' > /dev/null

# --- DEBUG: PRINT LIKES AND MATCHES ---
echo "Debug: Likes table (user1)"
curl -s -X GET http://wayd-interactions:8082/debug/likes | jq
echo "Debug: Matches table (user1)"
curl -s -X GET http://wayd-interactions:8082/debug/matches | jq

# --- GET MATCH ID ---
echo "Fetching match ID..."
MATCHES=$(curl -s -X GET $INTERACTIONS_URL/matches -H "Authorization: Bearer $USER1_JWT")
echo "Raw matches response: $MATCHES"
MATCH_ID=$(echo $MATCHES | jq -r '.[0].id')
echo "Match ID: $MATCH_ID"

# --- MESSAGING ---
echo "User1 sends message to User2..."
curl -s -X POST $INTERACTIONS_URL/message -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"match_id":"'$MATCH_ID'","content":"Hello from User1!"}' > /dev/null
echo "User2 sends message to User1..."
curl -s -X POST $INTERACTIONS_URL/message -H "Authorization: Bearer $USER2_JWT" -H "Content-Type: application/json" -d '{"match_id":"'$MATCH_ID'","content":"Hello from User2!"}' > /dev/null

# --- GET MESSAGES ---
echo "Fetching messages for the match..."
MESSAGES=$(curl -s -X GET $INTERACTIONS_URL/messages/$MATCH_ID -H "Authorization: Bearer $USER1_JWT")
echo $MESSAGES | jq

# --- DOUBLE LIKE EDGE CASE ---
echo "Testing double-like edge case (should fail on second like)..."
RESPONSE=$(curl -s -w "\nStatus: %{http_code}\n" -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}')
echo "First like response: $RESPONSE (expected 201 or 409)"
RESPONSE=$(curl -s -w "\nStatus: %{http_code}\n" -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}')
echo "Second like response: $RESPONSE (expected 409)"

# --- DISLIKE FLOW ---
echo "Testing dislike flow (should prevent match)..."
curl -s -X POST $INTERACTIONS_URL/dislike -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}' > /dev/null
RESPONSE=$(curl -s -w "\nStatus: %{http_code}\n" -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER2_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER1_UUID'"}')
echo "Like after dislike response: $RESPONSE (expected 409)"

# --- BLOCK-THEN-LIKE EDGE CASE ---
echo "Testing block-then-like edge case (should fail)..."
curl -s -X POST $INTERACTIONS_URL/block -H "Authorization: Bearer $USER2_JWT" -H "Content-Type: application/json" -d '{"blocked_id":"'$USER1_UUID'"}' > /dev/null
RESPONSE=$(curl -s -w "\nStatus: %{http_code}\n" -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}')
echo "Like after being blocked response: $RESPONSE (expected 403)"

# --- BLOCK FLOW ---
echo "User1 blocks User2..."
curl -s -X POST $INTERACTIONS_URL/block -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"blocked_id":"'$USER2_UUID'"}' > /dev/null

# --- TRY TO LIKE AFTER BLOCK (should fail) ---
echo "User1 tries to like User2 after block (should fail)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $INTERACTIONS_URL/like -H "Authorization: Bearer $USER1_JWT" -H "Content-Type: application/json" -d '{"target_id":"'$USER2_UUID'"}')
echo "HTTP status: $RESPONSE (expected 403 or 409)"

# --- UNAUTHORIZED ACCESS TEST ---
echo "Trying to get matches without JWT (should fail)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X GET $INTERACTIONS_URL/matches)
echo "HTTP status: $RESPONSE (expected 401 or 403)"

echo "All tests completed."

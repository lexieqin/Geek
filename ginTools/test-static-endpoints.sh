#!/bin/bash

echo "=== Testing Static Data Endpoints ==="

echo -e "\n1. Testing Job API endpoint:"
echo "URL: http://localhost:8080/tenant/testenv/jobs?requuid=81325fc3-b05e-4d9a-ada2-d2399aebe135&trace=true"
curl -s "http://localhost:8080/tenant/testenv/jobs?requuid=81325fc3-b05e-4d9a-ada2-d2399aebe135&trace=true" | jq '.' | head -20

echo -e "\n2. Testing Datadog Trace endpoint:"
echo "URL: http://localhost:8080/api/datadog/trace/81325fc3b05e4d9aada2d2399aebe135"
curl -s "http://localhost:8080/api/datadog/trace/81325fc3b05e4d9aada2d2399aebe135" | jq '.' | head -20

echo -e "\n3. Testing Sandbox Log endpoint:"
echo "URL: http://localhost:8080/api/sandbox/logs?path=/csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704&file=containers.log"
curl -s "http://localhost:8080/api/sandbox/logs?path=/csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704&file=containers.log" | head -10

echo -e "\n4. Testing Smart Log Analysis endpoint:"
echo "URL: http://localhost:8080/api/sandbox/logs/smart"
curl -s "http://localhost:8080/api/sandbox/logs/smart" | jq '.'

echo -e "\n=== All endpoints tested ==="
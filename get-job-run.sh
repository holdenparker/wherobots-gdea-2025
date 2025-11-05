REGION="aws-us-east-1"
API_KEY=$(cat "gdea-cert.api")

curl -X GET "https://api.cloud.wherobots.com/runs/0s3wureoxx7aw3" \
  -H "accept: application/json" \
  -H "X-API-Key: $API_KEY"
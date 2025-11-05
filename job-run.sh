REGION="aws-us-east-1"
API_KEY=$(cat "gdea-cert.api")
PYTHON_URI="s3://wbts-wbc-bbshnf4fcv/0dj2uiorbw/data/customer-3yn6lxe0dom28o/week-2-geom-corrections-crs.py"

curl -X POST "https://api.cloud.wherobots.com/runs?region=$REGION" \
  -H "accept: application/json" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"runtime\": \"tiny\",
    \"name\": \"bronze_correct_geoms\",
    \"runPython\": {
      \"uri\": \"$PYTHON_URI\"
    },
    \"timeoutSeconds\": 3600
  }"

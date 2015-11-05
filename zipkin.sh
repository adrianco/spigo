curl -vs 192.168.99.100:9411/api/v1/spans -X POST --data @json_metrics/$1_flow.json -H "Content-Type: application/json"

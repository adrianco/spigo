# clean out zipkin database, as we aren't feeding it continuous live data
mysql -h 192.168.99.100 -u zipkin -pzipkin -D zipkin -e 'delete from zipkin_spans; delete from zipkin_annotations'
# post newly created flows, timestamps must be less than one day old to be accepted
curl -vs 192.168.99.100:9411/api/v1/spans -X POST --data @json_metrics/$1_flow.json -H "Content-Type: application/json"

#! /bin/bash
set -e
function end {
  pkill -f aggregator
  exit 1
}

export REDIS_URL="localhost:6379"
./bin/aggregator >bin/writer_logs &
echo "Starting test"
sleep 0.1
mkdir -p bin/client_logs
for run in {1..10}
do
  ./bin/test_client 1000 &>bin/client_logs/$run &
  pids[${run}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done

namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
echo "Namespace count was" $namespaceCount
if [ $namespaceCount != "{\"Err\":null,\"Count\":10000}" ]
then
  end
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
echo "Namespace info was" $namespaceInfo
if [ $namespaceInfo != '{"Err":null,"data":{"metric_keys":{"1":1}}}' ]
then
  end
fi

pkill -f aggregator
echo "Test passed"
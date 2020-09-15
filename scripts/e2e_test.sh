#! /bin/bash
set -e
function end {
  pkill -f aggregator
  exit 1
}

./bin/aggregator "localhost:6379" @>bin/writer_logs &
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
if [ $namespaceCount != "{\"ErrCode\":0,\"Count\":10000}" ]
then
  end
fi

globalCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"metricKey":2,"metricID":1}' http://localhost:50051/count`
echo "Global count was" $globalCount
if [ $globalCount != "{\"ErrCode\":0,\"Count\":20000}" ]
then
  end
fi

globalInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":""}' http://localhost:50051/namespace/get_info`
echo "Global info was" $globalInfo
if [ $globalInfo != '{"error_code":0,"data":{"metric_keys":{"1":1}}}' ]
then
  end
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
echo "Namespace info was" $namespaceInfo
if [ $namespaceInfo != '{"error_code":0,"data":{"metric_keys":{"1":1}}}' ]
then
  end
fi

pkill -f aggregator
echo "Test passed"
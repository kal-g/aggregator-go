#! /bin/bash
set -e
function end {
  pkill -f aggregator
  wait $aggPid
}

export REDIS_URL="localhost:6379"
./bin/aggregator --config "config/aggregator_configs/global" --config "config/aggregator_configs/test" &>bin/writer_logs &
aggPid=$!
trap end EXIT
echo "Starting test"

sleep 0.1

namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
echo "Initial Namespace count was" $namespaceCount
if [ $namespaceCount != '{"Err":{},"Count":0}' ]
then
  end
  exit 1
fi


mkdir -p bin/client_logs
for run in {1..10}
do
  ./bin/test_client 1000 &>bin/client_logs/$run &
  pids[${run}]=$!
  echo "Starting run ${run}"
done

for pid in ${pids[*]}; do
    echo "Waiting on pid $pid"
    wait $pid
    echo "Done waiting on $pid"
done

namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
echo "Namespace count was" $namespaceCount
if [ $namespaceCount != '{"Err":null,"Count":10000}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
echo "Namespace info was" $namespaceInfo
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":1}}}' ]
then
  end
  exit 1
fi

# 
testConfig=`cat config/aggregator_configs/test`
namespaceSetCmd="curl -s --header \"Content-Type: application/json\" --request POST --data '{\"namespaceConfig\":${testConfig}}' http://localhost:50051/namespace/set"
namespaceSet=$(eval $namespaceSetCmd)
echo "Namespace set was" $namespaceSet
if [ $namespaceSet != '{"error":"Namespace exists"}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
echo "Namespace info was" $namespaceInfo
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":1}}}' ]
then
  end
  exit 1
fi

namespaceSetCmd="curl -s --header \"Content-Type: application/json\" --request POST --data '{\"namespaceConfig\":${testConfig}, \"overwriteIfExists\":true}' http://localhost:50051/namespace/set"
namespaceSet=$(eval $namespaceSetCmd)
echo "Namespace set was" $namespaceSet
if [ $namespaceSet != '{"error":""}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
echo "Namespace info was" $namespaceInfo
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":0}}}' ]
then
  end
  exit 1
fi

pkill -f aggregator
echo "Test passed"
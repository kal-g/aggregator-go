#! /bin/bash
set -e
function end {
  set +e
  pkill -f aggregator
  pkill -f test_client
  wait $aggPid
}

export REDIS_URL="localhost:6379"
export ZOOKEEPER_URL="localhost:2181"
./bin/aggregator --config "config/aggregator_configs/global" --config "config/aggregator_configs/test" &>bin/writer_logs &
aggPid=$!
trap end EXIT
sleep 5
echo "Starting test"

namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
printf "Initial Namespace count\n"
printf "%s\n\n" $namespaceCount
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
done

for pid in ${pids[*]}; do
    echo "Waiting on pid $pid"
    wait $pid
done
printf "\n"

namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
printf "Namespace count after ingest\n"
printf "%s\n\n" $namespaceCount
if [ $namespaceCount != '{"Err":null,"Count":10000}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
printf "Namespace info was\n"
printf "%s\n\n" $namespaceInfo
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":1}}}' ]
then
  end
  exit 1
fi

testConfig=`cat config/aggregator_configs/test`
namespaceSetCmd="curl -s --header \"Content-Type: application/json\" --request POST --data '{\"namespaceConfig\":${testConfig}}' http://localhost:50051/namespace/set"
namespaceSet=$(eval $namespaceSetCmd)
printf "Set namespace test, no override\n"
echo $namespaceSet
printf "\n"
if [ "$namespaceSet" != '{"error":"Namespace exists"}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
printf "Namespace info for test\n"
printf "$namespaceInfo\n\n"
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":1}}}' ]
then
  end
  exit 1
fi

testAltConfig=`cat config/aggregator_configs/test-alt`
namespaceSetCmd="curl -s --header \"Content-Type: application/json\" --request POST --data '{\"namespaceConfig\":${testAltConfig}, \"overwriteIfExists\":true}' http://localhost:50051/namespace/set"
namespaceSet=$(eval $namespaceSetCmd)
printf "Set namespace test, with override\n"
echo $namespaceSet
printf "\n"
if [ $namespaceSet != '{"error":""}' ]
then
  end
  exit 1
fi

sleep 5

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test"}' http://localhost:50051/namespace/get_info`
printf "Namespace info for test\n"
printf "$namespaceInfo\n\n"
if [ $namespaceInfo != '{"error":"","data":{"metric_keys":{"1":0}}}' ]
then
  end
  exit 1
fi

namespaceInfo=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test2"}' http://localhost:50051/namespace/get_info`
printf "Namespace info for test2\n"
printf "$namespaceInfo\n\n"
if [ "$namespaceInfo" != '{"error":"Namespace not found","data":{"metric_keys":null}}' ]
then
  end
  exit 1
fi

test2Config=`cat config/aggregator_configs/test2`
namespaceSetCmd="curl -s --header \"Content-Type: application/json\" --request POST --data '{\"namespaceConfig\":${test2Config}}' http://localhost:50051/namespace/set"
namespaceSet=$(eval $namespaceSetCmd)
printf "Set namespace test2, no override\n"
echo $namespaceSet
printf "\n"
if [ "$namespaceSet" != '{"error":""}' ]
then
  end
  exit 1
fi

echo "Test passed"
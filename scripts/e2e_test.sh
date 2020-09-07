#! /bin/bash

./bin/aggregator "bin/rocksdb_storage" @>bin/writer_logs &

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
sleep 1
#namespaceCount=`./bin/storage_reader test:1:2`
namespaceCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
if [ $namespaceCount != "{\"ErrCode\":0,\"Count\":10000}" ]
then
  echo "Namespace count was " $namespaceCount
  pkill -f aggregator
  rm -rf bin/rocksdb_storage
  exit 1
fi

#globalCount=`./bin/storage_reader :1:2`
globalCount=`curl -s --header "Content-Type: application/json" --request POST --data '{"metricKey":2,"metricID":1}' http://localhost:50051/count`
if [ $globalCount != "{\"ErrCode\":0,\"Count\":20000}" ]
then
  echo "Global count was " $globalCount
  pkill -f aggregator
  rm -rf bin/rocksdb_storage
  exit 1
fi

echo "Test passed"

M=`pgrep go_writer`
if [ `echo $M` ]
then
  echo "Warning: clients still running"
  echo $M
  pkill -f go_writer
fi

sleep 0.1
pkill -f aggregator
rm -rf bin/rocksdb_storage
exit 0

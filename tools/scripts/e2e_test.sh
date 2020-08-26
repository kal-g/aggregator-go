#! /bin/bash

./bin/aggregator "rocksdb_storage" @>writer_logs &

sleep 0.1
mkdir -p client_logs
for run in {1..10}
do
  ./bin/test_client 1000 &>client_logs/$run &
  pids[${run}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done

namespaceCount=`./bin/storage_reader test:1:2`
if [[ $namespaceCount -ne 10000 ]]
then
  echo "Namespace count was " $namespaceCount
  pkill -f aggregator
  rm -rf rocksdb_storage
  exit 1
fi

globalCount=`./bin/storage_reader :1:2`
if [[ $globalCount -ne 20000 ]]
then
  echo "Global count was " $globalCount
  pkill -f aggregator
  rm -rf rocksdb_storage
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
rm -rf rocksdb_storage
exit 0

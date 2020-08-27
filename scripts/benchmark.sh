#! /bin/bash

./bin/aggregator "rocksdb_storage" &>bin/writer_logs &

sleep 0.1
mkdir -p bin/client_logs

if [ "$1" != "" ]; then
    num_clients=$1
else
   num_clients=1
fi


SECONDS=0
for run in $(seq 1 $num_clients)
do
  ./bin/test_client 10000 &>bin/client_logs/$run &
  pids[${run}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done
duration=$SECONDS

count=`./bin/storage_reader`
echo "Number of concurrent clients: $num_clients"
echo -n "RPS "
echo "scale=2 ; $count / $duration" | bc


pkill test_client
pkill aggregator
rm -rf rocksdb_storage

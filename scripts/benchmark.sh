#! /bin/bash

./bin/aggregator "bin/rocksdb_storage" &>bin/writer_logs &

sleep 0.1
mkdir -p bin/client_logs

if [ "$1" != "" ]; then
    num_clients=$1
else
   num_clients=1
fi

echo "Number of concurrent clients: $num_clients"

SECONDS=0
for run in $(seq 1 $num_clients)
do
  ./bin/test_client 2000 &>bin/client_logs/$run &
  pids[${run}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done
duration=$SECONDS

sleep 1
count=`./bin/storage_reader test:1:2`
echo -n "Count "
echo $count
echo -n "RPS "
echo "scale=2 ; $count / $duration" | bc


pkill test_client
pkill aggregator
rm -rf bin/rocksdb_storage

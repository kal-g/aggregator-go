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

INIT_TS=`date +%s.%N`
for run in $(seq 1 $num_clients)
do
  ./bin/test_client 10000 &>bin/client_logs/$run &
  pids[${run}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done
END_TS=`date +%s.%N`

sleep 1
count=`./bin/storage_reader test:1:2`
echo -n "Count "
echo $count
echo -n "RPS "
echo "$count / ($END_TS - $INIT_TS)" | bc -l


pkill test_client
pkill aggregator
rm -rf bin/rocksdb_storage

#! /bin/bash
set -e

export REDIS_URL="localhost:6379"
./bin/aggregator &>bin/writer_logs &

sleep 3
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
    echo "Waiting on pid $pid"
    wait $pid
done

sleep 1

for run in $(seq 1 $num_clients); do
    wc -l bin/client_logs/$run
done

END_TS=`date +%s.%N`

count=`curl -s --header "Content-Type: application/json" --request POST --data '{"namespace":"test","metricKey":2,"metricID":1}' http://localhost:50051/count`
parsedCount=`echo $count | egrep -o Count.* | egrep -o [0-9][0-9]*`
echo -n "Count "
echo $parsedCount
echo -n "RPS "
parsedEnd=`echo $END_TS | egrep -o ^[0-9]*`
parsedInit=`echo $INIT_TS | egrep -o ^[0-9]*`
echo "$parsedCount / ($parsedEnd - $parsedInit)" | bc -l

pkill -f aggregator
echo "Benchmark finished"

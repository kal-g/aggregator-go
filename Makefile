
all: aggregator test_client

aggregator:
	@mkdir -p bin
	@go build -o bin/aggregator cmd/aggregator/main.go

test_client:
	@mkdir -p bin
	@go build -o bin/test_client cmd/test_client/main.go

clean:
	@rm -rf bin

test:
	@go test ./...

client: all
	@./scripts/client.sh

e2e_test: all
	@./scripts/e2e_test.sh

benchmark: all
	@NODE_NAME=agg ./scripts/benchmark.sh 1
	@sleep 5
	@NODE_NAME=agg ./scripts/benchmark.sh 3

start_redis:
	-docker rm -f redis
	-docker run -p 6379:6379 -d --network=agg --name redis redis 2>/dev/null

start_zk:
	-docker rm -f zk
	-docker run -p 2181:2181 -d --network=agg --name zk zookeeper 2>/dev/null

start_net:
	-docker network create agg 2>/dev/null

docker_build_local: all
	-docker image rm -f kalgg/aggregator-go:local
	cd .. && docker build -t kalgg/aggregator-go:local -f aggregator-go/config/docker/main/Dockerfile .

start_agg:
	-docker stop $(NODE_NAME)
	-docker rm $(NODE_NAME)
	docker run -e REDIS_URL=redis:6379 -e ZOOKEEPER_URL=zk:2181 -e NODE_NAME=$(NODE_NAME) --network=agg --name $(NODE_NAME) -p $(NODE_PORT):50051 kalgg/aggregator-go:local

dev:
	-docker rm -f agg
	docker run -e REDIS_URL=redis:6379 -e ZOOKEEPER_URL=zk:2181 -e NODE_NAME=agg --network=agg --name agg -p 50051:50051 kalgg/aggregator-go:local

docker_run_all: start_net start_redis start_zk start_agg

run: all
	NODE_NAME=agg REDIS_URL=localhost:6379 ZOOKEEPER_URL=localhost:2181 ./bin/aggregator

run_benchmark: all
	$(eval IMG=$(shell sh -c "docker build --no-cache -q config/docker/benchmark"))
	@docker run -it $(IMG)
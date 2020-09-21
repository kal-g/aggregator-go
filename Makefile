
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

e2e_test: all
	@./scripts/e2e_test.sh

benchmark: all
	@./scripts/benchmark.sh 1
	@./scripts/benchmark.sh 3

start_redis:
	docker run -p 6379:6379 -d redis

start_zk:
	docker run -p 2181:2181 -d zookeeper

run: all
	NODE_NAME=agg REDIS_URL=localhost:6379 ZOOKEEPER_URL=localhost:2181 ./bin/aggregator

run_benchmark: all
	$(eval IMG=$(shell sh -c "docker build --no-cache -q config/docker/benchmark"))
	@docker run -it $(IMG)
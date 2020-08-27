
all: aggregator test_client storage_reader

aggregator:
	@mkdir -p bin
	@go build -o bin/aggregator cmd/aggregator/main.go

test_client:
	@mkdir -p bin
	@go build -o bin/test_client cmd/test_client/main.go

storage_reader:
	@mkdir -p bin
	@go build -o bin/storage_reader cmd/storage_reader/main.go 

clean:
	@rm -rf bin

test:
	@go test ./...

e2e_test: all
	@./scripts/e2e_test.sh

benchmark: all
	@./scripts/benchmark.sh
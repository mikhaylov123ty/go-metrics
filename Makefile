SERVER_PORT:="30011"
AGENT_ARGS:="-config ./cmd/agent/config.json"
SERVER_ARGS:="-config ./cmd/server/config.json"
TEMP_FILE:="temp_file"
DSN:="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

build: buildAgent buildServer

buildAgent:
	go build -o ./cmd/agent/agent ./cmd/agent/

buildServer:
	go build -o ./cmd/server/server ./cmd/server/

lint:
	go vet ./...

unitTests:
	go clean --testcache &&\
	go test ./...

race:
	go test -v -race ./...

pre-push: lint race runMultichecker buildAgent buildServer unitTests tests cleanup

cleanup:
	rm -f ./cmd/agent/agent && \
    rm -f ./cmd/server/server && \
    rm -f ./cmd/staticlint/multicheck && \
    rm -f ./temp_file && \
    rm -f ./tempFile.txt && \
    rm -f ./cert/*.pem && \
    rm -f ./tempFileFromConfig.txt

#test increment 1-15
tests: buildAgent buildServer
	#increment 1
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$$ \
  			-binary-path=cmd/server/server

	#increment 2
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2[AB]*$$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent

	#increment 3
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration3[AB]*$$ \
            -source-path=. \
          	-agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server

    #increment 4
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration4$$ \
            -agent-binary-path=cmd/agent/agent \
          	-binary-path=cmd/server/server \
            -source-path=. \
            -server-port=${SERVER_PORT}

	#increment 5
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration5$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
           	-source-path=. \
            -server-port=${SERVER_PORT}

    #increment 6
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration6$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -source-path=. \
            -server-port=${SERVER_PORT}

	#increment 7
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration7$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
          	-source-path=. \
            -server-port=${SERVER_PORT}

    #increment 8
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration8$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=${SERVER_PORT} \
            -source-path=.

	#increment 9
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration9$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -file-storage-path=${TEMP_FILE} \
            -server-port=${SERVER_PORT} \
            -source-path=.

	#increent 10
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration10[AB]$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn=${DSN} \
           	-server-port=${SERVER_PORT} \
            -source-path=.

	#increment 11
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration11$$ \
			  -agent-binary-path=cmd/agent/agent \
			  -binary-path=cmd/server/server \
			  -database-dsn=${DSN} \
			  -server-port=${SERVER_PORT} \
			  -source-path=.

	#increment 12
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration12$$ \
			  -agent-binary-path=cmd/agent/agent \
			  -binary-path=cmd/server/server \
			  -database-dsn=${DSN} \
			  -server-port=${SERVER_PORT} \
			  -source-path=.

	#increment 13
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration13$$ \
              -agent-binary-path=cmd/agent/agent \
              -binary-path=cmd/server/server \
              -database-dsn=${DSN} \
              -server-port=${SERVER_PORT} \
              -source-path=.

	#increment 14
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration14$$ \
               -agent-binary-path=cmd/agent/agent \
               -binary-path=cmd/server/server \
               -database-dsn=${DSN} \
               -key="${TEMP_FILE}" \
               -server-port=${SERVER_PORT} \
               -source-path=.

#increment 16
benchAgentCPU:
	go test -bench=. ./internal/client/collector/

benchAgentMem:
	go test -bench=. ./internal/client/collector/ -benchmem

#increment 17
goimports:
	goimports -local metrics -w ./internal/..

#increment 18
doc:
	godoc -http=:9090

#increment 19
buildMultichecker:
	cd ./cmd/staticlint && \
	go build -o multicheck ./main.go

runMultichecker: buildMultichecker
	./cmd/staticlint/multicheck ./...

#increment 20
runAgentWithFlags:
	go run -ldflags "-X main.buildVersion=v0.01 -X 'main.buildDate=$$(date +'%Y/%m/%d')'" cmd/agent/main.go

runServerWithFlags:
	go run -ldflags "-X main.buildVersion=v0.01 -X 'main.buildDate=$$(date +'%Y/%m/%d')'" cmd/server/main.go

#increment 21
generateCert:
	cd ./cert && \
	go run main.go

#increment 25
generateProto:
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  internal/server/proto/handlers.proto
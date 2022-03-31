## help: show makefile usage
PROJECTNAME := $(shell basename "$(PWD)")
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo ""
	@find . -maxdepth 1 -type f \( -name Makefile -or -name "*.mk" \) -exec cat {} \+ | sed -n 's/^##//p' | column -t -s ':' |  sed -e 's/^/ /'

## start-deps: start all project dependencies like database, redis, etc..
MY_IP?=`ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1`
start-deps:
	@MY_IP=${MY_IP} docker-compose -f dev/docker-compose.yml up -d

stop-deps:
	@docker-compose -f dev/docker-compose.yml down

## build-docker: builds the docker image
build-docker:
	@docker build -t fo-aas .

## test: run the project's tests
test: 
	@go test ./... -cover -coverprofile=coverage.out

## coverage: show coverage report on the terminal using gocov
coverage: test
	@gocov convert coverage.out | gocov report

## coverage-html: open coverage report on the browser
coverage-html: test
	@go tool cover -html=coverage.out

## lint: run gometalinter following .gometalinter.json configuration on all directories
lint:
	@golangci-lint run

run: start-deps
	@docker run -p 8080:8080 -d fo-aas -session-length 1000 -tokens 10 -timeout 3000
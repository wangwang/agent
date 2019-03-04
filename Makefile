run:
	go run *.go run $(JOB)
.PHONY: run

serve:
	go run *.go serve --auth-token-secret 'TzRVcspTmxhM9fUkdi1T/0kVXNETCi8UdZ8dLM8va4E'
.PHONY: serve

test:
	go test -short -v ./...
.PHONY: test

build:
	rm -rf build
	go build -o build/agent main.go
.PHONY: build

docker.build: build
	-docker stop agent
	-docker rm agent
	-docker rmi agent
	docker build -t agent -f Dockerfile.test .
.PHONY: docker.build

docker.run: docker.build
	-docker stop agent
	docker run --net=host --name agent -tdi agent bash -c "./agent serve --auth-token-secret 'TzRVcspTmxhM9fUkdi1T/0kVXNETCi8UdZ8dLM8va4E'"
	sleep 2
.PHONY: docker.run

e2e: docker.run
	ruby test/e2e/$(TEST).rb
.PHONY: e2e

release.major:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1+1 ".0.0" }');          echo $$new; git tag $$new; git push origin $$new

release.minor:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1 "." $$2 + 1 ".0" }');  echo $$new; git tag $$new; git push origin $$new

release.patch:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1 "." $$2 "." $$3+1 }'); echo $$new; git tag $$new; git push origin $$new

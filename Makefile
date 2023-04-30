.PHONY: test
test: clean
	@go test ./... -v

.PHONY: clean
clean:
	@go mod tidy
	@go vet ./...
	@go fmt ./...


.PHONY: e2e
e2e:
	@go clean -testcache
	@go test ./... --tags=e2e -v
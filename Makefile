.PHONY: codeline
codeline:
	@tokei .
	
.PHONY: test
test: clean
# 只跑单元测试
	@go test ./... -v

.PHONY: e2e
# 单元测试和集成测试一起跑
# @go test -tags=integration ./... -v
	@go test -tags=integration -v

.PHONY: clean
clean:
	@go mod tidy
	@go clean -testcache
	@go vet ./...
	@go fmt ./...


.PHONY: e2e
e2e:
	@go clean -testcache
	@go test --tags=integration ./... -v

.PHONY: dockerup
dockerup:
	docker compose up

.PHONY: dockerdown
dockerdown:
	docker compose down


.PHONY: mockgen
mockgen:
	mockgen -destination=cache/mocks/mock_redis_cmdable.gen.go -package=mocks github.com/redis/go-redis/v9 Cmdable
	mockgen -destination=micro/net/mocks/net_conn.gen.go -package=mocks net Conn
	mockgen -destination=micro/rpc/proxy_test.gen.go -package=rpc -source=micro/rpc/types.go Proxy
	@cd micro/proto && protoc --go_out=. user.proto



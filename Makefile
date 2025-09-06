.PHONY: test_conver
test_conver:
	go test -cover -coverprofile=coverage.out -gcflags='all=-N -l' -coverpkg ./... -timeout=5m ./... -cpu=1
	@echo "Coverage report:"
	@go tool cover -html=coverage.out
	@echo "Coverage report generated successfully"

.PHONY: run
run:
	@echo "Building the application..."
	@go build -gcflags='all=-N -l' -o telecom_repair_hub main.go
	@./telecom_repair_hub

.PHONY: gen
gen:
	@go run command/gen.go
.PHONY: lint swag

lint:
	golangci-lint run ;

swag:
	swag init --generalInfo="./cmd/{{.module}}/main.go" --dir="./" --output="./api/swagger" ;

test:
	go test ./... -race -count=1 ;

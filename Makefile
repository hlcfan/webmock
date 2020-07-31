test:
	go test . -v -race

cover:
	go test -cover
	go tool cover -html=coverage.out


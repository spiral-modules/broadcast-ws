test:
	composer update
	go test -v -race -cover

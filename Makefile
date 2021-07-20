linux:
	GOOS=linux GOARCH=amd64 go build
windows:
	GOOS=windows GOARCH=amd64 go build
mac:
	GOOS=darwin GOARCH=amd64 go build

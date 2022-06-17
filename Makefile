make:
	GOOS=linux GOARCH=amd64 go build -o bin/clustsafe_exporter
	GOOS=windows GOARCH=amd64 go build -o bin/clustsafe_exporter.exe

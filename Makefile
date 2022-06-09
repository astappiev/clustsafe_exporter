linux:
	GOOS=linux GOARCH=amd64 go build -o bin/clustsafe_exporter

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/clustsafe_exporter.exe

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/clustsafe_exporter

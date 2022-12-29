env GOOS=linux GOARCH=amd64 go build -o ./target/main ./cmd/lambda
Compress-Archive -Path ./target/main -DestinationPath ./target/main.zip -Force
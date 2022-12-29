env GOOS=linux GOARCH=amd64 go build -o ./target/main ./cmd/lambda
cp ./internal/db/migrations/* ./target/internal/db/migrations
Compress-Archive -Path ./target/* -DestinationPath ./target/main.zip -Force
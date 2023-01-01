md ./target/main/internal/db/migrations -ea 0
env GOOS=linux GOARCH=amd64 go build -o ./target/main/main ./cmd/lambda
cp ./internal/db/migrations/* ./target/main/internal/db/migrations
Compress-Archive -Path ./target/main/* -DestinationPath ./target/main/main.zip -Force

md ./target/jobs -ea 0
env GOOS=linux GOARCH=amd64 go build -o ./target/jobs/main ./cmd/lambda-jobs
Compress-Archive -Path ./target/jobs/* -DestinationPath ./target/jobs/jobs.zip -Force
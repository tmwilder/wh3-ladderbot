package main

import (
	"context"
	"discordbot/internal/app"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"log"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Cold starting Gin...")
	r := app.GetGin()
	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// TODO secrets manager integration to populate env variables once we go from test app to prod.
	// For this model just env also fine.
	lambda.Start(Handler)
}

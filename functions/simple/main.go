package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, evt json.RawMessage) (string, error) {
	return fmt.Sprintf("Ctx: %s\nEvt: %s\n", ctx, evt), nil
}

func main() {
	lambda.Start(handler)
}

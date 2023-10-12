package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const (
	stableDiffusionXLModelID = "stability.stable-diffusion-xl-v0" //https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids-arns.html
	defaultRegion            = "us-east-1"
)

var brc *bedrockruntime.Client

func init() {

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	brc = bedrockruntime.NewFromConfig(cfg)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	prompt := req.Body

	log.Println("input -", prompt)
	log.Println("cfg scale -", req.QueryStringParameters["cfg_scale"])
	log.Println("seed -", req.QueryStringParameters["seed"])
	log.Println("steps -", req.QueryStringParameters["steps"])

	cfgScaleF, _ := strconv.ParseFloat(req.QueryStringParameters["cfg_scale"], 64)
	seed, _ := strconv.Atoi(req.QueryStringParameters["seed"])
	steps, _ := strconv.Atoi(req.QueryStringParameters["steps"])

	payload := Request{
		TextPrompts: []TextPrompt{{Text: prompt}},
		CfgScale:    cfgScaleF,
		//Seed:        seed,
		Steps: steps,
	}

	if seed > 0 {
		payload.Seed = seed
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	output, err := brc.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String(stableDiffusionXLModelID),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		log.Fatal("failed to invoke model: ", err)
	}

	var resp Response

	err = json.Unmarshal(output.Body, &resp)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}

	image := resp.Artifacts[0].Base64

	fmt.Println("got image from model\n", image)

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      http.StatusOK,
		Body:            image,
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST,OPTIONS",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}

type Request struct {
	TextPrompts []TextPrompt `json:"text_prompts"`
	CfgScale    float64      `json:"cfg_scale"`
	Steps       int          `json:"steps"`
	Seed        int          `json:"seed"`
}

type TextPrompt struct {
	Text string `json:"text"`
}

type Response struct {
	Result    string     `json:"result"`
	Artifacts []Artifact `json:"artifacts"`
}

type Artifact struct {
	Base64       string `json:"base64"`
	FinishReason string `json:"finishReason"`
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	_ "github.com/lib/pq"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	RequestId   string      `json:"requestId"`
	RequestTime string      `json:"requestTime"`
	Data        DataRequest `json:"data"`
}

type DataRequest struct {
	Image string `json:"image"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	ResponseId      string                                `json:"responseId"`
	ResponseTime    string                                `json:"responseTime"`
	ResponseCode    string                                `json:"responseCode"`
	ResponseMessage string                                `json:"responseMessage"`
	Data            *rekognition.SearchFacesByImageOutput `json:"data"`
}

var logger *log.Logger

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	datetime := time.Now().UTC()
	// BodyRequest will be used to take the json response from client and build it
	bodyRequest := BodyRequest{
		RequestId: "",
	}

	// Unmarshal the json, return 404 if error
	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 401}, nil
	}

	//verify uuid not null
	if bodyRequest.RequestId == "" {
		return events.APIGatewayProxyResponse{Body: "requestId can not be null", StatusCode: 401}, nil

	}

	logger = log.New(os.Stderr, "", log.LstdFlags)
	logger.SetPrefix("[requestId:" + bodyRequest.RequestId + "]")

	//verify datetime format RFC3339
	parsedTime, err := time.Parse(time.RFC3339, bodyRequest.RequestTime)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error() + "parsedTime: " + parsedTime.GoString(), StatusCode: 401}, nil
	}

	//get user from request
	ouput, err := searchface(bodyRequest.Data.Image)

	responseCode := "06"
	if err == nil {
		responseCode = "00"
	}

	// We will build the BodyResponse and send it back in json form
	bodyResponse := BodyResponse{
		ResponseId:      uuid.New().String(),
		ResponseTime:    datetime.Format(time.RFC3339),
		ResponseCode:    responseCode,
		ResponseMessage: "rows: " + strconv.Itoa(len(ouput.FaceMatches)),
		Data:            ouput,
	}

	// Marshal the response into json bytes, if error return 404
	response, err := json.Marshal(&bodyResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	// print response json
	responseJson, err := json.Marshal(bodyResponse)
	if err != nil {
		logger.Fatalln("cannot parse response to json")
	}
	logger.Println("RESPONSE: " + string(responseJson))

	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}

func searchface(image string) (*rekognition.SearchFacesByImageOutput, error) {

	collectionId := os.Getenv("COLLECTION_ID")

	// Load the SDK's configuration from environment and shared config, and
	// create the client with this.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Fatalf("failed to load SDK configuration, %v", err)
	}

	client := rekognition.NewFromConfig(cfg)

	input := &rekognition.SearchFacesByImageInput{
		Image: &types.Image{
			Bytes: []byte(image),
		},
		CollectionId: &collectionId,
	}

	output, err := client.SearchFacesByImage(context.TODO(), input)

	if err != nil {
		logger.Fatalf("err when searchd image, %v", err)
	}

	return output, err

}

package main

import (
	"encoding/json"
	"fmt"

	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Person struct {
	Name  string `json:"name`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	p := Person{
		Name:  "Peter Griffin",
		Email: "peterg@msn.com",
		Age:   45,
	}
	buf := bytes.Buffer{}
	json.NewEncoder(&buf).Encode(p)

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := sqs.New(sess)
	getQueueParams := &sqs.GetQueueUrlInput{
		QueueName: aws.String("test_notifications_queue"), // Required
	}
	getQueueResp, err := svc.GetQueueUrl(getQueueParams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(getQueueResp)
	queueUrl := getQueueResp.QueueUrl

	ch := make(chan string)
	go receiveMessage(*queueUrl, ch)

	sendParams := &sqs.SendMessageInput{
		MessageBody:  aws.String(buf.String()), // Required
		QueueUrl:     aws.String(*queueUrl),    // Required
		DelaySeconds: aws.Int64(1),
	}
	sendResp, err := svc.SendMessage(sendParams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(sendResp)

	for msg := range ch {
		fmt.Println(msg)
	}
}

func receiveMessage(queueUrl string, ch chan string) {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := sqs.New(sess)
	params := &sqs.ReceiveMessageInput{
		QueueUrl:          aws.String(queueUrl), // Required
		VisibilityTimeout: aws.Int64(5),
	}
	for {
		resp, err := svc.ReceiveMessage(params)
		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		if len(resp.Messages) > 0 {
			ch <- resp.String()
		}
	}
	close(ch)
}

package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func BuildSession() *session.Session {


	sessionConfig := &aws.Config{
		Region:      aws.String("us-west-2"),
	}

	sess, err := session.NewSession(sessionConfig)
	if err != nil {
		log.Println("error establishing session")
		return nil
	}
	return sess

}
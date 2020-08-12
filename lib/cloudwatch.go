package command

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/kvz/logstreamer"
)

var cwSvc = cloudwatchlogs.New(sess)

// Stream invocation logs to stderr/stdout
func Stream(logGroup, commandID, instanceID *string) {
	logger := log.New(os.Stdout, *instanceID+" ", 0)

	logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
	defer logStreamerOut.Close()

	go streamFromCloudwatch(
		logStreamerOut,
		*logGroup,
		fmt.Sprintf("%s/%s/aws-runShellScript/stdout", *commandID, *instanceID),
		os.Stderr,
	)

	logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", false)
	defer logStreamerErr.Close()

	go streamFromCloudwatch(
		logStreamerErr,
		*logGroup,
		fmt.Sprintf("%s/%s/aws-runShellScript/stderr", *commandID, *instanceID),
		os.Stdout,
	)

}

func streamFromCloudwatch(logStreamer *logstreamer.Logstreamer, logGroup, logStream string, w io.Writer) {

	var nextToken string

	for {
		logEventsInput := cloudwatchlogs.GetLogEventsInput{
			StartFromHead: aws.Bool(true),
			LogGroupName:  aws.String(logGroup),
			LogStreamName: aws.String(logStream),
		}

		if nextToken != "" {
			logEventsInput.NextToken = aws.String(nextToken)
		}

		logEvents, err := cwSvc.GetLogEvents(&logEventsInput)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// Get error details
				if awsErr.Code() == "ResourceNotFoundException" {
					time.Sleep(time.Second * 5)
					continue
				}
			} else {
				log.Fatal(err)
			}
		}

		for _, e := range logEvents.Events {
			fmt.Fprintln(logStreamer, *e.Message)
		}

		if logEvents.NextForwardToken != nil {
			nextToken = *logEvents.NextForwardToken
		}

		logStreamer.FlushRecord()

		time.Sleep(randomSeconds(5))
	}

}

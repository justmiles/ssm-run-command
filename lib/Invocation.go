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
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/kvz/logstreamer"
)

// Invocation is a running command
type Invocation struct {
	streaming         bool
	stopStreaming     chan struct{}
	commandInvocation *ssm.CommandInvocation
}

var cwSvc = cloudwatchlogs.New(sess)

// Stream invocation logs to stderr/stdout
func (inv *Invocation) Stream() {
	logger := log.New(os.Stdout, *inv.commandInvocation.InstanceId+" ", 0)

	logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
	defer logStreamerOut.Close()

	go streamFromCloudwatch(
		logStreamerOut,
		*inv.commandInvocation.CloudWatchOutputConfig.CloudWatchLogGroupName,
		fmt.Sprintf("%s/%s/aws-runShellScript/stdout", *inv.commandInvocation.CommandId, *inv.commandInvocation.InstanceId),
		os.Stderr,
	)

	logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", false)
	defer logStreamerErr.Close()

	go streamFromCloudwatch(
		logStreamerErr,
		*inv.commandInvocation.CloudWatchOutputConfig.CloudWatchLogGroupName,
		fmt.Sprintf("%s/%s/aws-runShellScript/stderr", *inv.commandInvocation.CommandId, *inv.commandInvocation.InstanceId),
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

package command

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ec2Svc           = ec2.New(sess)
	ssmSvc           = ssm.New(sess)
	exitCode         = 0
	executionTimeout = "172800" // 2 days, TODO: set as argument
)

// Command to execute
type Command struct {
	DryRun           bool // TODO
	Targets          []string
	TargetLimit      int
	ExecutionTimeout int
	MaxConcurrency   string
	MaxErrors        string
	Comment          string
	LogGroup         string
	Command          []string
	SSMCommand       *ssm.Command
	invocationErrors []error
}

// Run command and stream results to stdout
func (c *Command) Run() (int, error) {

	// Ensure we always target running instances
	c.Targets = append(c.Targets, "instance-state-name=running")

	targets, err := c.targets()
	if err != nil {
		return 1, err
	}

	// Randomize and limit the number of instances
	instanceIDs, err := randomTargets(targets, c.TargetLimit)
	if err != nil {
		return 1, err
	}

	// Split the instanceIDs into batches of 50 items.
	batch := 50
	for i := 0; i < len(instanceIDs); i += batch {
		j := i + batch
		if j > len(instanceIDs) {
			j = len(instanceIDs)
		}

		if err := c.RunCommand(instanceIDs[i:j]); err != nil {
			return 1, err
		}

	}

	return exitCode, nil
}

// RunCommand against EC2 instances
func (c *Command) RunCommand(instanceIDs []*string) (err error) {
	input := ssm.SendCommandInput{
		TimeoutSeconds: aws.Int64(30),
		InstanceIds:    instanceIDs,
		MaxConcurrency: &c.MaxConcurrency,
		MaxErrors:      &c.MaxErrors,
		DocumentName:   aws.String("AWS-RunShellScript"),
		Comment:        &c.Comment,
		CloudWatchOutputConfig: &ssm.CloudWatchOutputConfig{
			CloudWatchLogGroupName:  &c.LogGroup,
			CloudWatchOutputEnabled: aws.Bool(true),
		},
	}

	input.Parameters = map[string][]*string{
		"commands":         aws.StringSlice([]string{strings.Join(c.Command, " ")}),
		"executionTimeout": aws.StringSlice([]string{fmt.Sprintf("%d", c.ExecutionTimeout)}),
	}

	output, err := ssmSvc.SendCommand(&input)
	if err != nil {
		return fmt.Errorf("Error invoking SendCommand: %s", err)
	}

	c.SSMCommand = output.Command

	// Wait for each instance to finish
	var wg sync.WaitGroup

	for _, instanceID := range instanceIDs {
		wg.Add(1)
		go Stream(&c.LogGroup, output.Command.CommandId, instanceID)
		go WaitForInstance(output.Command.CommandId, instanceID, &wg)
	}

	wg.Wait()

	time.Sleep(10 * time.Second)
	return nil
}

// WaitForInstance waits for an instance to finish executing
func WaitForInstance(commandID, instanceID *string, wg *sync.WaitGroup) {

	defer wg.Done()
	var err error

	err = ssmSvc.WaitUntilCommandExecuted(&ssm.GetCommandInvocationInput{
		CommandId:  commandID,
		InstanceId: instanceID,
	})

	if err != nil {

		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case request.WaiterResourceNotReadyErrorCode:
				exitCode = 1
			default:
				fmt.Printf("ERROR waiting for instance %s to execute: %s\n", *instanceID, err)
			}
		}
	}

}

// Status returns a frendly status to show to user
func (c *Command) Status() string {

	if c.SSMCommand == nil {
		return "command not yet started"
	}

	switch *c.SSMCommand.StatusDetails {
	case "NoInstancesInTag":
		return "no instances matched your targets"
	case "Success":
		return "command exited successfully"
	default:
		return *c.SSMCommand.StatusDetails
	}
}

func (c Command) targets() (targets []*ssm.Target, err error) {

	for _, target := range c.Targets {
		s := strings.SplitN(target, "=", 2)
		if len(s) != 2 {
			return targets, fmt.Errorf("unable to derive target from: %s", target)
		}
		targets = append(targets, &ssm.Target{
			Key:    aws.String(s[0]),
			Values: aws.StringSlice(strings.Split(s[1], ",")),
		})
	}

	return targets, err
}

func randomTargets(targets []*ssm.Target, targetLimit int) (instances []*string, err error) {
	input := &ec2.DescribeInstancesInput{
		Filters: toFilter(targets),
	}
	result, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		return
	}

	for _, reservation := range result.Reservations {

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(reservation.Instances), func(i, j int) {
			reservation.Instances[i], reservation.Instances[j] = reservation.Instances[j], reservation.Instances[i]
		})

		for _, instance := range reservation.Instances {

			if instance.Platform != nil {
				if *instance.Platform == "windows" {
					continue
				}
			}

			instances = append(instances, instance.InstanceId)
			if len(instances) == targetLimit {
				return
			}
		}
	}

	if len(instances) == 0 {
		err = fmt.Errorf("no instances found for targets")
	}
	return
}

func toFilter(targets []*ssm.Target) (filters []*ec2.Filter) {
	for _, target := range targets {
		filters = append(filters, &ec2.Filter{
			Name:   target.Key,
			Values: target.Values,
		})
	}
	return filters
}

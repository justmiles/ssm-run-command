package command

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	invocations      Invocations
	invocationErrors []error
}

// Run command and stream results to stdout
func (c *Command) Run() (int, error) {

	if err := c.RunCommand(); err != nil {
		return 1, err
	}

	// poll for invocations
	for {
		if err := c.UpdateStatus(); err != nil {
			return 1, err
		}

		if err := c.UpdateInvocationList(); err != nil {
			return 1, err
		}

		if c.invocations.areDone() && c.Done() {
			break
		}

		for _, invocation := range c.invocations {

			// Start streaming stdout/stderr
			if !invocation.streaming {
				invocation.streaming = true
				invocation.Stream()
			}

		}

		time.Sleep(randomSeconds(3))

	}

	if err := c.UpdateStatus(); err != nil {
		return 1, err
	}

	for _, invocation := range c.invocations {
		if *invocation.commandInvocation.Status != ssm.CommandInvocationStatusSuccess {
			exitCode = 1
		}
	}

	return exitCode, nil
}

// RunCommand against EC2 instances
func (c *Command) RunCommand() (err error) {
	input := ssm.SendCommandInput{
		TimeoutSeconds: aws.Int64(30),
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

	targets, err := c.targets()
	if err != nil {
		return err
	}

	// If limiting the number of instances, provide instance ID directly
	if c.TargetLimit == 0 {
		input.Targets = targets
	} else {
		input.InstanceIds, err = randomTargets(targets, c.TargetLimit)
		if err != nil {
			return err
		}
	}

	output, err := ssmSvc.SendCommand(&input)
	if err != nil {
		return fmt.Errorf("Error invoking SendCommand: %s", err)
	}

	c.SSMCommand = output.Command

	return nil
}

// UpdateStatus of currently running command
func (c *Command) UpdateStatus() error {
	if c.SSMCommand == nil {
		return errors.New("command not yet executed")
	}

	update, err := ssmSvc.ListCommands(&ssm.ListCommandsInput{
		CommandId:  c.SSMCommand.CommandId,
		MaxResults: aws.Int64(1),
	})
	if err != nil {
		return fmt.Errorf("unable to update command invocation status: %s", err)
	}

	for _, command := range update.Commands {
		c.SSMCommand = command
	}

	return nil
}

// Status returns a frendly status to show to user
func (c *Command) Status() string {

	switch *c.SSMCommand.StatusDetails {
	case "NoInstancesInTag":
		return "no instances matched your targets"
	case "Success":
		return "command exited successfully"
	default:
		return *c.SSMCommand.StatusDetails
	}
}

// Done determine whether or not the command is finished
func (c *Command) Done() bool {

	switch *c.SSMCommand.StatusDetails {
	case "NoInstancesInTag":
		return true
	case "Success":
		return true
	default:
		return false
	}
}

// UpdateInvocationList of currently running commands
func (c *Command) UpdateInvocationList() error {

	if c.SSMCommand == nil {
		return errors.New("command not yet executed")
	}

	if c.invocations == nil {
		c.invocations = make(map[string]*Invocation)
	}

	output, err := ssmSvc.ListCommandInvocations(&ssm.ListCommandInvocationsInput{
		CommandId: c.SSMCommand.CommandId,
	})

	for _, invocation := range output.CommandInvocations {

		if _, ok := c.invocations[*invocation.InstanceId]; !ok {
			c.invocations[*invocation.InstanceId] = &Invocation{}
		}

		c.invocations[*invocation.InstanceId].commandInvocation = invocation
	}

	return err
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

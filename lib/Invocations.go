package command

import (
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Invocations is a map of invocations by EC2 ID
type Invocations map[string]*Invocation

// areDone is true if all invocations are completed
func (invocations *Invocations) areDone() bool {
	var completedCount = 0

	for _, invocation := range *invocations {

		switch *invocation.commandInvocation.Status {

		case ssm.CommandInvocationStatusPending:
			continue

		case ssm.CommandInvocationStatusInProgress:
			continue

		case ssm.CommandInvocationStatusDelayed:
			continue

		case ssm.CommandInvocationStatusCancelled:
			completedCount++

		case ssm.CommandInvocationStatusCancelling:
			continue

		case ssm.CommandInvocationStatusFailed:
			completedCount++

		case ssm.CommandInvocationStatusSuccess:
			completedCount++

		case ssm.CommandInvocationStatusTimedOut:
			completedCount++

		}

	}

	return len(*invocations) == completedCount
}

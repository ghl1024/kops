/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsmodel

import (
	"strings"

	"k8s.io/kops/pkg/model"

	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	NTHTemplate = `{
		"Version": "2012-10-17",
		"Statement": [{                     
			"Effect": "Allow",
			"Principal": {
				"Service": ["events.amazonaws.com", "sqs.amazonaws.com"]
			},
			"Action": "sqs:SendMessage",
			"Resource": [
				"arn:aws:sqs:{{ AWS_REGION }}:{{ ACCOUNT_ID }}:{{ SQS_QUEUE_NAME }}"
			]
		}]
	}`
	DefaultMessageRetentionPeriod = 300
)

type event struct {
	name    string
	pattern string
}

var (
	_ fi.ModelBuilder = &NodeTerminationHandlerBuilder{}

	events = []event{
		{
			name:    "ASGLifecycle",
			pattern: `{"source":["aws.autoscaling"],"detail-type":["EC2 Instance-terminate Lifecycle Action"]}`,
		},
		{
			name:    "SpotInterruption",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Spot Instance Interruption Warning"]}`,
		},
		{
			name:    "RebalanceRecommendation",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance Rebalance Recommendation"]}`,
		},
	}
)

type NodeTerminationHandlerBuilder struct {
	*AWSModelContext

	Lifecycle *fi.Lifecycle
}

func (b *NodeTerminationHandlerBuilder) Build(c *fi.ModelBuilderContext) error {

	for _, ig := range b.InstanceGroups {
		err := b.configureASG(c, ig)
		if err != nil {
			return err
		}
	}

	err := b.buildSQSQueue(c)
	if err != nil {
		return err
	}

	err = b.buildEventBridgeRules(c)
	if err != nil {
		return err
	}

	return nil
}

func (b *NodeTerminationHandlerBuilder) configureASG(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) error {
	name := ig.Name + "-NTHLifecycleHook"

	tags := b.CloudTags(name, false)
	tags["aws-node-termination-handler/managed"] = ""

	lifecyleTask := &awstasks.AutoscalingLifecycleHook{
		ID:                  aws.String(name),
		Name:                aws.String(name),
		Lifecycle:           b.Lifecycle,
		AutoscalingGroup:    b.LinkToAutoscalingGroup(ig),
		DefaultResult:       aws.String("CONTINUE"),
		HeartbeatTimeout:    aws.Int64(DefaultMessageRetentionPeriod),
		LifecycleTransition: aws.String("autoscaling:EC2_INSTANCE_TERMINATING"),
	}

	c.AddTask(lifecyleTask)

	return nil
}

func (b *NodeTerminationHandlerBuilder) buildSQSQueue(c *fi.ModelBuilderContext) error {
	queueName := model.QueueNamePrefix(b.ClusterName()) + "-nth"
	policy := strings.ReplaceAll(NTHTemplate, "{{ AWS_REGION }}", b.Region)
	policy = strings.ReplaceAll(policy, "{{ ACCOUNT_ID }}", b.AWSAccountID)
	policy = strings.ReplaceAll(policy, "{{ SQS_QUEUE_NAME }}", queueName)

	task := &awstasks.SQS{
		Name:                   aws.String(queueName),
		Lifecycle:              b.Lifecycle,
		Policy:                 fi.NewStringResource(policy),
		MessageRetentionPeriod: DefaultMessageRetentionPeriod,
		Tags:                   b.CloudTags(queueName, false),
	}

	c.AddTask(task)

	return nil
}

func (b *NodeTerminationHandlerBuilder) buildEventBridgeRules(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()
	queueName := model.QueueNamePrefix(clusterName) + "-nth"
	region := b.Region
	accountID := b.AWSAccountID
	targetArn := "arn:aws:sqs:" + region + ":" + accountID + ":" + queueName

	for _, event := range events {
		// build rule
		ruleName := aws.String(clusterName + "-" + event.name)
		pattern := event.pattern

		ruleTask := &awstasks.EventBridgeRule{
			Name:      ruleName,
			Lifecycle: b.Lifecycle,
			Tags:      b.CloudTags(*ruleName, false),

			EventPattern: &pattern,
			TargetArn:    &targetArn,
		}

		c.AddTask(ruleTask)

		// build target
		targetTask := &awstasks.EventBridgeTarget{
			Name:      aws.String(*ruleName + "-Target"),
			Lifecycle: b.Lifecycle,

			Rule:      ruleTask,
			TargetArn: &targetArn,
		}

		c.AddTask(targetTask)
	}

	return nil
}

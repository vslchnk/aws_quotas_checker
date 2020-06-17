package cloudformation

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type getUsageFunc func(client *cloudformation.CloudFormation) (*int, error)

type usageFuncMap map[string]getUsageFunc

type Cloudformation struct {
	region        string
	client        *cloudformation.CloudFormation
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates CloudFormation agent
func NewCloudformation(region string, allowedQuotas *[]string) (*Cloudformation, error) {
	c := Cloudformation{}
	c.region = region
	c.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(c.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	c.client = cloudformation.New(sess)
	c.usageFuncs = createUsageFuncMap()

	return &c, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (c *Cloudformation) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *c.usageFuncs {
		isQuotaAllowed := true
		if c.allowedQuotas != nil && !utils.Find(*c.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := v(c.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of Cloudformation quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-0485CB21"] = getUsageFunc(getUsageStacks)
	usageFuncs["L-31709F13"] = getUsageFunc(getUsageStackSets)

	return &usageFuncs
}

func getUsageStacks(client *cloudformation.CloudFormation) (*int, error) {
	stacks := make([]*cloudformation.Stack, 0, 0)

	err := client.DescribeStacksPages(nil,
		func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
			stacks = append(stacks, page.Stacks...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing cloudformation stacks: %v", err)
	}

	l := len(stacks)

	return &l, nil
}

func getUsageStackSets(client *cloudformation.CloudFormation) (*int, error) {
	sets := make([]*cloudformation.StackSetSummary, 0, 0)

	params := &cloudformation.ListStackSetsInput{
		Status: aws.String("ACTIVE"),
	}

	err := client.ListStackSetsPages(params,
		func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
			sets = append(sets, page.Summaries...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing cloudformation stack sets: %v", err)
	}

	l := len(sets)

	return &l, nil
}

// returns code for service in quotas
func GetCode() string {
	return "cloudformation"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-0485CB21"] = "cloudformation:DescribeStacks"
	actions["L-31709F13"] = "cloudformation:ListStackSets"

	return actions
}

package autoscaling

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type getUsageFunc func(client *autoscaling.AutoScaling) (*int, error)

type usageFuncMap map[string]getUsageFunc

type Autoscaling struct {
	region        string
	client        *autoscaling.AutoScaling
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates autoscaling agent
func NewAutoscaling(region string, allowedQuotas *[]string) (*Autoscaling, error) {
	a := Autoscaling{}
	a.region = region
	a.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(a.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	a.client = autoscaling.New(sess)
	a.usageFuncs = createUsageFuncMap()

	return &a, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (a *Autoscaling) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *a.usageFuncs {
		isQuotaAllowed := true
		if a.allowedQuotas != nil && !utils.Find(*a.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := v(a.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of Autoscaling quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-CDE20ADC"] = getUsageFunc(getUsageAutoscalingGroups)
	usageFuncs["L-6B80B8FA"] = getUsageFunc(getUsageLaunchConfigurations)

	return &usageFuncs
}

func getUsageAutoscalingGroups(client *autoscaling.AutoScaling) (*int, error) {
	groups := make([]*autoscaling.Group, 0, 0)

	err := client.DescribeAutoScalingGroupsPages(nil,
		func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			groups = append(groups, page.AutoScalingGroups...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing autoscaling groups: %v", err)
	}

	l := len(groups)

	return &l, nil
}

func getUsageLaunchConfigurations(client *autoscaling.AutoScaling) (*int, error) {
	configurations := make([]*autoscaling.LaunchConfiguration, 0, 0)

	err := client.DescribeLaunchConfigurationsPages(nil,
		func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
			configurations = append(configurations, page.LaunchConfigurations...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing launch configurations: %v", err)
	}

	l := len(configurations)

	return &l, nil
}

// returns code for service in quotas
func GetCode() string {
	return "autoscaling"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-CDE20ADC"] = "autoscaling:DescribeAutoScalingGroups"
	actions["L-6B80B8FA"] = "autoscaling:DescribeLaunchConfigurations"

	return actions
}

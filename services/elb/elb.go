package elb

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type usageFuncMap map[string]string

type ELB struct {
	region        string
	clientv2      *elbv2.ELBV2
	clientv1      *elb.ELB
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates ELB agent
func NewELB(region string, allowedQuotas *[]string) (*ELB, error) {
	e := ELB{}
	e.region = region
	e.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(e.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	e.clientv2 = elbv2.New(sess)
	e.clientv1 = elb.New(sess)
	e.usageFuncs = createUsageFuncMap()

	return &e, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (e *ELB) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *e.usageFuncs {
		isQuotaAllowed := true
		if e.allowedQuotas != nil && !utils.Find(*e.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			var usage *int
			var err error

			if v == "classic" {
				usage, err = getUsageClassicELBs(e.clientv1)
			} else if v == "application" {
				usage, err = getUsageApplicationELBs(e.clientv2)
			}

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of ELB quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-E9E9831D"] = "classic"
	usageFuncs["L-53DA6B97"] = "application"

	return &usageFuncs
}

func getUsageClassicELBs(client *elb.ELB) (*int, error) {
	elbs := make([]*elb.LoadBalancerDescription, 0, 0)

	err := client.DescribeLoadBalancersPages(nil,
		func(page *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
			elbs = append(elbs, page.LoadBalancerDescriptions...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing classic ELBs: %v", err)
	}

	l := len(elbs)

	return &l, nil
}

func getUsageApplicationELBs(client *elbv2.ELBV2) (*int, error) {
	count := 0
	elbs := make([]*elbv2.LoadBalancer, 0, 0)

	err := client.DescribeLoadBalancersPages(nil,
		func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
			elbs = append(elbs, page.LoadBalancers...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing application ELBs: %v", err)
	}

	for _, elb := range elbs {
		if *elb.Type == "application" {
			count++
		}
	}

	return &count, nil
}

// returns code for service in quotas
func GetCode() string {
	return "elasticloadbalancing"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-E9E9831D"] = "elasticloadbalancing:DescribeLoadBalancers"
	actions["L-53DA6B97"] = "elasticloadbalancing:DescribeLoadBalancers"

	return actions
}

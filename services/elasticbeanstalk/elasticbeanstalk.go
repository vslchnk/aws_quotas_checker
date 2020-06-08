package elasticbeanstalk

import (
	"checker/utils"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

type getUsageFunc func(client *elasticbeanstalk.ElasticBeanstalk) (*int, error)

type usageFuncMap map[string]getUsageFunc

type Elasticbeanstalk struct {
	region        string
	client        *elasticbeanstalk.ElasticBeanstalk
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates Elasticbeanstalk agent
func NewElasticbeanstalk(region string, allowedQuotas *[]string) (*Elasticbeanstalk, error) {
	e := Elasticbeanstalk{}
	e.region = region
	e.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(e.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	e.client = elasticbeanstalk.New(sess)
	e.usageFuncs = createUsageFuncMap()

	return &e, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (e *Elasticbeanstalk) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *e.usageFuncs {
		isQuotaAllowed := true
		if e.allowedQuotas != nil && !utils.Find(*e.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := v(e.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of ElasticBeanstalk quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-D64F1F14"] = getUsageFunc(getUsageApplicationVersions)
	usageFuncs["L-1CEABD17"] = getUsageFunc(getUsageApplications)
	usageFuncs["L-8EFC1C51"] = getUsageFunc(getUsageEnvironments)

	return &usageFuncs
}

func getUsageApplicationVersions(client *elasticbeanstalk.ElasticBeanstalk) (*int, error) {
	versions := make([]*elasticbeanstalk.ApplicationVersionDescription, 0, 0)

	res, err := client.DescribeApplicationVersions(nil)
	if err != nil {
		return nil, fmt.Errorf("Error while describing ElasticBeanstalk application versions: %v", err)
	}

	versions = append(versions, res.ApplicationVersions...)
	nextToken := res.NextToken

	for nextToken != nil {
		params := &elasticbeanstalk.DescribeApplicationVersionsInput{
			NextToken: nextToken,
		}

		res, err = client.DescribeApplicationVersions(params)
		if err != nil {
			return nil, fmt.Errorf("Error while describing ElasticBeanstalk application versions: %v", err)
		}

		versions = append(versions, res.ApplicationVersions...)
		nextToken = res.NextToken
	}

	l := len(versions)

	return &l, nil
}

func getUsageApplications(client *elasticbeanstalk.ElasticBeanstalk) (*int, error) {
	res, err := client.DescribeApplications(nil)

	if err != nil {
		return nil, fmt.Errorf("Error while describing ElasticBeanstalk applications: %v", err)
	}

	l := len(res.Applications)

	return &l, nil
}

func getUsageEnvironments(client *elasticbeanstalk.ElasticBeanstalk) (*int, error) {
	environments := make([]*elasticbeanstalk.EnvironmentDescription, 0, 0)

	res, err := client.DescribeEnvironments(nil)
	if err != nil {
		return nil, fmt.Errorf("Error while describing ElasticBeanstalk environments: %v", err)
	}

	environments = append(environments, res.Environments...)
	nextToken := res.NextToken

	for nextToken != nil {
		params := &elasticbeanstalk.DescribeEnvironmentsInput{
			NextToken: nextToken,
		}

		res, err = client.DescribeEnvironments(params)
		if err != nil {
			return nil, fmt.Errorf("Error while describing ElasticBeanstalk environments: %v", err)
		}

		environments = append(environments, res.Environments...)
		nextToken = res.NextToken
	}

	l := len(environments)

	return &l, nil
}

// returns code for service in quotas
func GetCode() string {
	return "elasticbeanstalk"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-D64F1F14"] = "elasticbeanstalk:DescribeApplicationVersions"
	actions["L-1CEABD17"] = "elasticbeanstalk:DescribeApplications"
	actions["L-8EFC1C51"] = "elasticbeanstalk:DescribeEnvironments"

	return actions
}

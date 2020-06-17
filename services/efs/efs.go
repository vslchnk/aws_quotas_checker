package efs

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
)

type getUsageFunc func(client *efs.EFS) (*int, error)

type usageFuncMap map[string]getUsageFunc

type EFS struct {
	region        string
	client        *efs.EFS
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates EFS agent
func NewEFS(region string, allowedQuotas *[]string) (*EFS, error) {
	e := EFS{}
	e.region = region
	e.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(e.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	e.client = efs.New(sess)
	e.usageFuncs = createUsageFuncMap()

	return &e, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (e *EFS) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *e.usageFuncs {
		isQuotaAllowed := true
		if e.allowedQuotas != nil && !utils.Find(*e.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := v(e.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of EFS quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-848C634D"] = getUsageFunc(getUsageFileSystems)

	return &usageFuncs
}

func getUsageFileSystems(client *efs.EFS) (*int, error) {
	fs := make([]*efs.FileSystemDescription, 0, 0)

	err := client.DescribeFileSystemsPages(nil,
		func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
			fs = append(fs, page.FileSystems...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing elastic file systems: %v", err)
	}

	l := len(fs)

	return &l, nil
}

// returns code for service in quotas
func GetCode() string {
	return "elasticfilesystem"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-848C634D"] = "elasticfilesystem:DescribeFileSystems"

	return actions
}

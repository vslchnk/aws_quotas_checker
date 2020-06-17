package s3

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/sts"
)

type usageFuncMap map[string]string

type S3 struct {
	region          string
	session         *session.Session
	clientS3        *s3.S3
	clientS3Control *s3control.S3Control
	usageFuncs      *usageFuncMap
	allowedQuotas   *[]string
}

// creates S3 agent
func NewS3(region string, allowedQuotas *[]string) (*S3, error) {
	s := S3{}
	s.region = region
	s.allowedQuotas = allowedQuotas

	var err error
	s.session, err = session.NewSession(&aws.Config{
		Region: aws.String(s.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	s.clientS3 = s3.New(s.session)
	s.clientS3Control = s3control.New(s.session)
	s.usageFuncs = createUsageFuncMap()

	return &s, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (s *S3) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *s.usageFuncs {
		isQuotaAllowed := true
		if s.allowedQuotas != nil && !utils.Find(*s.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			var usage *int
			var err error

			if v == "s3" {
				usage, err = getUsageBuckets(s.clientS3)
			} else if v == "s3control" {
				usage, err = getUsageAcessPoints(s.clientS3Control, s.session)
			}

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

	usageFuncs["L-DC2B2D3D"] = "s3"
	usageFuncs["L-FAABEEBA"] = "s3control"

	return &usageFuncs
}

func getUsageBuckets(client *s3.S3) (*int, error) {
	res, err := client.ListBuckets(nil)

	if err != nil {
		return nil, fmt.Errorf("Error while describing S3 buckets: %v", err)
	}

	l := len(res.Buckets)

	return &l, nil
}

func getUsageAcessPoints(client *s3control.S3Control, session *session.Session) (*int, error) {
	stsClient := sts.New(session)

	res, err := stsClient.GetCallerIdentity(nil)
	if err != nil {
		return nil, fmt.Errorf("Error while getting caller identity: %v", err)
	}

	points := make([]*s3control.AccessPoint, 0, 0)

	params := &s3control.ListAccessPointsInput{
		AccountId: res.Account,
	}

	err = client.ListAccessPointsPages(params,
		func(page *s3control.ListAccessPointsOutput, lastPage bool) bool {
			points = append(points, page.AccessPointList...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing S3 access points: %v", err)
	}

	l := len(points)

	return &l, nil
}

// returns code for service in quotas
func GetCode() string {
	return "s3"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-DC2B2D3D"] = "s3:ListAllMyBuckets"
	actions["L-FAABEEBA"] = "s3:ListAccessPoints; sts:GetCallerIdentity"

	return actions
}

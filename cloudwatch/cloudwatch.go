package cloudwatch

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/servicequotas"
)

type CW struct {
	region    string
	client    *cloudwatch.CloudWatch
	period    int64
	threshold int64
}

// creates new CW agent
func NewCW(region string) (*CW, error) {
	c := CW{}
	c.region = region
	c.period = 300
	c.threshold = 5

	ses, err := session.NewSession(&aws.Config{
		Region: aws.String(c.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	c.client = cloudwatch.New(ses)

	return &c, nil
}

// returns usage from for metric for the last 5 minutes
func (c *CW) GetUsageFromMetric(usageMetric *servicequotas.MetricInfo) (*int, error) {
	dimensions := make([]*cloudwatch.Dimension, 0, len(usageMetric.MetricDimensions))

	for k, v := range usageMetric.MetricDimensions {
		d := &cloudwatch.Dimension{}
		name := k
		d.Name = &name
		d.Value = v
		dimensions = append(dimensions, d)
	}

	loc, err := time.LoadLocation("UTC")

	if err != nil {
		return nil, fmt.Errorf("Error while loading time location: %v", err)
	}

	endTime := time.Now().In(loc)
	startTime := endTime.Add(time.Duration(-c.threshold) * time.Minute)

	params := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: dimensions,
		MetricName: usageMetric.MetricName,
		Namespace:  usageMetric.MetricNamespace,
		Statistics: []*string{usageMetric.MetricStatisticRecommendation},
		StartTime:  &startTime,
		EndTime:    &endTime,
		Period:     &c.period,
	}

	data, err := c.client.GetMetricStatistics(params)

	if err != nil {
		return nil, fmt.Errorf("Error while getting metric statistics: %v", err)
	}

	usage := 0

	if data.Datapoints != nil {
		if *usageMetric.MetricStatisticRecommendation == "Sum" {
			usage = int(*data.Datapoints[0].Sum)
		} else if *usageMetric.MetricStatisticRecommendation == "Maximum" {
			usage = int(*data.Datapoints[0].Maximum)
		}
	}

	return &usage, nil
}

// returns actions for IAM policy which allow to work with this package
func GetIam() []string {
	actions := []string{
		"cloudwatch:GetMetricStatistics",
	}

	return actions
}

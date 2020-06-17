package vpc

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type getUsageFunc func(client *ec2.EC2) (*int, error)

type usageFuncMap map[string]getUsageFunc

type VPC struct {
	region        string
	client        *ec2.EC2
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates S3 agent
func NewVPC(region string, allowedQuotas *[]string) (*VPC, error) {
	v := VPC{}
	v.region = region
	v.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(v.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	v.client = ec2.New(sess)
	v.usageFuncs = createUsageFuncMap()

	return &v, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (v *VPC) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, val := range *v.usageFuncs {
		isQuotaAllowed := true
		if v.allowedQuotas != nil && !utils.Find(*v.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := val(v.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of VPC quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-1B52E74A"] = getUsageFunc(getUsageGatewayVpcEndPoint)
	usageFuncs["L-29B6F2EB"] = getUsageFunc(getUsageInterfaceVpcEndPoint)
	usageFuncs["L-45FE3B85"] = getUsageFunc(getUsageEgressOnlyInternetGateways)
	usageFuncs["L-A4707A72"] = getUsageFunc(getUsageInternetGateways)
	usageFuncs["L-B4A6D682"] = getUsageFunc(getUsageNetworkAcls)
	usageFuncs["L-DF5E4CA3"] = getUsageFunc(getUsageNetworkInterfaces)
	usageFuncs["L-E79EC296"] = getUsageFunc(getUsageSecurityGroups)
	usageFuncs["L-F678F1CE"] = getUsageFunc(getUsageVpcs)

	return &usageFuncs
}

func getUsageGatewayVpcEndPoint(client *ec2.EC2) (*int, error) {
	return getUsageVpcEndpoints(client, "Gateway")
}

func getUsageInterfaceVpcEndPoint(client *ec2.EC2) (*int, error) {
	return getUsageVpcEndpoints(client, "Interface")
}

func getUsageEgressOnlyInternetGateways(client *ec2.EC2) (*int, error) {
	gateways := make([]*ec2.EgressOnlyInternetGateway, 0, 0)

	err := client.DescribeEgressOnlyInternetGatewaysPages(nil,
		func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
			gateways = append(gateways, page.EgressOnlyInternetGateways...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing egress only internet gateways: %v", err)
	}

	l := len(gateways)

	return &l, nil
}

func getUsageInternetGateways(client *ec2.EC2) (*int, error) {
	gateways := make([]*ec2.InternetGateway, 0, 0)

	err := client.DescribeInternetGatewaysPages(nil,
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			gateways = append(gateways, page.InternetGateways...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing internet gateways: %v", err)
	}

	l := len(gateways)

	return &l, nil
}

func getUsageNetworkAcls(client *ec2.EC2) (*int, error) {
	acls := make([]*ec2.NetworkAcl, 0, 0)

	err := client.DescribeNetworkAclsPages(nil,
		func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
			acls = append(acls, page.NetworkAcls...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing network acls: %v", err)
	}

	l := len(acls)

	return &l, nil
}

func getUsageNetworkInterfaces(client *ec2.EC2) (*int, error) {
	interfaces := make([]*ec2.NetworkInterface, 0, 0)

	err := client.DescribeNetworkInterfacesPages(nil,
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			interfaces = append(interfaces, page.NetworkInterfaces...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing network interfaces: %v", err)
	}

	l := len(interfaces)

	return &l, nil
}

func getUsageSecurityGroups(client *ec2.EC2) (*int, error) {
	sgs := make([]*ec2.SecurityGroup, 0, 0)

	err := client.DescribeSecurityGroupsPages(nil,
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			sgs = append(sgs, page.SecurityGroups...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing security groups: %v", err)
	}

	l := len(sgs)

	return &l, nil
}

func getUsageVpcs(client *ec2.EC2) (*int, error) {
	vpcs := make([]*ec2.Vpc, 0, 0)

	err := client.DescribeVpcsPages(nil,
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing VPCs: %v", err)
	}

	l := len(vpcs)

	return &l, nil
}

func getUsageVpcEndpoints(client *ec2.EC2, gType string) (*int, error) {
	count := 0
	endpoints := make([]*ec2.VpcEndpoint, 0, 0)

	err := client.DescribeVpcEndpointsPages(nil,
		func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
			endpoints = append(endpoints, page.VpcEndpoints...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing vpc endpoints: %v", err)
	}

	for _, endpoint := range endpoints {
		if *endpoint.VpcEndpointType == gType {
			count++
		}
	}

	return &count, nil
}

// returns code for service in quotas
func GetCode() string {
	return "vpc"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-1B52E74A"] = "ec2:DescribeVpcEndpoints"
	actions["L-29B6F2EB"] = "ec2:DescribeVpcEndpoints"
	actions["L-45FE3B85"] = "ec2:DescribeEgressOnlyInternetGateways"
	actions["L-A4707A72"] = "ec2:DescribeInternetGateways"
	actions["L-B4A6D682"] = "ec2:DescribeNetworkAcls"
	actions["L-DF5E4CA3"] = "ec2:DescribeNetworkInterfaces"
	actions["L-E79EC296"] = "ec2:DescribeSecurityGroups"
	actions["L-F678F1CE"] = "ec2:DescribeVpcs"

	return actions
}

package ec2

import (
	"fmt"
	"strings"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type getUsageFunc func(client *ec2.EC2) (*int, error)

type usageFuncMap map[string]getUsageFunc

type EC2 struct {
	region        string
	client        *ec2.EC2
	usageFuncs    *usageFuncMap
	allowedQuotas *[]string
}

// creates EC2 agent
func NewEC2(region string, allowedQuotas *[]string) (*EC2, error) {
	e := EC2{}
	e.region = region
	e.allowedQuotas = allowedQuotas

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(e.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	e.client = ec2.New(sess)
	e.usageFuncs = createUsageFuncMap()

	return &e, nil
}

// returns map of usage where key is the quotas code and value is the usage
func (e *EC2) GetUsage() (*map[string]int, error) {
	usageMap := make(map[string]int)

	for k, v := range *e.usageFuncs {
		isQuotaAllowed := true
		if e.allowedQuotas != nil && !utils.Find(*e.allowedQuotas, k) {
			isQuotaAllowed = false
		}
		if isQuotaAllowed {
			usage, err := v(e.client)

			if err != nil {
				return nil, fmt.Errorf("Error while getting usage of EC2 quota: %v", err)
			}

			usageMap[k] = *usage
		}
	}

	return &usageMap, nil
}

// creates map to associate quota code with func to calculate its usage
func createUsageFuncMap() *usageFuncMap {
	usageFuncs := make(usageFuncMap)

	usageFuncs["L-949445B0"] = getUsageFunc(getUsageDedicatedA1Hosts)
	usageFuncs["L-E4BF28E0"] = getUsageFunc(getUsageDedicatedC4Hosts)
	usageFuncs["L-81657574"] = getUsageFunc(getUsageDedicatedC5Hosts)
	usageFuncs["L-C93F66A2"] = getUsageFunc(getUsageDedicatedC5DHosts)
	usageFuncs["L-20F13EBD"] = getUsageFunc(getUsageDedicatedC5NHosts)
	usageFuncs["L-8B27377A"] = getUsageFunc(getUsageDedicatedD2Hosts)
	usageFuncs["L-DE82EABA"] = getUsageFunc(getUsageDedicatedG3Hosts)
	usageFuncs["L-9675FDCD"] = getUsageFunc(getUsageDedicatedG3SHosts)
	usageFuncs["L-CAE24619"] = getUsageFunc(getUsageDedicatedG4DNHosts)
	usageFuncs["L-84391ECC"] = getUsageFunc(getUsageDedicatedH1Hosts)
	usageFuncs["L-6222C1B6"] = getUsageFunc(getUsageDedicatedI2Hosts)
	usageFuncs["L-77EE2B11"] = getUsageFunc(getUsageDedicatedI3ENHosts)
	usageFuncs["L-EF30B25E"] = getUsageFunc(getUsageDedicatedM4Hosts)
	usageFuncs["L-8B7BF662"] = getUsageFunc(getUsageDedicatedM5Hosts)
	usageFuncs["L-B10F70D6"] = getUsageFunc(getUsageDedicatedM5AHosts)
	usageFuncs["L-8CCBD91B"] = getUsageFunc(getUsageDedicatedM5DHosts)
	usageFuncs["L-DA07429F"] = getUsageFunc(getUsageDedicatedM5DNHosts)
	usageFuncs["L-D50A37FA"] = getUsageFunc(getUsageDedicatedM6GHosts)
	usageFuncs["L-2753CF59"] = getUsageFunc(getUsageDedicatedP2Hosts)
	usageFuncs["L-A0A19F79"] = getUsageFunc(getUsageDedicatedP3Hosts)
	usageFuncs["L-B7208018"] = getUsageFunc(getUsageDedicatedR3Hosts)
	usageFuncs["L-313524BA"] = getUsageFunc(getUsageDedicatedR4Hosts)
	usageFuncs["L-EA4FD6CF"] = getUsageFunc(getUsageDedicatedR5Hosts)
	usageFuncs["L-8FE30D52"] = getUsageFunc(getUsageDedicatedR5AHosts)
	usageFuncs["L-EC7178B6"] = getUsageFunc(getUsageDedicatedR5ADHosts)
	usageFuncs["L-8814B54F"] = getUsageFunc(getUsageDedicatedR5DHosts)
	usageFuncs["L-4AB14223"] = getUsageFunc(getUsageDedicatedR5DNHosts)
	usageFuncs["L-52EF324A"] = getUsageFunc(getUsageDedicatedR5NHosts)
	usageFuncs["L-DE3D9563"] = getUsageFunc(getUsageDedicatedX1Hosts)
	usageFuncs["L-DEF8E115"] = getUsageFunc(getUsageDedicatedX1EHosts)
	usageFuncs["L-F035E935"] = getUsageFunc(getUsageDedicatedZ1DHosts)
	usageFuncs["L-7029FAB6"] = getUsageFunc(getUsageVpnGateways)
	usageFuncs["L-3E6EC3A3"] = getUsageFunc(getUsageVpnConnections)
	usageFuncs["L-A2478D36"] = getUsageFunc(getUsageTransitGateways)
	usageFuncs["L-4FB7FF5D"] = getUsageFunc(getUsageCustomerGateways)
	usageFuncs["L-0263D0A3"] = getUsageFunc(getUsageEIPVPC)

	return &usageFuncs
}

func getUsageDedicatedA1Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "a1")
}

func getUsageDedicatedC4Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "c4")
}

func getUsageDedicatedC5Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "c5")
}

func getUsageDedicatedC5DHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "c5d")
}

func getUsageDedicatedC5NHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "c5n")
}

func getUsageDedicatedD2Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "d2")
}

func getUsageDedicatedG3Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "g3")
}

func getUsageDedicatedG3SHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "g3s")
}

func getUsageDedicatedG4DNHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "g4dn")
}

func getUsageDedicatedH1Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "h1")
}

func getUsageDedicatedI2Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "i2")
}

func getUsageDedicatedI3Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "i3")
}

func getUsageDedicatedI3ENHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "i3en")
}

func getUsageDedicatedM4Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m4")
}

func getUsageDedicatedM5Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5")
}

func getUsageDedicatedM5AHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5a")
}

func getUsageDedicatedM5ADHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5ad")
}

func getUsageDedicatedM5DHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5d")
}

func getUsageDedicatedM5DNHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5dn")
}

func getUsageDedicatedM5NHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m5n")
}

func getUsageDedicatedM6GHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "m6g")
}

func getUsageDedicatedP2Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "p2")
}

func getUsageDedicatedP3Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "p3")
}

func getUsageDedicatedR3Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r3")
}

func getUsageDedicatedR4Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r4")
}

func getUsageDedicatedR5Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5")
}

func getUsageDedicatedR5AHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5a")
}

func getUsageDedicatedR5ADHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5ad")
}

func getUsageDedicatedR5DHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5d")
}

func getUsageDedicatedR5DNHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5dn")
}

func getUsageDedicatedR5NHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "r5n")
}

func getUsageDedicatedX1Hosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "x1")
}

func getUsageDedicatedX1EHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "x1e")
}

func getUsageDedicatedZ1DHosts(client *ec2.EC2) (*int, error) {
	return getDedicatedHostsUsage(client, "z1d")
}

func getUsageVpnGateways(client *ec2.EC2) (*int, error) {
	res, err := client.DescribeVpnGateways(nil)

	if err != nil {
		return nil, fmt.Errorf("Error while describing VPN gateways: %v", err)
	}

	l := len(res.VpnGateways)

	return &l, nil
}

func getUsageVpnConnections(client *ec2.EC2) (*int, error) {
	res, err := client.DescribeVpnConnections(nil)

	if err != nil {
		return nil, fmt.Errorf("Error while describing VPN connections: %v", err)
	}

	l := len(res.VpnConnections)

	return &l, nil
}

func getUsageTransitGateways(client *ec2.EC2) (*int, error) {
	gateways := make([]*ec2.TransitGateway, 0, 0)

	err := client.DescribeTransitGatewaysPages(nil,
		func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
			gateways = append(gateways, page.TransitGateways...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing transit gateways: %v", err)
	}

	l := len(gateways)

	return &l, nil
}

func getUsageCustomerGateways(client *ec2.EC2) (*int, error) {
	res, err := client.DescribeCustomerGateways(nil)

	if err != nil {
		return nil, fmt.Errorf("Error while describing customer gateways: %v", err)
	}

	l := len(res.CustomerGateways)

	return &l, nil
}

func getUsageEIPVPC(client *ec2.EC2) (*int, error) {
	params := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("domain"),
				Values: []*string{
					aws.String("vpc"),
				},
			},
		},
	}

	res, err := client.DescribeAddresses(params)

	if err != nil {
		return nil, fmt.Errorf("Error while describing EC2-VPC EIPs: %v", err)
	}

	l := len(res.Addresses)

	return &l, nil
}

func getDedicatedHostsUsage(client *ec2.EC2, family string) (*int, error) {
	count := 0
	hosts := make([]*ec2.Host, 0, 0)

	err := client.DescribeHostsPages(nil,
		func(page *ec2.DescribeHostsOutput, lastPage bool) bool {
			hosts = append(hosts, page.Hosts...)
			return true
		})

	if err != nil {
		return nil, fmt.Errorf("Error while describing dedicated hosts: %v", err)
	}

	for _, host := range hosts {
		if (host.HostProperties.InstanceType != nil && strings.Contains(*host.HostProperties.InstanceType, family+".")) || (host.HostProperties.InstanceFamily != nil && *host.HostProperties.InstanceFamily == family) {
			count++
		}
	}

	return &count, nil
}

// returns code for service in quotas
func GetCode() string {
	return "ec2"
}

// returns actions for IAM policy which allow to work with this package, every action is associated with the quota code
func GetIam() map[string]string {
	actions := make(map[string]string)

	actions["L-949445B0"] = "ec2:DescribeHosts"
	actions["L-E4BF28E0"] = "ec2:DescribeHosts"
	actions["L-81657574"] = "ec2:DescribeHosts"
	actions["L-C93F66A2"] = "ec2:DescribeHosts"
	actions["L-20F13EBD"] = "ec2:DescribeHosts"
	actions["L-8B27377A"] = "ec2:DescribeHosts"
	actions["L-DE82EABA"] = "ec2:DescribeHosts"
	actions["L-9675FDCD"] = "ec2:DescribeHosts"
	actions["L-CAE24619"] = "ec2:DescribeHosts"
	actions["L-84391ECC"] = "ec2:DescribeHosts"
	actions["L-6222C1B6"] = "ec2:DescribeHosts"
	actions["L-77EE2B11"] = "ec2:DescribeHosts"
	actions["L-EF30B25E"] = "ec2:DescribeHosts"
	actions["L-8B7BF662"] = "ec2:DescribeHosts"
	actions["L-B10F70D6"] = "ec2:DescribeHosts"
	actions["L-8CCBD91B"] = "ec2:DescribeHosts"
	actions["L-DA07429F"] = "ec2:DescribeHosts"
	actions["L-D50A37FA"] = "ec2:DescribeHosts"
	actions["L-2753CF59"] = "ec2:DescribeHosts"
	actions["L-A0A19F79"] = "ec2:DescribeHosts"
	actions["L-B7208018"] = "ec2:DescribeHosts"
	actions["L-313524BA"] = "ec2:DescribeHosts"
	actions["L-EA4FD6CF"] = "ec2:DescribeHosts"
	actions["L-8FE30D52"] = "ec2:DescribeHosts"
	actions["L-EC7178B6"] = "ec2:DescribeHosts"
	actions["L-8814B54F"] = "ec2:DescribeHosts"
	actions["L-4AB14223"] = "ec2:DescribeHosts"
	actions["L-52EF324A"] = "ec2:DescribeHosts"
	actions["L-DE3D9563"] = "ec2:DescribeHosts"
	actions["L-DEF8E115"] = "ec2:DescribeHosts"
	actions["L-F035E935"] = "ec2:DescribeHosts"
	actions["L-7029FAB6"] = "ec2:DescribeVpnGateways"
	actions["L-3E6EC3A3"] = "ec2:DescribeVpnConnections"
	actions["L-A2478D36"] = "ec2:DescribeTransitGateways"
	actions["L-4FB7FF5D"] = "ec2:DescribeCustomerGateways"
	actions["L-0263D0A3"] = "ec2:DescribeAddresses"

	return actions
}

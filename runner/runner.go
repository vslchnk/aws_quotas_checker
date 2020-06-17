package runner

import (
	"fmt"

	"github.com/vslchnk/aws_quotas_checker/cloudwatch"
	"github.com/vslchnk/aws_quotas_checker/quotas"
	"github.com/vslchnk/aws_quotas_checker/services/autoscaling"
	"github.com/vslchnk/aws_quotas_checker/services/cloudformation"
	"github.com/vslchnk/aws_quotas_checker/services/ec2"
	"github.com/vslchnk/aws_quotas_checker/services/efs"
	"github.com/vslchnk/aws_quotas_checker/services/elasticbeanstalk"
	"github.com/vslchnk/aws_quotas_checker/services/elb"
	"github.com/vslchnk/aws_quotas_checker/services/s3"
	"github.com/vslchnk/aws_quotas_checker/services/vpc"
	"github.com/vslchnk/aws_quotas_checker/utils"
)

type ServiceQuota struct {
	Adjustable   bool
	GlobalQuota  bool
	QuotaName    string
	QuotaCode    string
	DefaultValue float64
	Value        float64
	ServiceCode  string
	ServiceName  string
}

type ServiceQuotaUsage struct {
	ServiceCode string
	ServiceName string
	QuotaName   string
	QuotaCode   string
	Usage       int
	Value       int
	Type        string
}

type Warning struct {
	ServiceCode string
	ServiceName string
	QuotaName   string
	QuotaCode   string
	Limit       int
	Usage       int
	Name        string
	Threshold   int
}

type iamActions map[string][]string

type Runner struct {
	region            string
	allowedServices   *map[string]*[]string
	quotas            *quotas.Quotas
	quotaServiceCodes *[]string
	ec2               *ec2.EC2
	cf                *cloudformation.Cloudformation
	elastic           *elasticbeanstalk.Elasticbeanstalk
	as                *autoscaling.Autoscaling
	s3                *s3.S3
	vpc               *vpc.VPC
	elb               *elb.ELB
	efs               *efs.EFS
	cw                *cloudwatch.CW
	quotaMetricUsage  *map[string]int
	quotaApiUsage     *map[string]int
	quotaUsage        *map[string][]int
	quotasServiceInfo *map[string]string
	alarms            map[string]int
}

// creates runner agent
func NewRunner(region string, allowedServices *map[string]*[]string) (*Runner, error) {
	r := Runner{}
	r.region = region
	r.allowedServices = allowedServices
	r.alarms = make(map[string]int)

	var err error
	r.quotas, err = quotas.NewQuota(r.region, r.allowedServices)
	if err != nil {
		return nil, fmt.Errorf("Error while creating quota client: %v", err)
	}
	r.quotaServiceCodes = r.quotas.ListServicesCodes()

	err = r.createServicesClients()
	if err != nil {
		return nil, fmt.Errorf("Error while creating service clients: %v", err)
	}

	r.cw, err = cloudwatch.NewCW(r.region)
	if err != nil {
		return nil, fmt.Errorf("Error while creating cloudwatch client: %v", err)
	}

	r.quotaMetricUsage, err = r.getQuotaMetricUsage()
	if err != nil {
		return nil, fmt.Errorf("Error while getting usage for quotas from cloudwatch metrics: %v", err)
	}

	r.quotaApiUsage, err = r.getQuotaApiUsage()
	if err != nil {
		return nil, fmt.Errorf("Error while getting usage for quotas from api: %v", err)
	}

	r.quotaUsage = utils.MergeMaps(*r.quotaMetricUsage, *r.quotaApiUsage)
	r.quotasServiceInfo = r.createQuotasServiceInfo()

	return &r, nil
}

// add alarm to alarms map with key-name value-threshold
func (r *Runner) AddAlarm(name string, threshold int) {
	r.alarms[name] = threshold
}

// creates map where key is the quota code and value is the service code
func (r *Runner) createQuotasServiceInfo() *map[string]string {
	quotasServiceInfo := make(map[string]string)

	for _, s := range *r.quotaServiceCodes {
		quotas, _ := r.GetSupportedQuotasForService(s)
		for _, q := range quotas {
			quotasServiceInfo[q] = s
		}
	}

	return &quotasServiceInfo
}

// creates clients to get usage from AWS services
func (r *Runner) createServicesClients() error {
	var err error

	isServiceAllowed, allowedQuotas := r.checkIfServiceAllowed(ec2.GetCode())
	if isServiceAllowed {
		r.ec2, err = ec2.NewEC2(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating ec2 client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(cloudformation.GetCode())
	if isServiceAllowed {
		r.cf, err = cloudformation.NewCloudformation(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating cloudformation client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(elasticbeanstalk.GetCode())
	if isServiceAllowed {
		r.elastic, err = elasticbeanstalk.NewElasticbeanstalk(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating elasticbeanstalk client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(autoscaling.GetCode())
	if isServiceAllowed {
		r.as, err = autoscaling.NewAutoscaling(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating autoscaling client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(s3.GetCode())
	if isServiceAllowed {
		r.s3, err = s3.NewS3(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating S3 client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(vpc.GetCode())
	if isServiceAllowed {
		r.vpc, err = vpc.NewVPC(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating VPC client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(elb.GetCode())
	if isServiceAllowed {
		r.elb, err = elb.NewELB(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating ELB client: %v", err)
		}
	}

	isServiceAllowed, allowedQuotas = r.checkIfServiceAllowed(efs.GetCode())
	if isServiceAllowed {
		r.efs, err = efs.NewEFS(r.region, allowedQuotas)
		if err != nil {
			return fmt.Errorf("Error while creating EFS client: %v", err)
		}
	}

	return nil
}

// checks if service is allowed and returns result and list of allowed quotas
func (r *Runner) checkIfServiceAllowed(serviceCode string) (bool, *[]string) {
	isServiceAllowed := true
	var allowedQuotas *[]string

	if r.allowedServices != nil {
		if !utils.CheckKeyMap(*r.allowedServices, serviceCode) {
			isServiceAllowed = false
		} else {
			allowedQuotas, _ = (*r.allowedServices)[serviceCode]
		}
	}

	return isServiceAllowed, allowedQuotas
}

// returns map with the usage of quotas as a value and quota code as a key, usage is from services API
func (r *Runner) getQuotaApiUsage() (*map[string]int, error) {
	var err error
	var mapPointer *map[string]int

	isServiceAllowed, _ := r.checkIfServiceAllowed(ec2.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.ec2.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for EC2 quotas: %v", err)
		}
	}
	ec2UsageMap := make(map[string]int)
	if mapPointer != nil {
		ec2UsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(cloudformation.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.cf.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for EC2 quotas: %v", err)
		}
	}
	cfUsageMap := make(map[string]int)
	if mapPointer != nil {
		cfUsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(elasticbeanstalk.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.elastic.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for ElasticBeanstalk quotas: %v", err)
		}
	}
	elasticUsageMap := make(map[string]int)
	if mapPointer != nil {
		elasticUsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(autoscaling.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.as.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for AutoScaling quotas: %v", err)
		}
	}
	autoscalingUsageMap := make(map[string]int)
	if mapPointer != nil {
		autoscalingUsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(s3.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.s3.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for S3 quotas: %v", err)
		}
	}
	s3UsageMap := make(map[string]int)
	if mapPointer != nil {
		s3UsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(vpc.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.vpc.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for VPC quotas: %v", err)
		}
	}
	vpcUsageMap := make(map[string]int)
	if mapPointer != nil {
		vpcUsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(elb.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.elb.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for ELB quotas: %v", err)
		}
	}
	elbUsageMap := make(map[string]int)
	if mapPointer != nil {
		elbUsageMap = *mapPointer
	}

	isServiceAllowed, _ = r.checkIfServiceAllowed(efs.GetCode())
	if isServiceAllowed {
		mapPointer, err = r.efs.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("Error while getting usage from API for EFS quotas: %v", err)
		}
	}
	efsUsageMap := make(map[string]int)
	if mapPointer != nil {
		efsUsageMap = *mapPointer
	}

	return utils.MergeMapsUniqueKeys(ec2UsageMap, cfUsageMap, elasticUsageMap, autoscalingUsageMap, s3UsageMap, vpcUsageMap, elbUsageMap, efsUsageMap), nil
}

// returns map with the usage of quotas as a value and quota code as a key, usage is from cloudwatch metrics
func (r *Runner) getQuotaMetricUsage() (*map[string]int, error) {
	quotaMetricUsage := make(map[string]int)

	for _, service := range *r.quotaServiceCodes {
		s, _ := r.quotas.GetService(service)
		quotas, err := r.quotas.ListQuotasCodes(service)
		if err != nil {
			return nil, fmt.Errorf("Error while listing quotas codes: %v", err)
		}

		for _, quota := range *quotas {
			q, _ := s.GetServiceQuota(quota)

			if q.UsageMetric != nil {
				usage, err := r.cw.GetUsageFromMetric(q.UsageMetric)
				if err != nil {
					return nil, fmt.Errorf("Error while getting usage metric: %v", err)
				}

				quotaMetricUsage[quota] = *usage
			}
		}
	}

	return &quotaMetricUsage, nil
}

// updates usage for quotas
func (r *Runner) UpdateQuotasUsage() error {
	var err error
	r.quotaMetricUsage, err = r.getQuotaMetricUsage()
	if err != nil {
		return fmt.Errorf("Error while getting usage for quotas from cloudwatch metrics: %v", err)
	}

	r.quotaApiUsage, err = r.getQuotaApiUsage()
	if err != nil {
		return fmt.Errorf("Error while getting usage for quotas from api: %v", err)
	}

	r.quotaUsage = utils.MergeMaps(*r.quotaMetricUsage, *r.quotaApiUsage)

	return nil
}

// updates info for quotas
func (r *Runner) UpdateQuotasInfo() (err error) {
	r.quotas, err = quotas.NewQuota(r.region, r.allowedServices)
	r.quotaServiceCodes = r.quotas.ListServicesCodes()

	return
}

// returns slice of service codes from quotas
func (r *Runner) ListSupportedServicesFromQuotas() []string {
	return *r.quotaServiceCodes
}

// returns service name by its code
func (r *Runner) GetServiceName(serviceCode string) string {
	s, err := r.quotas.GetService(serviceCode)

	if err != nil {
		utils.ExitErrorf("Error while getting service: %v", err)
	}

	return s.GetServiceName()
}

// returns slice of supported quotas codes for service code
func (r *Runner) GetSupportedQuotasForService(serviceCode string) ([]string, error) {
	res, err := r.quotas.ListQuotasCodes(serviceCode)

	if err != nil {
		return nil, fmt.Errorf("Error while listing quotas: %v", err)
	}

	return *res, nil
}

// returns ServiceQuota object by service code and quota code
func (r *Runner) GetServiceQuota(serviceCode string, quotaCode string) (*ServiceQuota, error) {
	s, err := r.quotas.GetService(serviceCode)

	if err != nil {
		return nil, fmt.Errorf("Error while getting service: %v", err)
	}

	q, err := s.GetServiceQuota(quotaCode)

	if err != nil {
		return nil, fmt.Errorf("Error while getting service quota: %v", err)
	}

	sq := ServiceQuota{}
	sq.Adjustable = *q.Adjustable
	sq.GlobalQuota = *q.GlobalQuota
	sq.QuotaName = *q.QuotaName
	sq.QuotaCode = *q.QuotaCode
	sq.DefaultValue = *q.Value
	sq.Value = *q.ValueApplied
	sq.ServiceCode = *q.ServiceCode
	sq.ServiceName = *q.ServiceName

	return &sq, nil
}

// returns slice of ServiceQuotaUsage objects
func (r *Runner) GetQuotasUsage() []ServiceQuotaUsage {
	squs := make([]ServiceQuotaUsage, 0, len(*r.quotaApiUsage))

	for k, v := range *r.quotaApiUsage {
		squ := ServiceQuotaUsage{}

		serviceCode, _ := (*r.quotasServiceInfo)[k]
		q, _ := r.GetServiceQuota(serviceCode, k)
		var infoType string
		if _, ok := (*r.quotaUsage)[k]; ok {
			infoType = "api"
		} else if _, ok := (*r.quotaMetricUsage)[k]; ok {
			infoType = "metrics"
		}

		squ.ServiceCode = serviceCode
		squ.ServiceName = r.GetServiceName(serviceCode)
		squ.QuotaName = q.QuotaName
		squ.QuotaCode = k
		squ.Usage = v
		squ.Value = int(q.Value)
		squ.Type = infoType

		squs = append(squs, squ)
	}

	return squs
}

// checks for alarms and returns slice of warnings objects
func (r *Runner) CheckAlarms() []Warning {
	warnings := make([]Warning, 0, 0)

	squs := r.GetQuotasUsage()

	for _, squ := range squs {
		warning := Warning{}
		max := 0
		warning.Limit = squ.Value
		warning.Usage = squ.Usage
		warning.ServiceCode = squ.ServiceCode
		warning.ServiceName = squ.ServiceName
		warning.QuotaName = squ.QuotaName
		warning.QuotaCode = squ.QuotaCode
		found := false
		for k, v := range r.alarms {
			percents := float64(squ.Usage) * 100.0 / float64(squ.Value)
			if percents >= float64(v) && v > max {
				found = true
				max = v
				warning.Name = k
				warning.Threshold = v
			}
		}
		if found {
			warnings = append(warnings, warning)
		}
	}

	return warnings
}

// prints Warning object
func (w Warning) Print() {
	fmt.Println("Limit: ", w.Limit)
	fmt.Println("Usage: ", w.Usage)
	fmt.Println("Name: ", w.Name)
	fmt.Println("Threshold: ", w.Threshold)
	fmt.Println("Service code: ", w.ServiceCode)
	fmt.Println("Service name: ", w.ServiceName)
	fmt.Println("Quota name: ", w.QuotaName)
	fmt.Println("Quota code: ", w.QuotaCode)
}

// prints ServiceQuotaUsage object
func (squ ServiceQuotaUsage) Print() {
	fmt.Println("ServiceCode: ", squ.ServiceCode)
	fmt.Println("ServiceName: ", squ.ServiceName)
	fmt.Println("QuotaName: ", squ.QuotaName)
	fmt.Println("QuotaCode: ", squ.QuotaCode)
	fmt.Println("Usage: ", squ.Usage)
	fmt.Println("Value: ", squ.Value)
	fmt.Println("Type: ", squ.Type)
}

// prints ServiceQuota object
func (sq *ServiceQuota) Print() {
	fmt.Println("Adjustable: ", sq.Adjustable)
	fmt.Println("GlobalQuota: ", sq.GlobalQuota)
	fmt.Println("QuotaName: ", sq.QuotaName)
	fmt.Println("QuotaCode: ", sq.QuotaCode)
	fmt.Println("DefaultValue: ", sq.DefaultValue)
	fmt.Println("Value: ", sq.Value)
	fmt.Println("ServiceCode: ", sq.ServiceCode)
	fmt.Println("ServiceName: ", sq.ServiceName)
}

// returns summary of actions which are needed to work with the packages
func GetIam(allowedServices *map[string]*[]string) iamActions {
	serviceActions := make(iamActions)

	serviceActions["quotas"] = quotas.GetIam()
	serviceActions["cloudwatch"] = quotas.GetIam()

	actionsWithQuotas := make(map[string]map[string]string)
	actionsWithQuotas["autoscaling"] = autoscaling.GetIam()
	actionsWithQuotas["cloudformation"] = cloudformation.GetIam()
	actionsWithQuotas["ec2"] = ec2.GetIam()
	actionsWithQuotas["elasticfilesystem"] = efs.GetIam()
	actionsWithQuotas["elasticbeanstalk"] = elasticbeanstalk.GetIam()
	actionsWithQuotas["elasticloadbalancing"] = elb.GetIam()
	actionsWithQuotas["s3"] = s3.GetIam()
	actionsWithQuotas["vpc"] = vpc.GetIam()

	for service, quotas := range actionsWithQuotas {
		isServiceAllowed := true
		var allowedQuotas *[]string
		var serviceAvailable bool

		if allowedServices != nil {
			if allowedQuotas, serviceAvailable = (*allowedServices)[service]; !serviceAvailable {
				isServiceAllowed = false
			}
		}

		if isServiceAllowed {
			actions := make([]string, 0, 0)
			for quota, action := range quotas {
				isQuotaAllowed := true
				if allowedQuotas != nil && !utils.Find(*allowedQuotas, quota) {
					isQuotaAllowed = false
				}

				if isQuotaAllowed {
					actions = append(actions, action)
				}
			}
			serviceActions[service] = actions
		}
	}

	return serviceActions
}

// prints iamActions object
func (i iamActions) Print() {
	for k, v := range i {
		fmt.Println("Service: ", k)
		fmt.Println("Actions:")
		for _, a := range v {
			fmt.Println(a)
		}
	}
}

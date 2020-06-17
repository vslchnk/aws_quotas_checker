package quotas

import (
	"fmt"
	"strings"

	"github.com/vslchnk/aws_quotas_checker/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicequotas"
)

type serviceQuota struct {
	servicequotas.ServiceQuota
	ValueApplied *float64
}

type serviceInfo struct {
	serviceName   string
	serviceQuotas map[string]*serviceQuota
}

type Quotas struct {
	region          string
	allowedServices *map[string]*[]string
	client          *servicequotas.ServiceQuotas
	servicesMap     map[string]*serviceInfo
}

// creates Quotas agent
func NewQuota(region string, allowedServices *map[string]*[]string) (*Quotas, error) {
	q := Quotas{}
	q.region = region
	q.allowedServices = allowedServices

	ses, err := session.NewSession(&aws.Config{
		Region: aws.String(q.region)},
	)

	if err != nil {
		return nil, fmt.Errorf("Error while creating session: %v", err)
	}

	q.client = servicequotas.New(ses)
	q.servicesMap, err = getServicesMap(q.client, q.allowedServices)

	if err != nil {
		return nil, fmt.Errorf("Error while getting services information: %v", err)
	}

	return &q, nil
}

// retunrs map with the key as a service code and value as a serviceInfo object
func getServicesMap(client *servicequotas.ServiceQuotas, allowedServices *map[string]*[]string) (map[string]*serviceInfo, error) {
	services := make(map[string]*serviceInfo)

	err := client.ListServicesPages(nil, func(page *servicequotas.ListServicesOutput, lastPage bool) bool {
		for _, value := range page.Services {
			isServiceAllowed := true
			var allowedQuotas *[]string

			if allowedServices != nil {
				if !utils.CheckKeyMap(*allowedServices, *value.ServiceCode) {
					isServiceAllowed = false
				} else {
					allowedQuotas, _ = (*allowedServices)[*value.ServiceCode]
				}
			}

			if isServiceAllowed {
				sq, err := getQuotasMap(client, *value.ServiceCode, allowedQuotas)

				if err != nil {
					utils.ExitErrorf("Unable to get quotas, %v", err)
				}

				si := &serviceInfo{}
				si.serviceName = *value.ServiceName
				si.serviceQuotas = sq

				services[*value.ServiceCode] = si
			}
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("Error while getting list of services: %v", err)
	}

	return services, nil
}

// returns map with information about quotas where key is quota code and value is serviceQuota object
func getQuotasMap(client *servicequotas.ServiceQuotas, service string, allowedQuotas *[]string) (map[string]*serviceQuota, error) {
	quotas := make(map[string]*serviceQuota)

	params := &servicequotas.ListAWSDefaultServiceQuotasInput{
		ServiceCode: aws.String(service),
	}

	err := client.ListAWSDefaultServiceQuotasPages(params, func(page *servicequotas.ListAWSDefaultServiceQuotasOutput, lastPage bool) bool {
		for _, value := range page.Quotas {
			isQuotaAllowed := true
			if allowedQuotas != nil && !utils.Find(*allowedQuotas, *value.QuotaCode) {
				isQuotaAllowed = false
			}
			if isQuotaAllowed {
				quota := &serviceQuota{}
				quota.ServiceQuota = *value
				var err error
				quota.ValueApplied, err = getAppliedQuotaValue(client, service, *value.QuotaCode, value.Value)

				if err != nil {
					utils.ExitErrorf("Unable to get applied quota value, %v", err)
				}

				quotas[*value.QuotaCode] = quota
			}
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("Error while getting list of quotas: %v", err)
	}

	return quotas, nil
}

// returns applied quota value, could be the default one
func getAppliedQuotaValue(client *servicequotas.ServiceQuotas, service string, quotaCode string, defaultValue *float64) (*float64, error) {
	params := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String(service),
		QuotaCode:   aws.String(quotaCode),
	}

	value, err := client.GetServiceQuota(params)

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchResourceException") {
			return defaultValue, nil
		} else {
			return nil, err
		}
	}

	return value.Quota.Value, nil
}

// returns slice of codes of available service
func (q *Quotas) ListServicesCodes() *[]string {
	serviceCodes := make([]string, 0, len(q.servicesMap))

	for k, _ := range q.servicesMap {
		serviceCodes = append(serviceCodes, k)
	}

	return &serviceCodes
}

// returns string slice of quotas codes for the service's code
func (q *Quotas) ListQuotasCodes(serviceCode string) (*[]string, error) {
	if service, ok := q.servicesMap[serviceCode]; ok {
		quotasCodes := make([]string, 0, len(service.serviceQuotas))

		for k, _ := range service.serviceQuotas {
			quotasCodes = append(quotasCodes, k)
		}

		return &quotasCodes, nil
	}

	return nil, fmt.Errorf("No service with such code: %v", serviceCode)
}

// returns serviceInfo objects by the code of the service
func (q *Quotas) GetService(serviceCode string) (*serviceInfo, error) {
	if service, ok := q.servicesMap[serviceCode]; ok {
		return service, nil
	}

	return nil, fmt.Errorf("No service with such code: %v", serviceCode)
}

// returns name of the service for serviceInfo object
func (s *serviceInfo) GetServiceName() string {
	return s.serviceName
}

// returns serviceQuota object by quota code
func (s *serviceInfo) GetServiceQuota(quotaCode string) (*serviceQuota, error) {
	if quota, ok := s.serviceQuotas[quotaCode]; ok {
		return quota, nil
	}

	return nil, fmt.Errorf("No quota with such code: %v", quotaCode)
}

// returns actions for IAM policy which allow to work with this package
func GetIam() []string {
	actions := []string{
		"servicequotas:GetServiceQuota",
		"servicequotas:ListAWSDefaultServiceQuotas",
		"servicequotas:ListServices",
	}

	return actions
}

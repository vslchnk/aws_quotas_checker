package main

import (
	"checker/runner"
	"fmt"
)

func main() {
	actions := runner.GetIam(nil)
	actions.Print()

	allowedServices := make(map[string]*[]string)

	quotasEC2 := make([]string, 0, 0)
	quotasEC2 = append(quotasEC2, "L-74FC7D96")
	quotasEC2 = append(quotasEC2, "L-4FB7FF5D")
	quotasEC2 = append(quotasEC2, "L-1216C47A")

	quotasAs := make([]string, 0, 0)
	quotasAs = append(quotasAs, "L-CDE20ADC")

	allowedServices["ec2"] = &quotasEC2
	allowedServices["cloudformation"] = nil
	allowedServices["access-analyzer"] = nil
	allowedServices["elasticbeanstalk"] = nil
	allowedServices["autoscaling"] = &quotasAs
	allowedServices["s3"] = nil
	allowedServices["vpc"] = nil
	allowedServices["elasticloadbalancing"] = nil
	allowedServices["elasticfilesystem"] = nil

	actions = runner.GetIam(&allowedServices)
	actions.Print()

	r, err := runner.NewRunner("us-east-2", &allowedServices)
	fmt.Println(err)

	// list all
	for _, s := range r.ListSupportedServicesFromQuotas() {
		fmt.Println("Service:", s, r.GetServiceName(s))
		quotas, _ := r.GetSupportedQuotasForService(s)
		for _, q := range quotas {
			quota, _ := r.GetServiceQuota(s, q)
			quota.Print()
		}
	}

	// get usage
	usage := r.GetQuotasUsage()
	for _, u := range usage {
		u.Print()
	}

	// check alarms
	r.AddAlarm("low", 10)
	r.AddAlarm("very low", 5)
	warnings := r.CheckAlarms()
	for _, w := range warnings {
		w.Print()
	}
}

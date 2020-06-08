# AWS Quota Checker
## Purpose
AWS Quota Checker allows you to list services and quotas from AWS Service Quotas, find their usage, create alarms with thresholds and check them.

## State
Package is under development and is in early stage. Now you usage of Quotas API is fully available, however usage can be found not for all quotas and services. List of supported Quotas can be shown from the code.

## Usage
### First run:
To run AWS Quota Checker you need AWS user or IAM role with proper policies. To list all needed actions for all services run:
```golang
actions := runner.GetIam(nil)
actions.Print()
```

When you creare a runner agent with NewRunner() function it takes some time to get all information about usage from API endpoints. To view all available services and quotas with information about them from Quotas API run passing the region you need to runner agent:
```golang
r, err := runner.NewRunner("eu-west-1", nil)
fmt.Println(err)

for _, s := range r.ListSupportedServicesFromQuotas() {
	fmt.Println("Service:", s, r.GetServiceName(s))
	quotas, _ := r.GetSupportedQuotasForService(s)
	for _, q := range quotas {
		quota, _ := r.GetServiceQuota(s, q)
		quota.Print()
	}
}
```
### Usage examples:
After you have a list of services and codes you can specify allowed services and quotas by their codes:
```golang
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

r, err := runner.NewRunner("us-east-2", &allowedServices)
```
Allowed services should be a map containig service codes as keys and string slices of quotas code as values. To allow all services you pass nil to runner agent. Same goes for quotas. You specify nil as a value if you want all quotas to be allowed.
Also if you can get actions for AWS policy for allowed services by passing created map:
```golang
actions = runner.GetIam(&allowedServices)
actions.Print()
```
To get usage of quotas run:
```golang
usage := r.GetQuotasUsage()
for _, u := range usage {
	u.Print()
}
```
It returns a list of allowed and supported quotas.
Finally to set alarms with thresholds and to check them run:
```golang
r.AddAlarm("low", 10)
r.AddAlarm("very low", 5)

warnings := r.CheckAlarms()
for _, w := range warnings {
	w.Print()
}
```
Here we create two alarms. The first argument for AddAlarm function is alarm name and the seconde one is percentes which describe maximum percentage usage for the alarm.
Warning objects describe alarm and quota.

Example of usage can be found in example folder.

## License
Apache 2.0 is used.

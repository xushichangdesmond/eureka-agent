package agent

type Registration struct {
	Instance Instance `json:"instance"`
}

type Instance struct {
	InstanceId     string         `json:"instanceId"`
	HostName       string         `json:"hostName"`
	App            string         `json:"app"`
	IpAddr         string         `json:"ipAddr"`
	VipAddr        string         `json:"vipAddress"`
	SecureVipAddr  string         `json:"secureVipAddress"`
	Status         string         `json:"status"`
	Port           Port           `json:"port"`
	SecurePort     Port           `json:"securePort"`
	HealthCheckUrl string         `json:"healthCheckUrl"`
	StatusPageUrl  string         `json:"statusPageUrl"`
	HomePageUrl    string         `json:"homePageUrl"`
	DataCenterInfo DataCenterInfo `json:"dataCenterInfo"`
}

type Port struct {
	Port    string `json:"$"`
	Enabled string `json:"@enabled"`
}

type DataCenterInfo struct {
	Class string `json:"@class"`
	Name  string `json:"name"`
}

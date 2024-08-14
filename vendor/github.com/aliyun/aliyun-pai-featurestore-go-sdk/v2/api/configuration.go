package api

import "fmt"

type Configuration struct {
	regionId        string
	AccessKeyId     string
	AccessKeySecret string
	projectName     string
	UserAgent       string
	domain          string
}

func NewConfiguration(regionId, accessKeyId, accessKeySecret, projectName string) *Configuration {
	cfg := &Configuration{
		UserAgent:       "PAI-FeatureStore/1.0.0/go",
		regionId:        regionId,
		projectName:     projectName,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	return cfg
}

func (c *Configuration) SetDomain(domain string) {
	c.domain = domain
}

func (c *Configuration) GetDomain() string {
	if c.domain == "" {
		c.domain = fmt.Sprintf("paifeaturestore-vpc.%s.aliyuncs.com", c.regionId)
	}

	return c.domain
}

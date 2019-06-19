package model

// K8sInfoData holds all scraped data
type K8sInfoData struct {
	Deployments []DeploymentData
}

// DeploymentData contains all scraped data for one deployment
type DeploymentData struct {
	Context        string
	DeploymentName string
	ContainerName  string
	Namespace      string
	Image          string
	Labels         map[string]string
}

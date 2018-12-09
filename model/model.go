package model

// K8sInfoData holds all scarped data
type K8sInfoData struct {
	Deployments []DeploymentData
}

// DeploymentData contains all scarped data for one deployment
type DeploymentData struct {
	Context   string
	Name      string
	Namespace string
	Image     string
}

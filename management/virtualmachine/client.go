// Package virtualmachine provides a client for Virtual Machines.
package virtualmachine

import (
	"encoding/xml"
	"fmt"

	"github.com/forqlift/azure-sdk-for-go/management"
)

const (
	azureDeploymentListURL   = "services/hostedservices/%s/deployments"
	azureDeploymentURL       = "services/hostedservices/%s/deployments/%s"
	deleteAzureDeploymentURL = "services/hostedservices/%s/deployments/%s?comp=media"
	azureRoleURL             = "services/hostedservices/%s/deployments/%s/roles/%s"
	azureOperationsURL       = "services/hostedservices/%s/deployments/%s/roleinstances/%s/Operations"
	azureRoleSizeListURL     = "rolesizes"

	errParamNotSpecified = "Parameter %s is not specified."
)

//NewClient is used to instantiate a new VirtualMachineClient from an Azure client
func NewClient(client management.Client) VirtualMachineClient {
	return VirtualMachineClient{client: client}
}

// CreateDeploymentOptions can be used to create a customized deployement request
type CreateDeploymentOptions struct {
	DNSServers         []DNSServer
	LoadBalancers      []LoadBalancer
	ReservedIPName     string
	VirtualNetworkName string
}

// CreateDeployment creates a deployment and then creates a virtual machine
// in the deployment based on the specified configuration.
//
// https://msdn.microsoft.com/en-us/library/azure/jj157194.aspx
func (vm VirtualMachineClient) CreateDeployment(
	role Role,
	cloudServiceName string,
	options CreateDeploymentOptions) (management.OperationID, error) {

	req := DeploymentRequest{
		Name:               role.RoleName,
		DeploymentSlot:     "Production",
		Label:              role.RoleName,
		RoleList:           []Role{role},
		DNSServers:         options.DNSServers,
		LoadBalancers:      options.LoadBalancers,
		ReservedIPName:     options.ReservedIPName,
		VirtualNetworkName: options.VirtualNetworkName,
	}

	data, err := xml.Marshal(req)
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf(azureDeploymentListURL, cloudServiceName)
	return vm.client.SendAzurePostRequest(requestURL, data)
}

func (vm VirtualMachineClient) GetDeployment(cloudServiceName, deploymentName string) (DeploymentResponse, error) {
	var deployment DeploymentResponse
	if cloudServiceName == "" {
		return deployment, fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return deployment, fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	requestURL := fmt.Sprintf(azureDeploymentURL, cloudServiceName, deploymentName)
	response, azureErr := vm.client.SendAzureGetRequest(requestURL)
	if azureErr != nil {
		return deployment, azureErr
	}

	err := xml.Unmarshal(response, &deployment)
	return deployment, err
}

func (vm VirtualMachineClient) DeleteDeployment(cloudServiceName, deploymentName string) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}

	requestURL := fmt.Sprintf(deleteAzureDeploymentURL, cloudServiceName, deploymentName)
	return vm.client.SendAzureDeleteRequest(requestURL)
}

func (vm VirtualMachineClient) GetRole(cloudServiceName, deploymentName, roleName string) (*Role, error) {
	if cloudServiceName == "" {
		return nil, fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return nil, fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return nil, fmt.Errorf(errParamNotSpecified, "roleName")
	}

	role := new(Role)

	requestURL := fmt.Sprintf(azureRoleURL, cloudServiceName, deploymentName, roleName)
	response, azureErr := vm.client.SendAzureGetRequest(requestURL)
	if azureErr != nil {
		return nil, azureErr
	}

	err := xml.Unmarshal(response, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// UpdateRole updates the configuration of the specified virtual machine
// See https://msdn.microsoft.com/en-us/library/azure/jj157187.aspx
func (vm VirtualMachineClient) UpdateRole(cloudServiceName, deploymentName, roleName string, role Role) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	data, err := xml.Marshal(PersistentVMRole{Role: role})
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf(azureRoleURL, cloudServiceName, deploymentName, roleName)
	return vm.client.SendAzurePutRequest(requestURL, "text/xml", data)
}

func (vm VirtualMachineClient) StartRole(cloudServiceName, deploymentName, roleName string) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	startRoleOperationBytes, err := xml.Marshal(StartRoleOperation{
		OperationType: "StartRoleOperation",
	})
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudServiceName, deploymentName, roleName)
	return vm.client.SendAzurePostRequest(requestURL, startRoleOperationBytes)
}

func (vm VirtualMachineClient) ShutdownRole(cloudServiceName, deploymentName, roleName string) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	shutdownRoleOperationBytes, err := xml.Marshal(ShutdownRoleOperation{
		OperationType: "ShutdownRoleOperation",
	})
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudServiceName, deploymentName, roleName)
	return vm.client.SendAzurePostRequest(requestURL, shutdownRoleOperationBytes)
}

func (vm VirtualMachineClient) RestartRole(cloudServiceName, deploymentName, roleName string) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	restartRoleOperationBytes, err := xml.Marshal(RestartRoleOperation{
		OperationType: "RestartRoleOperation",
	})
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudServiceName, deploymentName, roleName)
	return vm.client.SendAzurePostRequest(requestURL, restartRoleOperationBytes)
}

func (vm VirtualMachineClient) DeleteRole(cloudServiceName, deploymentName, roleName string) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	requestURL := fmt.Sprintf(azureRoleURL, cloudServiceName, deploymentName, roleName)
	return vm.client.SendAzureDeleteRequest(requestURL)
}

func (vm VirtualMachineClient) GetRoleSizeList() (RoleSizeList, error) {
	roleSizeList := RoleSizeList{}

	response, err := vm.client.SendAzureGetRequest(azureRoleSizeListURL)
	if err != nil {
		return roleSizeList, err
	}

	err = xml.Unmarshal(response, &roleSizeList)
	return roleSizeList, err
}

// CaptureRole captures a VM role. If reprovisioningConfigurationSet is non-nil,
// the VM role is redeployed after capturing the image, otherwise, the original
// VM role is deleted.
//
// NOTE: an image resulting from this operation shows up in
// osimage.GetImageList() as images with Category "User".
func (vm VirtualMachineClient) CaptureRole(cloudServiceName, deploymentName, roleName, imageName, imageLabel string,
	reprovisioningConfigurationSet *ConfigurationSet) (management.OperationID, error) {
	if cloudServiceName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "cloudServiceName")
	}
	if deploymentName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "deploymentName")
	}
	if roleName == "" {
		return "", fmt.Errorf(errParamNotSpecified, "roleName")
	}

	if reprovisioningConfigurationSet != nil &&
		!(reprovisioningConfigurationSet.ConfigurationSetType == ConfigurationSetTypeLinuxProvisioning ||
			reprovisioningConfigurationSet.ConfigurationSetType == ConfigurationSetTypeWindowsProvisioning) {
		return "", fmt.Errorf("ConfigurationSet type can only be WindowsProvisioningConfiguration or LinuxProvisioningConfiguration")
	}

	operation := CaptureRoleOperation{
		OperationType:             "CaptureRoleOperation",
		PostCaptureAction:         PostCaptureActionReprovision,
		ProvisioningConfiguration: reprovisioningConfigurationSet,
		TargetImageLabel:          imageLabel,
		TargetImageName:           imageName,
	}
	if reprovisioningConfigurationSet == nil {
		operation.PostCaptureAction = PostCaptureActionDelete
	}

	data, err := xml.Marshal(operation)
	if err != nil {
		return "", err
	}

	return vm.client.SendAzurePostRequest(fmt.Sprintf(azureOperationsURL, cloudServiceName, deploymentName, roleName), data)
}

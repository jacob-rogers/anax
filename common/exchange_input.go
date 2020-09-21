package common

import (
	"errors"
	"fmt"
	"github.com/open-horizon/anax/containermessage"
	"github.com/open-horizon/anax/exchange"
	"github.com/open-horizon/anax/i18n"
	"github.com/open-horizon/anax/policy"
	"github.com/open-horizon/anax/semanticversion"
	"golang.org/x/text/message"
)

type AbstractServiceFile interface {
	GetOrg() string
	GetURL() string
	GetVersion() string
	GetArch() string
	GetServiceType() string // device, cluster or both
	GetRequiredServices() []exchange.ServiceDependency
	GetUserInputs() []exchange.UserInput
	NeedsUserInput() bool
	GetDeployment() interface{}
	GetClusterDeployment() interface{}
}

type AbstractPatternFile interface {
	GetOrg() string
	GetServices() []exchange.ServiceReference
	GetUserInputs() []policy.UserInput
}

// This is used when reading json file the user gives us as an input to create the pattern
type PatternFile struct {
	Name               string                       `json:"name,omitempty"`
	Org                string                       `json:"org,omitempty"` // optional
	Label              string                       `json:"label"`
	Description        string                       `json:"description,omitempty"`
	Public             bool                         `json:"public"`
	Services           []ServiceReferenceFile       `json:"services"`
	AgreementProtocols []exchange.AgreementProtocol `json:"agreementProtocols,omitempty"`
	UserInput          []policy.UserInput           `json:"userInput,omitempty"`
}

func (p *PatternFile) GetOrg() string {
	return p.Org
}

// convert the []ServiceReferenceFile to []exchange.ServiceReference
// Not converting te depployment strings for now.
func (p *PatternFile) GetServices() []exchange.ServiceReference {
	service_refs := []exchange.ServiceReference{}
	if p.Services != nil {
		for _, svc := range p.Services {
			sref := exchange.ServiceReference{}
			sref.ServiceURL = svc.ServiceURL
			sref.ServiceOrg = svc.ServiceOrg
			sref.ServiceArch = svc.ServiceArch
			sref.AgreementLess = svc.AgreementLess
			if svc.DataVerify != nil {
				sref.DataVerify = *svc.DataVerify
			}
			if svc.NodeH != nil {
				sref.NodeH = *svc.NodeH
			}

			versions := []exchange.WorkloadChoice{}
			if svc.ServiceVersions != nil {
				for _, v := range svc.ServiceVersions {
					c := exchange.WorkloadChoice{Version: v.Version}
					if v.Priority != nil {
						c.Priority = *v.Priority
					}
					if v.Upgrade != nil {
						c.Upgrade = *v.Upgrade
					}
					versions = append(versions, c)
				}
			}
			sref.ServiceVersions = versions

			service_refs = append(service_refs, sref)
		}
	}

	return service_refs
}

func (p *PatternFile) GetUserInputs() []policy.UserInput {
	return p.UserInput
}

type ServiceReferenceFile struct {
	ServiceURL      string                     `json:"serviceUrl"`                 // refers to a service definition in the exchange
	ServiceOrg      string                     `json:"serviceOrgid"`               // the org holding the service definition
	ServiceArch     string                     `json:"serviceArch"`                // the hardware architecture of the service definition
	AgreementLess   bool                       `json:"agreementLess,omitempty"`    // a special case where this service will also be required by others
	ServiceVersions []ServiceChoiceFile        `json:"serviceVersions"`            // a list of service version for rollback
	DataVerify      *exchange.DataVerification `json:"dataVerification,omitempty"` // policy for verifying that the node is sending data
	NodeH           *exchange.NodeHealth       `json:"nodeHealth,omitempty"`       // this needs to be a ptr so it will be omitted if not specified, so exchange will default it
}

type ServiceChoiceFile struct {
	Version                      string                     `json:"version"`            // the version of the service
	Priority                     *exchange.WorkloadPriority `json:"priority,omitempty"` // the highest priority service is tried first for an agreement, if it fails, the next priority is tried. Priority 1 is the highest, priority 2 is next, etc.
	Upgrade                      *exchange.UpgradePolicy    `json:"upgradePolicy,omitempty"`
	DeploymentOverrides          interface{}                `json:"deployment_overrides,omitempty"`           // env var overrides for the service
	DeploymentOverridesSignature string                     `json:"deployment_overrides_signature,omitempty"` // signature of env var overrides
}

// This is used when reading json file the user gives us as input to create the service
type ServiceFile struct {
	Org                        string                       `json:"org"` // optional
	Label                      string                       `json:"label"`
	Description                string                       `json:"description"`
	Public                     bool                         `json:"public"`
	Documentation              string                       `json:"documentation"`
	URL                        string                       `json:"url"`
	Version                    string                       `json:"version"`
	Arch                       string                       `json:"arch"`
	Sharable                   string                       `json:"sharable"`
	MatchHardware              map[string]interface{}       `json:"matchHardware,omitempty"`
	RequiredServices           []exchange.ServiceDependency `json:"requiredServices"`
	UserInputs                 []exchange.UserInput         `json:"userInput"`
	Deployment                 interface{}                  `json:"deployment,omitempty"` // interface{} because pre-signed services can be stringified json
	DeploymentSignature        string                       `json:"deploymentSignature,omitempty"`
	ClusterDeployment          interface{}                  `json:"clusterDeployment,omitempty"`
	ClusterDeploymentSignature string                       `json:"clusterDeploymentSignature,omitempty"`
}

func (sf *ServiceFile) GetOrg() string {
	return sf.Org
}

func (sf *ServiceFile) GetURL() string {
	return sf.URL
}

func (sf *ServiceFile) GetVersion() string {
	return sf.Version
}

func (sf *ServiceFile) GetArch() string {
	return sf.Arch
}

func (sf *ServiceFile) GetRequiredServices() []exchange.ServiceDependency {
	return sf.RequiredServices
}

func (sf *ServiceFile) GetUserInputs() []exchange.UserInput {
	return sf.UserInputs
}

func (s *ServiceFile) NeedsUserInput() bool {
	if s.UserInputs == nil || len(s.UserInputs) == 0 {
		return false
	}

	for _, ui := range s.UserInputs {
		if ui.Name != "" && ui.DefaultValue == "" {
			return true
		}
	}
	return false
}

func (sf *ServiceFile) GetDeployment() interface{} {
	return sf.Deployment
}

func (sf *ServiceFile) GetClusterDeployment() interface{} {
	return sf.ClusterDeployment
}

// Get the service type
// Check for nil, "" and {} for deployment and cluster deployment.
func (s *ServiceFile) GetServiceType() string {
	sType := exchange.SERVICE_TYPE_DEVICE
	if s.ClusterDeployment != nil && s.ClusterDeployment != "" {
		if s.Deployment == nil || s.Deployment == "" {
			sType = exchange.SERVICE_TYPE_CLUSTER
		} else {
			sType = exchange.SERVICE_TYPE_BOTH
		}
	}
	return sType
}

// Returns true if the service definition userinputs define the variable.
func (sf *ServiceFile) DefinesVariable(name string) string {
	for _, ui := range sf.UserInputs {
		if ui.Name == name && ui.Type != "" {
			return ui.Type
		}
	}
	return ""
}

// Returns true if the service definition has required services.
func (sf *ServiceFile) HasDependencies() bool {
	if len(sf.RequiredServices) == 0 {
		return false
	}
	return true
}

// Return true if the service definition is a dependency in the input list of service references.
func (sf *ServiceFile) IsDependent(deps []exchange.ServiceDependency) bool {
	for _, dep := range deps {
		if sf.URL == dep.URL && sf.Org == dep.Org {
			return true
		}
	}
	return false
}

// Convert the Deployment Configuration to a full Deployment Description.
func (sf *ServiceFile) ConvertToDeploymentDescription(agreementService bool) (*DeploymentConfig, *containermessage.DeploymentDescription, error) {
	depConfig, err := ConvertToDeploymentConfig(sf.Deployment)
	if err != nil {
		return nil, nil, err
	}
	infra := !agreementService
	return depConfig, &containermessage.DeploymentDescription{
		Services: depConfig.Services,
		ServicePattern: containermessage.Pattern{
			Shared: map[string][]string{},
		},
		Infrastructure: infra,
		Overrides:      map[string]*containermessage.Service{},
	}, nil
}

// Verify that non default user inputs are set in the input map.
func (sf *ServiceFile) RequiredVariablesAreSet(setVarNames []string) error {
	for _, ui := range sf.UserInputs {
		if ui.DefaultValue == "" && ui.Name != "" {
			found := false
			for _, v := range setVarNames {
				if v == ui.Name {
					found = true
				}
			}
			if !found {
				return errors.New(i18n.GetMessagePrinter().Sprintf("user input %v has no default value and is not set", ui.Name))
			}
		}
	}
	return nil
}

func (sf *ServiceFile) SupportVersionRange() {
	for ix, sdep := range sf.RequiredServices {
		if sdep.VersionRange == "" {
			sf.RequiredServices[ix].VersionRange = sf.RequiredServices[ix].Version
		}
	}
}

// Validate a service definition.
// Varifies the existance of the dependent services.
// Verifies consistence for the dependent service types
// Make sure userinput and requiredServices are not supported for cluster services.
func ValidateService(serviceDefResolverHandler exchange.ServiceDefResolverHandler, svcFile AbstractServiceFile, msgPrinter *message.Printer) error {
	// get default message printer if nil
	if msgPrinter == nil {
		msgPrinter = i18n.GetMessagePrinter()
	}

	// cluster type, userinput and requiredServices are not allowed
	topSvcType := svcFile.GetServiceType()
	requiredServices := svcFile.GetRequiredServices()
	if topSvcType == exchange.SERVICE_TYPE_CLUSTER {
		if requiredServices != nil && len(requiredServices) != 0 {
			return fmt.Errorf(msgPrinter.Sprintf("'requiredServices' is not supported for cluster type service."))
		}
	} else {
		// if it the service type is 'device' or 'both', make sure all the dependent services are 'device' or 'both' types
		if requiredServices != nil {
			for _, reqSvc := range requiredServices {

				// get the service definition for the required service and all of it dependents
				ver := reqSvc.GetVersionRange()
				vExp, err := semanticversion.Version_Expression_Factory(ver)
				if err != nil {
					return fmt.Errorf(msgPrinter.Sprintf("Failed to convert version %v for service %v to version range expression.", ver, reqSvc))
				}
				svc_map, sDef, sId, err := serviceDefResolverHandler(reqSvc.URL, reqSvc.Org, vExp.Get_expression(), reqSvc.Arch)
				if err != nil {
					return fmt.Errorf(msgPrinter.Sprintf("Error retrieving service from the Exchange for %v. %v", reqSvc, err))
				}

				// check the node type for the required service
				sType := sDef.GetServiceType()
				if sType == exchange.SERVICE_TYPE_CLUSTER {
					return fmt.Errorf(msgPrinter.Sprintf("The required service %v has the wrong service type: %v.", sId, sType))
				}

				// check the node type of the dependent services of the required service
				for id, s := range svc_map {
					sType1 := s.GetServiceType()
					if sType == exchange.SERVICE_TYPE_CLUSTER {
						return fmt.Errorf(msgPrinter.Sprintf("The dependent service %v for the required service %v has the wrong service type: %v.", id, sId, sType1))
					}
				}
			}
		}
	}

	return nil
}

// check if the deployment is empty. The following cases are considered empty in JSON:
// "deployment": {}
// "deployment": null
// "deployment": ""
func DeploymentIsEmpty(deployment interface{}) bool {
	switch deployment.(type) {
	case nil:
		return true
	case map[string]interface{}:
		if len(deployment.(map[string]interface{})) == 0 {
			return true
		}
	case string:
		if deployment.(string) == "" {
			return true
		}
	}

	return false
}

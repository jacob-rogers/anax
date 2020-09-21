package exchange

import (
	"github.com/open-horizon/anax/externalpolicy"
	"testing"
	"time"
)

type TestStruct struct {
	testString string
}

func TestUpdateCache(t *testing.T) {
	testResource := TestStruct{testString: "Unit testing"}
	UpdateCache("test/resource", "TEST_TYPE", testResource)

	foundResource := GetResourceFromCache("test/resource", "TEST_TYPE", 0)
	if typedFoundResource, ok := foundResource.(TestStruct); !ok {
		t.Errorf("Error: cached resource returned not of expected type.")
	} else if typedFoundResource.testString != "Unit testing" {
		t.Errorf("Error: cached resource found is different than what was put there.")
	}

	testResource = TestStruct{testString: "Another test"}
	UpdateCache("test/resource", "TEST_TYPE", testResource)

	foundResource = GetResourceFromCache("test/resource", "TEST_TYPE", 0)
	if typedFoundResource, ok := foundResource.(TestStruct); !ok {
		t.Errorf("Error: cached resource returned not of expected type.")
	} else if typedFoundResource.testString != "Unit testing" {
		t.Errorf("Error: cached resource found is different than what was put there.")
	}

	testResource = TestStruct{testString: "Another test"}
	UpdateCache("test/resource", "TEST_TYPE", testResource)

	time.Sleep(time.Second * 2)
	foundResource = GetResourceFromCache("test/resource", "TEST_TYPE", 1)
	if foundResource != nil {
		t.Errorf("Error: expired resource returned from cache.")
	}
}

func TestGetNodeFromCache(t *testing.T) {
	nodeDef := Device{Name: "test-node-1", Arch: "amd64", NodeType: "cluster", Pattern: "A Pattern"}
	UpdateCache(NodeCacheMapKey("userdev", "test-node-1"), NODE_DEF_TYPE_CACHE, nodeDef)

	cachedNodeDef := GetNodeFromCache("userdev", "test-node-1")
	if cachedNodeDef.Name != "test-node-1" || cachedNodeDef.Arch != "amd64" || cachedNodeDef.NodeType != "cluster" || cachedNodeDef.Pattern != "A Pattern" {
		t.Errorf("Error: cache returned different resource than what was stored.")
	}
}

func TestGetServiceFromCache(t *testing.T) {
	svcDefs := map[string]ServiceDefinition{}
	svcDefs["0.0.0"] = ServiceDefinition{Owner: "joe@somecomp.com", URL: "a-new-service", Arch: "amd64", Version: "0.0.0", Deployment: "abcdefg12345"}
	svcDefs["0.0.1"] = ServiceDefinition{Owner: "juan@somecomp.com", URL: "a-new-service", Arch: "amd64", Version: "0.0.0", Deployment: "gfedcba54321"}

	UpdateCache(ServiceCacheMapKey("e2edev@somecomp.com", "a-new-service", "amd64"), SVC_DEF_TYPE_CACHE, svcDefs)

	cachedSvcDefs := GetServiceFromCache("e2edev@somecomp.com", "a-new-service", "amd64")

	if cachedSvcDefs["0.0.0"].Deployment != "abcdefg12345" || cachedSvcDefs["0.0.1"].Deployment != "gfedcba54321" {
		t.Errorf("Error: unexpected value found in cached service.")
	}
}

func TestGetServicePolicyFromCache(t *testing.T) {
	svcPol := externalpolicy.ExternalPolicy{Properties: externalpolicy.PropertyList{*externalpolicy.Property_Factory("prop1", 5)}, Constraints: externalpolicy.ConstraintExpression{"openhorizon.cpu > 2"}}
	exchPol := ExchangePolicy{svcPol, "12:00:00"}

	UpdateCache(ServicePolicyCacheMapKey("e2edev@somecomp.com", "test-service", "amd64", "2.9.13"), SVC_POL_TYPE_CACHE, exchPol)

	cachedSvcPol := GetServicePolicyFromCache("e2edev@somecomp.com", "test-service", "amd64", "2.9.13")

	if cachedSvcPol.Properties[0].Name != "prop1" || cachedSvcPol.Properties[0].Value != 5 || cachedSvcPol.Constraints[0] != "openhorizon.cpu > 2" {
		t.Errorf("Error: unexpected value found in cached service policy.")
	}
}

func TestDeleteCacheResourceFromChange(t *testing.T) {
	nodeDef1 := Device{Name: "test-node-1", Arch: "amd64", NodeType: "cluster", Pattern: "A Pattern"}
	nodeDef2 := Device{Name: "test-node-2", Arch: "amd64", NodeType: "cluster", Pattern: "Different Pattern"}
	nodeDef3 := Device{Name: "test-node-2", Arch: "amd64", NodeType: "edge", Pattern: "EdgePattern"}

	svcDefs1 := map[string]ServiceDefinition{}
	svcDefs1["0.0.0"] = ServiceDefinition{Owner: "joe@somecomp.com", URL: "a-new-service", Arch: "amd64", Version: "0.0.0", Deployment: "abcdefg12345"}
	svcDefs1["0.0.1"] = ServiceDefinition{Owner: "juan@somecomp.com", URL: "a-new-service", Arch: "amd64", Version: "0.0.0", Deployment: "gfedcba54321"}

	svcDefs2 := map[string]ServiceDefinition{}
	svcDefs2["0.0.0"] = ServiceDefinition{Owner: "charlie", URL: "another-service", Arch: "amd64", Version: "0.0.0", Deployment: "abcdefg12345"}
	svcDefs2["0.0.1"] = ServiceDefinition{Owner: "lucy", URL: "another-service", Arch: "amd64", Version: "0.0.0", Deployment: "gfedcba54321"}

	svcPol1 := ExchangePolicy{ExternalPolicy: externalpolicy.ExternalPolicy{Properties: externalpolicy.PropertyList{*externalpolicy.Property_Factory("prop1", 5)}, Constraints: externalpolicy.ConstraintExpression{"openhorizon.cpu > 2"}}}
	svcPol2 := ExchangePolicy{ExternalPolicy: externalpolicy.ExternalPolicy{Properties: externalpolicy.PropertyList{*externalpolicy.Property_Factory("color", "green")}, Constraints: externalpolicy.ConstraintExpression{"serviceVersion in [0.0.0,1.3.6)"}}}

	nodePol1 := ExchangePolicy{ExternalPolicy: externalpolicy.ExternalPolicy{Properties: externalpolicy.PropertyList{*externalpolicy.Property_Factory("prop1", 5)}, Constraints: externalpolicy.ConstraintExpression{"openhorizon.cpu > 2"}}}
	nodePol2 := ExchangePolicy{ExternalPolicy: externalpolicy.ExternalPolicy{Properties: externalpolicy.PropertyList{*externalpolicy.Property_Factory("color", "green")}, Constraints: externalpolicy.ConstraintExpression{"serviceVersion in [0.0.0,1.3.6)"}}}

	UpdateCache(NodeCacheMapKey("e2edev@somecomp.com", "test-node-1"), NODE_DEF_TYPE_CACHE, nodeDef1)
	UpdateCache(NodeCacheMapKey("e2edev@somecomp.com", "test-node-2"), NODE_DEF_TYPE_CACHE, nodeDef2)
	UpdateCache(NodeCacheMapKey("userdev", "test-node-3"), NODE_DEF_TYPE_CACHE, nodeDef3)
	UpdateCache(ServiceCacheMapKey("e2edev@somecomp.com", "a-new-service", "amd64"), SVC_DEF_TYPE_CACHE, svcDefs1)
	UpdateCache(ServiceCacheMapKey("userdev", "another-service", "amd64"), SVC_DEF_TYPE_CACHE, svcDefs2)
	UpdateCache(NodeCacheMapKey("e2edev@somecomp.com", "test-node-1"), NODE_POL_TYPE_CACHE, nodePol1)
	UpdateCache(NodeCacheMapKey("userdev", "test-node-3"), NODE_POL_TYPE_CACHE, nodePol2)
	UpdateCache(ServicePolicyCacheMapKey("e2edev@somecomp.com", "a-new-service", "amd64", "0.0.1"), SVC_POL_TYPE_CACHE, svcPol1)
	UpdateCache(ServicePolicyCacheMapKey("userdev", "another-service", "amd64", "0.0.0"), SVC_POL_TYPE_CACHE, svcPol2)

	change := ExchangeChange{OrgID: "e2edev@somecomp.com", ID: "test-node-2", Resource: "node"}

	DeleteCacheResourceFromChange(change, "")

	if cachedNode := GetNodeFromCache("e2edev@somecomp.com", "test-node-2"); cachedNode != nil {
		t.Errorf("Error: failed to remove resource from cache using an exchange change.")
	}

	change = ExchangeChange{OrgID: "e2edev@somecomp.com", ID: "e2edev@somecomp.com/e2edev@somecomp.com", Resource: "org", Operation: "created"}

	DeleteCacheResourceFromChange(change, "")

	if cachedSvc := GetServiceFromCache("e2edev@somecomp.com", "a-new-service", "amd64"); cachedSvc != nil {
		t.Errorf("Error: failed to remove service resource from cache after exchange org create change")
	} else if cachedSvc = GetServiceFromCache("userdev", "another-service", "amd64"); cachedSvc == nil {
		t.Errorf("Error: service userdev/another-service deleted from cache from change to a different org")
	} else if cachedNode := GetNodeFromCache("e2edev@somecomp.com", "test-node-1"); cachedNode != nil {
		t.Errorf("Error: failed to remove node resource from cache after exchange org create change")
	} else if cachedNode = GetNodeFromCache("userdev", "test-node-3"); cachedNode == nil {
		t.Errorf("Error: node test-node-3 removed by exchange org create change on a different org")
	} else if cachedNodePol := GetNodePolicyFromCache("e2edev@somecomp.com", "test-node-1"); cachedNodePol != nil {
		t.Errorf("Error: failed to remove node policy resource from cache after exchange org create change")
	} else if cachedNodePol = GetNodePolicyFromCache("userdev", "test-node-3"); cachedNodePol == nil {
		t.Errorf("Error: policy for node test-node-3 removed by exchange org create change on a different org")
	} else if cachedSvcPol := GetServicePolicyFromCache("e2edev@somecomp.com", "a-new-service", "amd64", "0.0.1"); cachedSvcPol != nil {
		t.Errorf("Error: failed to remove service policy resource from cache after exchange org create change")
	} else if cachedSvcPol = GetServicePolicyFromCache("userdev", "another-service", "amd64", "0.0.0"); cachedSvcPol == nil {
		t.Errorf("Error: policy for service another-service removed by exchange org create change on a different org")
	}
}

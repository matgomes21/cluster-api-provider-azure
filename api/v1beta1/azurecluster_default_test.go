/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceGroupDefault(t *testing.T) {
	cases := map[string]struct {
		cluster *AzureCluster
		output  *AzureCluster
	}{
		"default empty rg": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "foo",
				},
			},
		},
		"don't change if mismatched": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "bar",
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "bar",
				},
			},
		},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c.cluster.setResourceGroupDefault()
			if !reflect.DeepEqual(c.cluster, c.output) {
				expected, _ := json.MarshalIndent(c.output, "", "\t")
				actual, _ := json.MarshalIndent(c.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestVnetDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name:    "resource group vnet specified",
			cluster: createValidCluster(),
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							ResourceGroup: "custom-vnet",
							Name:          "my-vnet",
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{DefaultVnetCIDR},
							},
						},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
						APIServerLB: LoadBalancerSpec{
							Name: "my-lb",
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{
									{
										Name: "ip-config",
										PublicIP: &PublicIPSpec{
											Name:    "public-ip",
											DNSName: "myfqdn.azure.com",
										},
									},
								},
								Type: Public,
							},
						},
						NodeOutboundLB: &LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								FrontendIPsCount: to.Int32Ptr(1),
							},
						},
					},
				},
			},
		},
		{
			name: "vnet not specified",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							ResourceGroup: "cluster-test",
							Name:          "cluster-test-vnet",
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{DefaultVnetCIDR},
							},
						},
					},
				},
			},
		},
		{
			name: "custom CIDR",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{"10.0.0.0/16"},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							ResourceGroup: "cluster-test",
							Name:          "cluster-test-vnet",
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{"10.0.0.0/16"},
							},
						},
					},
				},
			},
		},
		{
			name: "IPv6 enabled",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{DefaultVnetCIDR, "2001:1234:5678:9a00::/56"},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							ResourceGroup: "cluster-test",
							Name:          "cluster-test-vnet",
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{DefaultVnetCIDR, "2001:1234:5678:9a00::/56"},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setVnetDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestSubnetDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name: "no subnets",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{DefaultControlPlaneSubnetCIDR},
								},
								Name:          "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{DefaultNodeSubnetCIDR},
								},
								Name:          "cluster-test-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
						},
					},
				},
			},
		},
		{
			name: "subnets with custom attributes",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{"10.0.0.16/24"},
								},
								Name: "my-controlplane-subnet",
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{"10.1.0.16/24"},
								},
								Name:       "my-node-subnet",
								NatGateway: NatGateway{Name: "foo-natgw"},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{"10.0.0.16/24"},
								},
								Name:          "my-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{"10.1.0.16/24"},
								},
								Name:          "my-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
								NatGateway: NatGateway{
									Name: "foo-natgw",
									NatGatewayIP: PublicIPSpec{
										Name: "pip-cluster-test-my-node-subnet-natgw",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "subnets specified",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name: "cluster-test-controlplane-subnet",
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name: "cluster-test-node-subnet",
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{DefaultControlPlaneSubnetCIDR},
								},

								Name:          "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{DefaultNodeSubnetCIDR},
								},
								Name:          "cluster-test-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
						},
					},
				},
			},
		},
		{
			name: "subnets route tables specified",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name: "cluster-test-controlplane-subnet",
								RouteTable: RouteTable{
									Name: "control-plane-custom-route-table",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name: "cluster-test-node-subnet",
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{DefaultControlPlaneSubnetCIDR},
								},
								Name:          "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{Name: "control-plane-custom-route-table"},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{DefaultNodeSubnetCIDR},
								},
								Name:          "cluster-test-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
						},
					},
				},
			},
		},
		{
			name: "only node subnet specified",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name: "my-node-subnet",
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{DefaultNodeSubnetCIDR},
								},
								Name:          "my-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{DefaultControlPlaneSubnetCIDR},
								},
								Name:          "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{},
							},
						},
					},
				},
			},
		},
		{
			name: "subnets specified with IPv6 enabled",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{"2001:be00::1/56"},
							},
						},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       "control-plane",
									CIDRBlocks: []string{"2001:beef::1/64"},
								},
								Name: "cluster-test-controlplane-subnet",
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       "node",
									CIDRBlocks: []string{"2001:beea::1/64"},
								},
								Name: "cluster-test-node-subnet",
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							VnetClassSpec: VnetClassSpec{
								CIDRBlocks: []string{"2001:be00::1/56"},
							},
						},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetControlPlane,
									CIDRBlocks: []string{"2001:beef::1/64"},
								},
								Name:          "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-controlplane-nsg"},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{"2001:beea::1/64"},
								},
								Name:          "cluster-test-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
						},
					},
				},
			},
		},
		{
			name: "subnets with custom security group",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: "control-plane",
								},
								Name: "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{
									SecurityGroupClass: SecurityGroupClass{
										SecurityRules: []SecurityRule{
											{
												Name:             "allow_port_50000",
												Description:      "allow port 50000",
												Protocol:         "*",
												Priority:         2202,
												SourcePorts:      to.StringPtr("*"),
												DestinationPorts: to.StringPtr("*"),
												Source:           to.StringPtr("*"),
												Destination:      to.StringPtr("*"),
											},
										},
									},
									Name: "my-custom-sg",
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       "control-plane",
									CIDRBlocks: []string{DefaultControlPlaneSubnetCIDR},
								},
								Name: "cluster-test-controlplane-subnet",
								SecurityGroup: SecurityGroup{
									SecurityGroupClass: SecurityGroupClass{
										SecurityRules: []SecurityRule{
											{
												Name:             "allow_port_50000",
												Description:      "allow port 50000",
												Protocol:         "*",
												Priority:         2202,
												SourcePorts:      to.StringPtr("*"),
												DestinationPorts: to.StringPtr("*"),
												Source:           to.StringPtr("*"),
												Destination:      to.StringPtr("*"),
												Direction:        SecurityRuleDirectionInbound,
											},
										},
									},
									Name: "my-custom-sg",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role:       SubnetNode,
									CIDRBlocks: []string{DefaultNodeSubnetCIDR},
								},
								Name:          "cluster-test-node-subnet",
								SecurityGroup: SecurityGroup{Name: "cluster-test-node-nsg"},
								RouteTable:    RouteTable{Name: "cluster-test-node-routetable"},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setSubnetDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestVnetPeeringDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name: "no peering",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{},
				},
			},
		},
		{
			name: "peering with resource group",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							Peerings: VnetPeerings{
								{
									RemoteVnetName: "my-vnet",
									ResourceGroup:  "cluster-test",
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							Peerings: VnetPeerings{
								{
									RemoteVnetName: "my-vnet",
									ResourceGroup:  "cluster-test",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "peering without resource group",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							Peerings: VnetPeerings{
								{
									RemoteVnetName: "my-vnet",
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					ResourceGroup: "cluster-test",
					NetworkSpec: NetworkSpec{
						Vnet: VnetSpec{
							Peerings: VnetPeerings{
								{
									RemoteVnetName: "my-vnet",
									ResourceGroup:  "cluster-test",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setVnetPeeringDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestAPIServerLBDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name: "no lb",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							Name: "cluster-test-public-lb",
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{
									{
										Name: "cluster-test-public-lb-frontEnd",
										PublicIP: &PublicIPSpec{
											Name:    "pip-cluster-test-apiserver",
											DNSName: "",
										},
									},
								},
								Type:                 Public,
								IdleTimeoutInMinutes: to.Int32Ptr(DefaultOutboundRuleIdleTimeoutInMinutes),
							},
						},
					},
				},
			},
		},
		{
			name: "internal lb",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Internal,
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{
									{
										Name: "cluster-test-internal-lb-frontEnd",
										FrontendIPClass: FrontendIPClass{
											PrivateIPAddress: DefaultInternalLBIPAddress,
										},
									},
								},
								Type:                 Internal,
								IdleTimeoutInMinutes: to.Int32Ptr(DefaultOutboundRuleIdleTimeoutInMinutes),
							},
							Name: "cluster-test-internal-lb",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setAPIServerLBDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestAzureEnviromentDefault(t *testing.T) {
	cases := map[string]struct {
		cluster *AzureCluster
		output  *AzureCluster
	}{
		"default empty azure env": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					AzureClusterClassSpec: AzureClusterClassSpec{
						AzureEnvironment: DefaultAzureCloud,
					},
				},
			},
		},
		"azure env set to AzurePublicCloud": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					AzureClusterClassSpec: AzureClusterClassSpec{
						AzureEnvironment: DefaultAzureCloud,
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					AzureClusterClassSpec: AzureClusterClassSpec{
						AzureEnvironment: DefaultAzureCloud,
					},
				},
			},
		},
		"azure env set to AzureGermanCloud": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					AzureClusterClassSpec: AzureClusterClassSpec{
						AzureEnvironment: "AzureGermanCloud",
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					AzureClusterClassSpec: AzureClusterClassSpec{
						AzureEnvironment: "AzureGermanCloud",
					},
				},
			},
		},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c.cluster.setAzureEnvironmentDefault()
			if !reflect.DeepEqual(c.cluster, c.output) {
				expected, _ := json.MarshalIndent(c.output, "", "\t")
				actual, _ := json.MarshalIndent(c.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestNodeOutboundLBDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name: "default lb for public clusters",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
						NodeOutboundLB: &LoadBalancerSpec{
							Name: "cluster-test",
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{{
									Name: "cluster-test-frontEnd",
									PublicIP: &PublicIPSpec{
										Name: "pip-cluster-test-node-outbound",
									},
								}},
								Type:                 Public,
								FrontendIPsCount:     to.Int32Ptr(1),
								IdleTimeoutInMinutes: to.Int32Ptr(DefaultOutboundRuleIdleTimeoutInMinutes),
							},
						},
					},
				},
			},
		},
		{
			name: "NAT gateway enabled - no LB",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
						},
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
					},
				},
			},
		},
		{
			name: "NAT gateway enabled on 1 of 2 node subnets",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
						NodeOutboundLB: &LoadBalancerSpec{
							Name: "cluster-test",
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{{
									Name: "cluster-test-frontEnd",
									PublicIP: &PublicIPSpec{
										Name: "pip-cluster-test-node-outbound",
									},
								}},
								Type:                 Public,
								FrontendIPsCount:     to.Int32Ptr(1),
								IdleTimeoutInMinutes: to.Int32Ptr(DefaultOutboundRuleIdleTimeoutInMinutes),
							},
						},
					},
				},
			},
		},
		{
			name: "multiple node subnets, NAT gateway not enabled in any of them",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-3",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-3",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
						},
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
						NodeOutboundLB: &LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{{
									Name: "cluster-test-frontEnd",
									PublicIP: &PublicIPSpec{
										Name: "pip-cluster-test-node-outbound",
									},
								}},
								Type:                 Public,
								FrontendIPsCount:     to.Int32Ptr(1),
								IdleTimeoutInMinutes: to.Int32Ptr(DefaultOutboundRuleIdleTimeoutInMinutes),
							},
							Name: "cluster-test",
						},
					},
				},
			},
		},
		{
			name: "multiple node subnets, NAT gateway enabled on all of them",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway-2",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-3",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway-3",
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						Subnets: Subnets{
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetControlPlane,
								},
								Name:          "control-plane-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-2",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway-2",
								},
							},
							{
								SubnetClassSpec: SubnetClassSpec{
									Role: SubnetNode,
								},
								Name:          "node-subnet-3",
								SecurityGroup: SecurityGroup{},
								RouteTable:    RouteTable{},
								NatGateway: NatGateway{
									Name: "node-natgateway-3",
								},
							},
						},
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
					},
				},
			},
		},
		{
			name: "no lb for private clusters",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Internal}},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Internal,
							},
						},
					},
				},
			},
		},
		{
			name: "NodeOutboundLB declared as input with non-default IdleTimeoutInMinutes and FrontendIPsCount values",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
						NodeOutboundLB: &LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								FrontendIPsCount:     to.Int32Ptr(2),
								IdleTimeoutInMinutes: to.Int32Ptr(15),
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
						NodeOutboundLB: &LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{
									{
										Name: "cluster-test-frontEnd-1",
										PublicIP: &PublicIPSpec{
											Name: "pip-cluster-test-node-outbound-1",
										},
									},
									{
										Name: "cluster-test-frontEnd-2",
										PublicIP: &PublicIPSpec{
											Name: "pip-cluster-test-node-outbound-2",
										},
									},
								},
								Type:                 Public,
								FrontendIPsCount:     to.Int32Ptr(2),  // we expect the original value to be respected here
								IdleTimeoutInMinutes: to.Int32Ptr(15), // we expect the original value to be respected here
							},
							Name: "cluster-test",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setNodeOutboundLBDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestControlPlaneOutboundLBDefaults(t *testing.T) {
	cases := []struct {
		name    string
		cluster *AzureCluster
		output  *AzureCluster
	}{
		{
			name: "no cp lb for public clusters",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Public}},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Public,
							},
						},
					},
				},
			},
		},
		{
			name: "no cp lb for private clusters",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Internal}},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Internal,
							},
						},
					},
				},
			},
		},
		{
			name: "frontendIPsCount > 1",
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{LoadBalancerClassSpec: LoadBalancerClassSpec{Type: Internal}},
						ControlPlaneOutboundLB: &LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								FrontendIPsCount:     to.Int32Ptr(2),
								IdleTimeoutInMinutes: to.Int32Ptr(15),
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-test",
				},
				Spec: AzureClusterSpec{
					NetworkSpec: NetworkSpec{
						APIServerLB: LoadBalancerSpec{
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								Type: Internal,
							},
						},
						ControlPlaneOutboundLB: &LoadBalancerSpec{
							Name: "cluster-test-outbound-lb",
							LoadBalancerClassSpec: LoadBalancerClassSpec{
								SKU: SKUStandard,
								FrontendIPs: []FrontendIP{
									{
										Name: "cluster-test-outbound-lb-frontEnd-1",
										PublicIP: &PublicIPSpec{
											Name: "pip-cluster-test-controlplane-outbound-1",
										},
									},
									{
										Name: "cluster-test-outbound-lb-frontEnd-2",
										PublicIP: &PublicIPSpec{
											Name: "pip-cluster-test-controlplane-outbound-2",
										},
									},
								},
								Type:                 Public,
								FrontendIPsCount:     to.Int32Ptr(2),
								IdleTimeoutInMinutes: to.Int32Ptr(15),
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		tc := c
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.cluster.setControlPlaneOutboundLBDefaults()
			if !reflect.DeepEqual(tc.cluster, tc.output) {
				expected, _ := json.MarshalIndent(tc.output, "", "\t")
				actual, _ := json.MarshalIndent(tc.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

func TestBastionDefault(t *testing.T) {
	cases := map[string]struct {
		cluster *AzureCluster
		output  *AzureCluster
	}{
		"no bastion set": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{},
			},
		},
		"azure bastion enabled with no settings": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "foo-azure-bastion",
							Subnet: SubnetSpec{
								Name: "AzureBastionSubnet",
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{DefaultAzureBastionSubnetCIDR},
									Role:       DefaultAzureBastionSubnetRole,
								},
							},
							PublicIP: PublicIPSpec{
								Name: "foo-azure-bastion-pip",
							},
						},
					},
				},
			},
		},
		"azure bastion enabled with name set": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "my-fancy-name",
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "my-fancy-name",
							Subnet: SubnetSpec{
								Name: "AzureBastionSubnet",
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{DefaultAzureBastionSubnetCIDR},
									Role:       DefaultAzureBastionSubnetRole,
								},
							},
							PublicIP: PublicIPSpec{
								Name: "foo-azure-bastion-pip",
							},
						},
					},
				},
			},
		},
		"azure bastion enabled with subnet partially set": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Subnet: SubnetSpec{},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "foo-azure-bastion",
							Subnet: SubnetSpec{
								Name: "AzureBastionSubnet",
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{DefaultAzureBastionSubnetCIDR},
									Role:       DefaultAzureBastionSubnetRole,
								},
							},
							PublicIP: PublicIPSpec{
								Name: "foo-azure-bastion-pip",
							},
						},
					},
				},
			},
		},
		"azure bastion enabled with subnet fully set": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Subnet: SubnetSpec{
								Name: "my-superfancy-name",
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{"10.10.0.0/16"},
								},
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "foo-azure-bastion",
							Subnet: SubnetSpec{
								Name: "my-superfancy-name",
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{"10.10.0.0/16"},
									Role:       DefaultAzureBastionSubnetRole,
								},
							},
							PublicIP: PublicIPSpec{
								Name: "foo-azure-bastion-pip",
							},
						},
					},
				},
			},
		},
		"azure bastion enabled with public IP name set": {
			cluster: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							PublicIP: PublicIPSpec{
								Name: "my-ultrafancy-pip-name",
							},
						},
					},
				},
			},
			output: &AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: AzureClusterSpec{
					BastionSpec: BastionSpec{
						AzureBastion: &AzureBastion{
							Name: "foo-azure-bastion",
							Subnet: SubnetSpec{
								SubnetClassSpec: SubnetClassSpec{
									CIDRBlocks: []string{DefaultAzureBastionSubnetCIDR},
									Role:       DefaultAzureBastionSubnetRole,
								},
								Name: "AzureBastionSubnet",
							},
							PublicIP: PublicIPSpec{
								Name: "my-ultrafancy-pip-name",
							},
						},
					},
				},
			},
		},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c.cluster.setBastionDefaults()
			if !reflect.DeepEqual(c.cluster, c.output) {
				expected, _ := json.MarshalIndent(c.output, "", "\t")
				actual, _ := json.MarshalIndent(c.cluster, "", "\t")
				t.Errorf("Expected %s, got %s", string(expected), string(actual))
			}
		})
	}
}

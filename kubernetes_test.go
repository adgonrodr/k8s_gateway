package gateway

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	nginx "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1"
	k8s_nginx "github.com/nginxinc/kubernetes-ingress/pkg/client/clientset/versioned"
	k8s_nginx_fake "github.com/nginxinc/kubernetes-ingress/pkg/client/clientset/versioned/fake"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayClient "sigs.k8s.io/gateway-api/pkg/client/clientset/gateway/versioned"
	gwFake "sigs.k8s.io/gateway-api/pkg/client/clientset/gateway/versioned/fake"
	"istio.io/api/meta/v1alpha1"
	"istio.io/api/networking/v1beta1"
	istioNetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istio "istio.io/client-go/pkg/clientset/versioned/fake"
)

func TestController(t *testing.T) {
	client := fake.NewSimpleClientset()
	gwClient := gwFake.NewSimpleClientset()
	nginxClient := k8s_nginx_fake.NewSimpleClientset()
	istioClient := istio.NewSimpleClientset()
	ctrl := &KubeController{
		client:      client,
		gwClient:    gwClient,
		nginxClient: nginxClient,
		hasSynced:   true,
	}
	addServices(client)
	addIngresses(client)
	addGateways(gwClient)
	addHTTPRoutes(gwClient)
	addVirtualServers(nginxClient)
	addIstioGateways(istioClient)

	gw := newGateway()
	gw.Zones = []string{"example.com."}
	gw.Next = test.NextHandler(dns.RcodeSuccess, nil)
	gw.Controller = ctrl

	for index, testObj := range testIngresses {
		found, _ := ingressHostnameIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("Ingress key %s not found in index: %v", index, found)
		}
	}

	for index, testObj := range testServices {
		found, _ := serviceHostnameIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("Service key %s not found in index: %v", index, found)
		}
	}

	for index, testObj := range testVirtualServers {
		found, _ := virtualServerHostnameIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("VirtualServer ksy %s not found in index: %v", index, found)
		}
	}

	for index, testObj := range testBadServices {
		found, _ := serviceHostnameIndexFunc(testObj)
		if isFound(index, found) {
			t.Errorf("Unexpected service key %s found in index: %v", index, found)
		}
	}

	for index, testObj := range testHTTPRoutes {
		found, _ := httpRouteHostnameIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("HTTPRoute key %s not found in index: %v", index, found)
		}
	}

	for index, testObj := range testGateways {
		found, _ := gatewayIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("Gateway key %s not found in index: %v", index, found)
		}
	}

	for index, testObj := range testIstioGateways {
		found, _ := istioGatewayHostnameIndexFunc(testObj)
		if !isFound(index, found) {
			t.Errorf("Istio Gatewat key %s not found in index: %v", index, found)
		}
	}
}

func addIstioGateways(istioClient *istio.Clientset) {
	ctx := context.TODO()
	for _, gateway := range testIstioGateways {
		_, err := istioClient.NetworkingV1beta1().Gateways("istions1").Create(ctx, gateway, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create Istio Gateway Objects :%s", err)
		}
	}
}

func isFound(s string, ss []string) bool {
	for _, str := range ss {
		if str == s {
			return true
		}
	}
	return false
}

func addServices(client kubernetes.Interface) {
	ctx := context.TODO()
	for _, svc := range testServices {
		_, err := client.CoreV1().Services("ns1").Create(ctx, svc, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create Service Objects :%s", err)
		}
	}
}

func addIngresses(client kubernetes.Interface) {
	ctx := context.TODO()
	for _, ingress := range testIngresses {
		_, err := client.NetworkingV1().Ingresses("ns1").Create(ctx, ingress, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create Ingress Objects :%s", err)
		}
	}
}

func addVirtualServers(client k8s_nginx.Interface) {
	ctx := context.TODO()
	for _, virtualServer := range testVirtualServers {
		_, err := client.K8sV1().VirtualServers("ns1").Create(ctx, virtualServer, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create VirtualServer Objects :%s", err)
		}
	}
}

func addGateways(client gatewayClient.Interface) {
	ctx := context.TODO()
	for _, gw := range testGateways {
		_, err := client.GatewayV1alpha2().Gateways("ns1").Create(ctx, gw, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create a Gateway Object :%s", err)
		}
	}
}

func addHTTPRoutes(client gatewayClient.Interface) {
	ctx := context.TODO()
	for _, r := range testHTTPRoutes {
		_, err := client.GatewayV1alpha2().HTTPRoutes("ns1").Create(ctx, r, meta.CreateOptions{})
		if err != nil {
			log.Warningf("Failed to Create a HTTPRoute Object :%s", err)
		}
	}
}

var testIngresses = map[string]*networking.Ingress{
	"a.example.org": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "ing1",
			Namespace: "ns1",
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: "a.example.org",
				},
			},
		},
		Status: networking.IngressStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.1"},
				},
			},
		},
	},
	"example.org": {
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: "example.org",
				},
			},
		},
		Status: networking.IngressStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.2"},
				},
			},
		},
	},
}

var testIstioGateways = map[string]*istioNetworkingv1beta1.Gateway{
	"istio.example.org": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "gateway1",
			Namespace: "istions1",
		},
		Spec: v1beta1.Gateway{Servers: []*v1beta1.Server{{
			Port:  &v1beta1.Port{Number: 80, Name: "http", Protocol: "http", TargetPort: 80},
			Hosts: []string{"istio.example.org"},
		}}},
		Status: v1alpha1.IstioStatus{},
	},
}

var testServices = map[string]*core.Service{
	"svc1.ns1": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "svc1",
			Namespace: "ns1",
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeLoadBalancer,
		},
		Status: core.ServiceStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.1"},
				},
			},
		},
	},
	"svc2.ns1": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "svc2",
			Namespace: "ns1",
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeLoadBalancer,
		},
		Status: core.ServiceStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.2"},
				},
			},
		},
	},
	"annotation": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "svc3",
			Namespace: "ns1",
			Annotations: map[string]string{
				"coredns.io/hostname": "annotation",
			},
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeLoadBalancer,
		},
		Status: core.ServiceStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.3"},
				},
			},
		},
	},
}

var testVirtualServers = map[string]*nginx.VirtualServer{
	"vs1.example.org": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "vs1",
			Namespace: "ns1",
		},
		Spec: nginx.VirtualServerSpec{
			Host: "vs1.example.org",
		},
		Status: nginx.VirtualServerStatus{
			ExternalEndpoints: []nginx.ExternalEndpoint{
				{IP: "192.0.0.1"},
			},
		},
	},
}

var testGateways = map[string]*gatewayapi_v1alpha2.Gateway{
	"ns1/gw-1": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "gw-1",
			Namespace: "ns1",
		},
		Spec: gatewayapi_v1alpha2.GatewaySpec{},
		Status: gatewayapi_v1alpha2.GatewayStatus{
			Addresses: []gatewayapi_v1alpha2.GatewayAddress{
				{
					Value: "192.0.2.100",
				},
			},
		},
	},
	"ns1/gw-2": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "gw-2",
			Namespace: "ns1",
		},
	},
}

var testHTTPRoutes = map[string]*gatewayapi_v1alpha2.HTTPRoute{
	"route-1.gw-1.example.com": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "route-1",
			Namespace: "ns1",
		},
		Spec: gatewayapi_v1alpha2.HTTPRouteSpec{
			//ParentRefs: []gatewayapi_v1alpha2.ParentRef{},
			Hostnames: []gatewayapi_v1alpha2.Hostname{"route-1.gw-1.example.com"},
		},
	},
}

var testBadServices = map[string]*core.Service{
	"svc1.ns2": {
		ObjectMeta: meta.ObjectMeta{
			Name:      "svc1",
			Namespace: "ns2",
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeClusterIP,
		},
		Status: core.ServiceStatus{
			LoadBalancer: core.LoadBalancerStatus{
				Ingress: []core.LoadBalancerIngress{
					{IP: "192.0.0.1"},
				},
			},
		},
	},
}

module github.com/ori-edge/k8s_gateway

go 1.16

require (
	github.com/coredns/caddy v1.1.0
	github.com/coredns/coredns v1.8.3
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/miekg/dns v1.1.41
	github.com/nginxinc/kubernetes-ingress v1.12.3
	google.golang.org/grpc v1.43.0 // indirect
	istio.io/api v0.0.0-20211122181927-8da52c66ff23 // indirect
	istio.io/client-go v1.12.1-0.20211122182438-57dcd971baaa
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/gateway-api v0.4.0
)

// https://github.com/etcd-io/etcd/issues/12124
replace google.golang.org/grpc v1.43.0 => google.golang.org/grpc v1.29.1

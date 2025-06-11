package k8s

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
)

const (
	// Label keys
	LabelAppKey          = "app"
	LabelEthNetworkKey   = "golem-base.io/eth-network"
	LabelOptimismRoleKey = "golem-base.io/optimism-role"

	// Role values
	RoleSequencerBootstrap = "sequencer-bootstrap"

	// Port names
	PortConductorRPC = "cndctr-rpc"
	PortNodeRPC      = "op-node-rpc"
	PortHTTP         = "http"

	// Network addresses
	RaftPort      = 50050
	KubeDNSSuffix = "svc.cluster.local"
	HTTPProtocol  = "http"

	// Default timeout for HTTP client
	DefaultHTTPTimeout = 30 * time.Second

	// Default port numbers
	DefaultConductorRPCPort = 8555
	DefaultNodeRPCPort      = 9545
)

// Client provides access to the Kubernetes API
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient creates a new Kubernetes client from a kubeconfig file or in-cluster config
func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig at %s: %w", kubeconfigPath, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client from config: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// makeAPIProxyURL generates a URL for accessing a service via the Kubernetes API proxy
func (c *Client) makeAPIProxyURL(namespace, serviceName string, portNumber int) string {
	// Remove any trailing slash from the API server host
	host := strings.TrimSuffix(c.config.Host, "/")

	// Create the proxy URL
	return fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%d/proxy/",
		host, namespace, serviceName, portNumber)
}

// AuthenticatedTransport creates an HTTP transport that handles Kubernetes API authentication
func (c *Client) AuthenticatedTransport() (http.RoundTripper, error) {
	transportConfig, err := c.config.TransportConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create transport config for Kubernetes API: %w", err)
	}

	// Create the transport with proper token refresh capability
	authTransport, err := transport.New(transportConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated transport for Kubernetes API: %w", err)
	}

	return authTransport, nil
}

// AuthenticatedHTTPClient returns an HTTP client that can authenticate to the Kubernetes API
func (c *Client) AuthenticatedHTTPClient() (*http.Client, error) {
	transport, err := rest.TransportFor(c.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP transport for Kubernetes API: %w", err)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   DefaultHTTPTimeout,
	}, nil
}

// DiscoverSequencers finds all sequencers in a namespace with optional label selector
func (c *Client) DiscoverSequencers(ctx context.Context, namespace, labelSelector string) ([]*SequencerResource, error) {
	// List StatefulSets with the given label selector
	statefulSets, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list StatefulSets in namespace %s with selector %s: %w", namespace, labelSelector, err)
	}

	// List Services to match with StatefulSets
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Services in namespace %s: %w", namespace, err)
	}

	// Create a map of app name to service
	serviceMap := make(map[string]corev1.Service)
	for _, svc := range services.Items {
		if appName, ok := svc.Labels[LabelAppKey]; ok {
			serviceMap[appName] = svc
		}
	}

	// Create the authenticated HTTP client for API access
	httpClient, err := c.AuthenticatedHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated HTTP client for Kubernetes API: %w", err)
	}

	var sequencers []*SequencerResource
	for _, sts := range statefulSets.Items {
		// Skip if not a sequencer (sanity check)
		role, hasRole := sts.Labels[LabelOptimismRoleKey]
		if !hasRole || (!strings.Contains(role, "sequencer")) {
			continue
		}

		// Get network name from labels
		network, hasNetwork := sts.Labels[LabelEthNetworkKey]
		if !hasNetwork {
			// Skip if no network is defined
			continue
		}

		// Determine if this is a bootstrap node
		isBootstrap := role == RoleSequencerBootstrap

		// Default to voting true for all sequencers unless explicitly marked
		isVoting := true

		// Find matching service
		var service corev1.Service
		var hasService bool
		if appName, ok := sts.Labels[LabelAppKey]; ok {
			service, hasService = serviceMap[appName]
		}

		if !hasService {
			return nil, fmt.Errorf("no matching service found for StatefulSet %s in namespace %s", sts.Name, namespace)
		}

		// Construct sequencer resource
		sequencer := &SequencerResource{
			Name:        sts.Name,
			Namespace:   sts.Namespace,
			Network:     network,
			Role:        role,
			IsBootstrap: isBootstrap,
			IsVoting:    isVoting,
			RPCURLs:     make(map[string]string),
			HTTPClient:  httpClient,
		}

		// Find port numbers from the service
		var conductorRPCPort, nodeRPCPort int

		for _, port := range service.Spec.Ports {
			switch port.Name {
			case PortConductorRPC:
				conductorRPCPort = int(port.Port)
			case PortNodeRPC:
				nodeRPCPort = int(port.Port)
			}
		}

		// Set default ports if not found
		if conductorRPCPort == 0 {
			conductorRPCPort = DefaultConductorRPCPort
		}
		if nodeRPCPort == 0 {
			nodeRPCPort = DefaultNodeRPCPort
		}

		// Set API proxy URLs instead of direct cluster DNS names
		sequencer.RPCURLs[PortConductorRPC] = c.makeAPIProxyURL(
			namespace, service.Name, conductorRPCPort)

		sequencer.RPCURLs[PortNodeRPC] = c.makeAPIProxyURL(
			namespace, service.Name, nodeRPCPort)

		// For Raft address, still use the internal DNS name
		// since it's not accessed via HTTP
		fqdn := fmt.Sprintf("%s.%s.%s", service.Name, namespace, KubeDNSSuffix)
		sequencer.RaftAddr = fmt.Sprintf("%s:%d", fqdn, RaftPort)

		sequencers = append(sequencers, sequencer)
	}

	return sequencers, nil
}

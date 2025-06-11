package provider

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/rpc"
	"github.com/golem-base/seqctl/pkg/sequencer"
)

// Constants for provider configuration
const (
	// Connection modes
	ConnectionModeAuto   = "auto"
	ConnectionModeProxy  = "proxy"
	ConnectionModeDirect = "direct"

	// Kubernetes DNS suffix
	K8sDNSSuffix = "svc.cluster.local"

	// Default timeouts
	DefaultHTTPTimeout      = 30 * time.Second
	DefaultSequencerTimeout = 10 * time.Second
)

// K8sProvider discovers sequencers from Kubernetes
//
// Service Discovery:
// The provider expects both StatefulSets and Services to have the same label selector
// (configured via k8s.selector). This ensures efficient discovery by only fetching
// relevant resources. The StatefulSet and its corresponding Service must share the
// same app label (configured via k8s.app_label) for matching.
type K8sProvider struct {
	clientset   *kubernetes.Clientset
	config      *rest.Config
	k8sConfig   config.K8sConfig
	httpClient  *http.Client
	logger      *slog.Logger
	isInCluster bool
	urlBuilder  *urlBuilder
}

// urlBuilder helps construct URLs based on connection context
type urlBuilder struct {
	config      *rest.Config
	isInCluster bool
	mode        string
}

// serviceEndpoint holds service connection information
type serviceEndpoint struct {
	namespace string
	name      string
	port      int
}

// IsInCluster detects if we're running inside a Kubernetes cluster
func IsInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

// NewK8sProvider creates a new Kubernetes provider
func NewK8sProvider(cfg *config.Config) (*K8sProvider, error) {
	logger := slog.Default().With(slog.String("provider", "k8s"))

	// Log the config path for debugging
	if cfg.K8s.ConfigPath != "" {
		logger.Debug("Using kubeconfig", "path", cfg.K8s.ConfigPath)
	}

	k8sConfig, err := buildK8sConfig(cfg.K8s.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	isInCluster := IsInCluster()
	httpClient, err := createHTTPClient(k8sConfig, cfg.K8s.ConnectionMode, isInCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Validate selectors before creating provider
	if err := validateK8sSelectors(cfg.K8s); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	provider := &K8sProvider{
		clientset:   clientset,
		config:      k8sConfig,
		k8sConfig:   cfg.K8s,
		httpClient:  httpClient,
		logger:      slog.Default().With(slog.String("provider", "k8s")),
		isInCluster: isInCluster,
		urlBuilder: &urlBuilder{
			config:      k8sConfig,
			isInCluster: isInCluster,
			mode:        cfg.K8s.ConnectionMode,
		},
	}

	provider.logger.Info("Kubernetes provider initialized",
		"connection_mode", cfg.K8s.ConnectionMode,
		"in_cluster", isInCluster,
		"kubeconfig", cfg.K8s.ConfigPath != "",
	)

	return provider, nil
}

// buildK8sConfig creates Kubernetes configuration from various sources
func buildK8sConfig(configPath string) (*rest.Config, error) {
	// Priority: explicit path > in-cluster > default locations
	if configPath != "" {
		// Path is already expanded at config loading time
		return clientcmd.BuildConfigFromFlags("", configPath)
	}

	if IsInCluster() {
		return rest.InClusterConfig()
	}

	// Try default kubeconfig locations
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

// createHTTPClient creates an HTTP client based on connection context
func createHTTPClient(k8sConfig *rest.Config, mode string, isInCluster bool) (*http.Client, error) {
	needsAuth := !isDirectInCluster(mode, isInCluster)

	if needsAuth {
		transport, err := rest.TransportFor(k8sConfig)
		if err != nil {
			return nil, err
		}
		return &http.Client{
			Transport: transport,
			Timeout:   DefaultHTTPTimeout,
		}, nil
	}

	return &http.Client{
		Timeout: DefaultHTTPTimeout,
	}, nil
}

// isDirectInCluster checks if we're using direct mode inside a cluster
func isDirectInCluster(mode string, isInCluster bool) bool {
	return mode == ConnectionModeDirect && isInCluster
}

// Name returns the provider type
func (p *K8sProvider) Name() string {
	return "kubernetes"
}

// DiscoverNetworks discovers all networks and their sequencers
func (p *K8sProvider) DiscoverNetworks(ctx context.Context) (map[string]*network.Network, error) {
	namespaces, err := p.getNamespacesToScan(ctx)
	if err != nil {
		return nil, err
	}

	networks := make(map[string]*network.Network)

	for _, namespace := range namespaces {
		p.logger.Debug("Scanning namespace", "namespace", namespace)

		sequencers, err := p.discoverSequencersInNamespace(ctx, namespace)
		if err != nil {
			p.logger.Warn("Failed to discover sequencers in namespace",
				"namespace", namespace, "error", err)
			continue
		}

		p.groupSequencersByNetwork(sequencers, networks, namespace)
	}

	return networks, nil
}

// getNamespacesToScan returns the list of namespaces to scan
func (p *K8sProvider) getNamespacesToScan(ctx context.Context) ([]string, error) {
	// If specific namespaces are configured, use those
	if len(p.k8sConfig.Namespaces) > 0 {
		return p.filterNamespacesByMode(ctx, p.k8sConfig.Namespaces)
	}

	// Otherwise, list all namespaces
	listOptions := metav1.ListOptions{}

	// Apply mode filter if configured
	if p.k8sConfig.SequencerModeFilter != "" {
		listOptions.LabelSelector = p.k8sConfig.SequencerModeFilter
		p.logger.Debug("Filtering namespaces by mode",
			"filter", p.k8sConfig.SequencerModeFilter)
	}

	nsList, err := p.clientset.CoreV1().Namespaces().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	if len(namespaces) == 0 && p.k8sConfig.SequencerModeFilter != "" {
		p.logger.Warn("No namespaces found matching mode filter",
			"filter", p.k8sConfig.SequencerModeFilter)
	}

	return namespaces, nil
}

// filterNamespacesByMode filters specific namespaces by mode label if configured
func (p *K8sProvider) filterNamespacesByMode(ctx context.Context, namespaces []string) ([]string, error) {
	if p.k8sConfig.SequencerModeFilter == "" {
		return namespaces, nil
	}

	// Parse the selector once
	selector, err := labels.Parse(p.k8sConfig.SequencerModeFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid sequencer mode filter: %w", err)
	}

	// For small lists, individual GETs might be more efficient
	// For larger lists, we should use a List with field selector
	if len(namespaces) > 10 {
		// Use field selector for bulk fetch
		fieldSelector := "metadata.name in (" + strings.Join(namespaces, ",") + ")"
		nsList, err := p.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
			LabelSelector: p.k8sConfig.SequencerModeFilter,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces with filter: %w", err)
		}

		filtered := make([]string, 0, len(nsList.Items))
		for _, ns := range nsList.Items {
			filtered = append(filtered, ns.Name)
		}
		return filtered, nil
	}

	// For small lists, use individual GETs
	filtered := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		namespace, err := p.clientset.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err != nil {
			p.logger.Warn("Failed to get namespace for mode filtering",
				"namespace", ns, "error", err)
			continue
		}

		// Check if namespace matches the selector
		if selector.Matches(labels.Set(namespace.Labels)) {
			filtered = append(filtered, ns)
		}
	}

	return filtered, nil
}

// groupSequencersByNetwork groups sequencers into their respective networks
func (p *K8sProvider) groupSequencersByNetwork(
	sequencers []*sequencer.Sequencer,
	networks map[string]*network.Network,
	namespace string,
) {
	for _, seq := range sequencers {
		networkName := seq.Network()
		if networkName == "" {
			p.logger.Warn("Sequencer has no network label",
				"sequencer", seq.ID(), "namespace", namespace)
			continue
		}

		if networks[networkName] == nil {
			networks[networkName] = network.NewNetwork(networkName, []*sequencer.Sequencer{})
		}

		// Add sequencer to network
		existingSeqs := networks[networkName].Sequencers()
		networks[networkName] = network.NewNetwork(networkName, append(existingSeqs, seq))
	}
}

// discoverSequencersInNamespace discovers sequencers in a specific namespace
func (p *K8sProvider) discoverSequencersInNamespace(ctx context.Context, namespace string) ([]*sequencer.Sequencer, error) {
	// Build label selector for StatefulSets
	stsSelector := p.k8sConfig.StatefulSetSelector
	if p.k8sConfig.SequencerModeFilter != "" {
		stsSelector = combineLabelSelectors(stsSelector, p.k8sConfig.SequencerModeFilter)
	}

	statefulSets, err := p.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: stsSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list StatefulSets: %w", err)
	}

	// Build label selector for Services
	svcSelector := p.k8sConfig.ServiceSelector
	if p.k8sConfig.SequencerModeFilter != "" {
		svcSelector = combineLabelSelectors(svcSelector, p.k8sConfig.SequencerModeFilter)
	}

	services, err := p.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: svcSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Services: %w", err)
	}

	serviceMap := p.buildServiceMap(services.Items)
	return p.createSequencersFromResources(namespace, statefulSets.Items, serviceMap)
}

// buildServiceMap creates a map of app name to service
func (p *K8sProvider) buildServiceMap(services []corev1.Service) map[string]*corev1.Service {
	serviceMap := make(map[string]*corev1.Service)
	for i := range services {
		svc := &services[i]
		if appName, ok := svc.Labels[p.k8sConfig.AppLabel]; ok {
			serviceMap[appName] = svc
		}
	}
	return serviceMap
}

// createSequencersFromResources creates sequencers from Kubernetes resources
func (p *K8sProvider) createSequencersFromResources(
	namespace string,
	statefulSets []appsv1.StatefulSet,
	serviceMap map[string]*corev1.Service,
) ([]*sequencer.Sequencer, error) {
	var sequencers []*sequencer.Sequencer

	for _, sts := range statefulSets {
		seq, err := p.processStatefulSet(namespace, &sts, serviceMap)
		if err != nil {
			if err == errSkipResource {
				continue
			}
			p.logger.Warn("Failed to create sequencer",
				"statefulset", sts.Name, "error", err)
			continue
		}
		sequencers = append(sequencers, seq)
	}

	return sequencers, nil
}

// errSkipResource is returned when a resource should be skipped (not an error)
var errSkipResource = fmt.Errorf("skip resource")

// processStatefulSet processes a single StatefulSet
func (p *K8sProvider) processStatefulSet(
	namespace string,
	sts *appsv1.StatefulSet,
	serviceMap map[string]*corev1.Service,
) (*sequencer.Sequencer, error) {
	// Validate network label
	networkName := sts.Labels[p.k8sConfig.NetworkLabel]
	if networkName == "" {
		p.logger.Debug("StatefulSet has no network label",
			"statefulset", sts.Name, "namespace", namespace)
		return nil, errSkipResource
	}

	// Find matching service
	service, err := p.findMatchingService(sts, serviceMap)
	if err != nil {
		return nil, err
	}

	return p.createSequencer(namespace, sts, service, networkName)
}

// findMatchingService finds the service for a StatefulSet
func (p *K8sProvider) findMatchingService(
	sts *appsv1.StatefulSet,
	serviceMap map[string]*corev1.Service,
) (*corev1.Service, error) {
	appName := sts.Labels[p.k8sConfig.AppLabel]
	service, found := serviceMap[appName]
	if !found {
		// Log available services for debugging
		var availableApps []string
		for app := range serviceMap {
			availableApps = append(availableApps, app)
		}
		p.logger.Debug("Service matching failed",
			"statefulset", sts.Name,
			"looking_for_app", appName,
			"available_apps", availableApps)
		return nil, fmt.Errorf("no matching service found for StatefulSet %s (app=%s)",
			sts.Name, appName)
	}
	return service, nil
}

// createSequencer creates a sequencer from Kubernetes resources
func (p *K8sProvider) createSequencer(
	namespace string,
	sts *appsv1.StatefulSet,
	svc *corev1.Service,
	networkName string,
) (*sequencer.Sequencer, error) {
	ports := p.extractPorts(svc)
	urls := p.buildURLs(namespace, svc.Name, ports)

	// Check if this sequencer is a voter based on label
	isVoter := false
	if role, ok := sts.Labels[p.k8sConfig.SequencerRoleLabel]; ok {
		// Check if the role value is in the list of voter values
		for _, voterValue := range p.k8sConfig.SequencerVoterValues {
			if role == voterValue {
				isVoter = true
				break
			}
		}
		p.logger.Debug("Sequencer role detected",
			"sequencer", sts.Name,
			"role", role,
			"is_voter", isVoter,
			"label_key", p.k8sConfig.SequencerRoleLabel,
			"voter_values", p.k8sConfig.SequencerVoterValues)
	} else {
		p.logger.Debug("No sequencer role label found, defaulting to non-voter",
			"sequencer", sts.Name,
			"label_key", p.k8sConfig.SequencerRoleLabel)
	}

	cfg := sequencer.Config{
		ID:           sts.Name,
		RaftAddr:     p.buildRaftAddress(namespace, svc.Name),
		ConductorURL: urls.conductor,
		NodeURL:      urls.node,
		Voting:       isVoter,
		Network:      networkName,
	}

	seq, err := sequencer.New(context.Background(), cfg,
		rpc.WithHTTPClient(p.selectHTTPClient()),
		rpc.WithTimeout(DefaultSequencerTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sequencer %s: %w", sts.Name, err)
	}

	return seq, nil
}

// portPair holds conductor and node ports
type portPair struct {
	conductor int
	node      int
}

// extractPorts extracts RPC ports from service definition
func (p *K8sProvider) extractPorts(svc *corev1.Service) portPair {
	ports := portPair{
		conductor: p.k8sConfig.ConductorPort,
		node:      p.k8sConfig.NodePort,
	}

	for _, port := range svc.Spec.Ports {
		switch port.Name {
		case p.k8sConfig.ConductorPortName:
			ports.conductor = int(port.Port)
		case p.k8sConfig.NodePortName:
			ports.node = int(port.Port)
		}
	}

	return ports
}

// urlPair holds conductor and node URLs
type urlPair struct {
	conductor string
	node      string
}

// buildURLs constructs the RPC URLs based on connection mode
func (p *K8sProvider) buildURLs(namespace, serviceName string, ports portPair) urlPair {
	conductorEP := serviceEndpoint{namespace, serviceName, ports.conductor}
	nodeEP := serviceEndpoint{namespace, serviceName, ports.node}

	return urlPair{
		conductor: p.urlBuilder.buildURL(conductorEP),
		node:      p.urlBuilder.buildURL(nodeEP),
	}
}

// buildURL constructs a URL for the given endpoint
func (ub *urlBuilder) buildURL(ep serviceEndpoint) string {
	if ub.shouldUseDirectConnection() {
		return ub.buildDirectURL(ep)
	}
	return ub.buildProxyURL(ep)
}

// shouldUseDirectConnection determines if direct connection should be used
func (ub *urlBuilder) shouldUseDirectConnection() bool {
	switch ub.mode {
	case ConnectionModeDirect:
		return ub.isInCluster
	case ConnectionModeProxy:
		return false
	default: // auto mode
		return ub.isInCluster
	}
}

// buildDirectURL builds a direct service URL
func (ub *urlBuilder) buildDirectURL(ep serviceEndpoint) string {
	return fmt.Sprintf("http://%s.%s.%s:%d",
		ep.name, ep.namespace, K8sDNSSuffix, ep.port)
}

// buildProxyURL builds a Kubernetes API proxy URL
func (ub *urlBuilder) buildProxyURL(ep serviceEndpoint) string {
	host := strings.TrimSuffix(ub.config.Host, "/")
	return fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%d/proxy/",
		host, ep.namespace, ep.name, ep.port)
}

// buildRaftAddress builds the Raft consensus address
func (p *K8sProvider) buildRaftAddress(namespace, serviceName string) string {
	return fmt.Sprintf("%s.%s.%s:%d",
		serviceName, namespace, K8sDNSSuffix, p.k8sConfig.RaftPort)
}

// selectHTTPClient returns the appropriate HTTP client for current context
func (p *K8sProvider) selectHTTPClient() *http.Client {
	if isDirectInCluster(p.k8sConfig.ConnectionMode, p.isInCluster) {
		return &http.Client{Timeout: DefaultSequencerTimeout}
	}
	return p.httpClient
}

// combineLabelSelectors combines two label selectors with comma separation
func combineLabelSelectors(selector1, selector2 string) string {
	if selector1 == "" {
		return selector2
	}
	if selector2 == "" {
		return selector1
	}
	return selector1 + "," + selector2
}

// matchesLabelSelector checks if labels match a label selector string
// Uses Kubernetes' official label selector parsing for full syntax support
func matchesLabelSelector(resourceLabels map[string]string, selectorStr string) bool {
	if selectorStr == "" {
		return true
	}

	// Parse the selector string using Kubernetes' parser
	selector, err := labels.Parse(selectorStr)
	if err != nil {
		// If we can't parse the selector, log and return false
		slog.Warn("Failed to parse label selector",
			"selector", selectorStr,
			"error", err)
		return false
	}

	// Check if the labels match the selector
	return selector.Matches(labels.Set(resourceLabels))
}

// validateK8sSelectors validates all Kubernetes label selectors in the configuration
func validateK8sSelectors(cfg config.K8sConfig) error {
	selectors := map[string]string{
		"statefulset_selector":  cfg.StatefulSetSelector,
		"service_selector":      cfg.ServiceSelector,
		"sequencer_mode_filter": cfg.SequencerModeFilter,
	}

	for name, selector := range selectors {
		if selector == "" {
			continue
		}
		if _, err := labels.Parse(selector); err != nil {
			return fmt.Errorf("invalid %s: %w", name, err)
		}
	}

	return nil
}

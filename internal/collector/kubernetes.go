package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/ThomasCrouzet/inframap-d2/internal/util"
)

func init() {
	Register(func() RegisteredCollector { return &KubernetesCollector{} })
}

// KubernetesCollector collects workloads from Kubernetes via kubectl.
type KubernetesCollector struct {
	Kubeconfig string
	Context    string
	Namespaces []string
	// TestData paths for testing (bypasses kubectl)
	TestPods      string
	TestServices  string
	TestIngresses string
}

func (kc *KubernetesCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "kubernetes",
		DisplayName: "Kubernetes",
		Description: "Collects pods, services, and ingresses from Kubernetes clusters",
		ConfigKey:   "kubernetes",
		DetectHint:  "kubectl",
	}
}

func (kc *KubernetesCollector) Enabled(sources map[string]any) bool {
	_, ok := sources["kubernetes"].(map[string]any)
	return ok
}

func (kc *KubernetesCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	if v, ok := section["kubeconfig"].(string); ok {
		kc.Kubeconfig = util.ExpandPath(v)
	}
	if v, ok := section["context"].(string); ok {
		kc.Context = v
	}
	if v, ok := section["namespaces"].([]any); ok {
		for _, ns := range v {
			if s, ok := ns.(string); ok {
				kc.Namespaces = append(kc.Namespaces, s)
			}
		}
	}
	return nil
}

func (kc *KubernetesCollector) Validate() []ValidationError {
	var errs []ValidationError
	if kc.Kubeconfig != "" {
		if _, err := os.Stat(kc.Kubeconfig); err != nil {
			errs = append(errs, ValidationError{
				Field:      "sources.kubernetes.kubeconfig",
				Message:    fmt.Sprintf("file not found: %s", kc.Kubeconfig),
				Suggestion: "check the path to your kubeconfig file",
			})
		}
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		errs = append(errs, ValidationError{
			Field:      "sources.kubernetes",
			Message:    "kubectl not found in PATH",
			Suggestion: "install kubectl: https://kubernetes.io/docs/tasks/tools/",
		})
	}
	return errs
}

// k8sPodList is the JSON structure returned by kubectl get pods.
type k8sPodList struct {
	Items []k8sPod `json:"items"`
}

type k8sPod struct {
	Metadata k8sMeta   `json:"metadata"`
	Spec     k8sPodSpec `json:"spec"`
	Status   struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

type k8sMeta struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type k8sPodSpec struct {
	Containers []k8sContainer `json:"containers"`
}

type k8sContainer struct {
	Name  string     `json:"name"`
	Image string     `json:"image"`
	Ports []k8sPort  `json:"ports"`
}

type k8sPort struct {
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// k8sServiceList is the JSON structure returned by kubectl get svc.
type k8sServiceList struct {
	Items []k8sService `json:"items"`
}

type k8sService struct {
	Metadata k8sMeta        `json:"metadata"`
	Spec     k8sServiceSpec `json:"spec"`
}

type k8sServiceSpec struct {
	Type     string           `json:"type"`
	Ports    []k8sServicePort `json:"ports"`
	Selector map[string]string `json:"selector"`
}

type k8sServicePort struct {
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	NodePort   int    `json:"nodePort"`
	Protocol   string `json:"protocol"`
}

// k8sIngressList is the JSON structure returned by kubectl get ingress.
type k8sIngressList struct {
	Items []k8sIngress `json:"items"`
}

type k8sIngress struct {
	Metadata k8sMeta       `json:"metadata"`
	Spec     k8sIngressSpec `json:"spec"`
}

type k8sIngressSpec struct {
	Rules []k8sIngressRule `json:"rules"`
}

type k8sIngressRule struct {
	Host string        `json:"host"`
	HTTP *k8sHTTPRule  `json:"http"`
}

type k8sHTTPRule struct {
	Paths []k8sHTTPPath `json:"paths"`
}

type k8sHTTPPath struct {
	Path    string          `json:"path"`
	Backend k8sIngressBackend `json:"backend"`
}

type k8sIngressBackend struct {
	Service k8sBackendService `json:"service"`
}

type k8sBackendService struct {
	Name string `json:"name"`
	Port struct {
		Number int `json:"number"`
	} `json:"port"`
}

func (kc *KubernetesCollector) Collect(infra *model.Infrastructure) error {
	pods, err := kc.getPods()
	if err != nil {
		return fmt.Errorf("getting pods: %w", err)
	}

	services, err := kc.getServices()
	if err != nil {
		return fmt.Errorf("getting services: %w", err)
	}

	ingresses, err := kc.getIngresses()
	if err != nil {
		return fmt.Errorf("getting ingresses: %w", err)
	}

	// Build a port map from services: svcName@ns → port
	svcPorts := make(map[string]int)
	for _, svc := range services.Items {
		if len(svc.Spec.Ports) > 0 {
			key := svc.Metadata.Name + "@" + svc.Metadata.Namespace
			port := svc.Spec.Ports[0].Port
			if svc.Spec.Ports[0].NodePort > 0 {
				port = svc.Spec.Ports[0].NodePort
			}
			svcPorts[key] = port
		}
	}

	// Build ingress map: svcName@ns → host
	ingressHosts := make(map[string]string)
	for _, ing := range ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				key := path.Backend.Service.Name + "@" + ing.Metadata.Namespace
				ingressHosts[key] = rule.Host
			}
		}
	}

	// Group pods by namespace → create a server per namespace
	nsPods := make(map[string][]k8sPod)
	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running" {
			continue
		}
		ns := pod.Metadata.Namespace
		if len(kc.Namespaces) > 0 && !containsStr(kc.Namespaces, ns) {
			continue
		}
		nsPods[ns] = append(nsPods[ns], pod)
	}

	// Deduplicate pods by app label (deployment pods share the same label)
	for ns, podList := range nsPods {
		serverName := fmt.Sprintf("k8s-%s", ns)

		server, exists := infra.Servers[serverName]
		if !exists {
			server = &model.Server{
				Hostname: serverName,
				Label:    fmt.Sprintf("k8s/%s", ns),
				Type:     model.ServerTypeCluster,
				Online:   true,
			}
			infra.Servers[serverName] = server
		}

		seen := make(map[string]bool)
		for _, pod := range podList {
			// Use app label as service name for dedup, fall back to container name
			for _, container := range pod.Spec.Containers {
				svcName := container.Name
				if appLabel, ok := pod.Metadata.Labels["app"]; ok {
					svcName = appLabel
				}

				if seen[svcName] {
					continue
				}
				seen[svcName] = true

				svc := &model.Service{
					Name:  svcName,
					Image: container.Image,
					Type:  detectServiceType(container.Image, svcName),
				}

				// Add port from k8s service if available
				key := svcName + "@" + ns
				if port, ok := svcPorts[key]; ok {
					svc.Ports = append(svc.Ports, model.PortMapping{
						HostPort:      port,
						ContainerPort: port,
						Protocol:      "tcp",
					})
				} else if len(container.Ports) > 0 {
					svc.Ports = append(svc.Ports, model.PortMapping{
						ContainerPort: container.Ports[0].ContainerPort,
						Protocol:      strings.ToLower(container.Ports[0].Protocol),
					})
				}

				svc.Category = "kubernetes"
				server.AddService(svc)
			}
		}
	}

	return nil
}

func (kc *KubernetesCollector) getPods() (*k8sPodList, error) {
	var result k8sPodList
	if kc.TestPods != "" {
		if err := loadJSONFile(kc.TestPods, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if err := kc.kubectlGet(&result, "get", "pods", "-A", "-o", "json"); err != nil {
		return nil, err
	}
	return &result, nil
}

func (kc *KubernetesCollector) getServices() (*k8sServiceList, error) {
	var result k8sServiceList
	if kc.TestServices != "" {
		if err := loadJSONFile(kc.TestServices, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if err := kc.kubectlGet(&result, "get", "svc", "-A", "-o", "json"); err != nil {
		return nil, err
	}
	return &result, nil
}

func (kc *KubernetesCollector) getIngresses() (*k8sIngressList, error) {
	var result k8sIngressList
	if kc.TestIngresses != "" {
		if err := loadJSONFile(kc.TestIngresses, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if err := kc.kubectlGet(&result, "get", "ingress", "-A", "-o", "json"); err != nil {
		return nil, err
	}
	return &result, nil
}

func (kc *KubernetesCollector) kubectlGet(result any, args ...string) error {
	cmdArgs := args
	if kc.Kubeconfig != "" {
		cmdArgs = append([]string{"--kubeconfig", kc.Kubeconfig}, cmdArgs...)
	}
	if kc.Context != "" {
		cmdArgs = append([]string{"--context", kc.Context}, cmdArgs...)
	}

	cmd := exec.Command("kubectl", cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("kubectl %s: %w", strings.Join(args, " "), err)
	}

	if err := json.Unmarshal(out, result); err != nil {
		return fmt.Errorf("parsing kubectl output: %w", err)
	}
	return nil
}

func loadJSONFile(path string, result any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func containsStr(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

package src

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// HealthStatus represents the health check status of a registered service instance.
type HealthStatus string

const (
	HealthPassing  HealthStatus = "PASSING"
	HealthWarning  HealthStatus = "WARNING"
	HealthCritical HealthStatus = "CRITICAL"
)

// ServiceInstance represents an active endpoint discovered via Consul DNS or Catalog.
type ServiceInstance struct {
	ID          string            `json:"id"`
	ServiceName string            `json:"service_name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Priority    int               `json:"priority"` // Lower value = higher priority (RFC 2782)
	Weight      int               `json:"weight"`   // For load balancing among same priority
	Status      HealthStatus      `json:"status"`
	LatencyMs   float64           `json:"latency_ms"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	LastChecked time.Time         `json:"last_checked"`
}

// Address returns host:port formatted string.
func (s *ServiceInstance) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// ServiceDiscoveryResolver defines the discovery contract for Consul DNS / API.
type ServiceDiscoveryResolver interface {
	Resolve(serviceName string) ([]*ServiceInstance, error)
	GetHealthyInstance(serviceName string) (*ServiceInstance, error)
	RegisterInstance(inst *ServiceInstance) error
	DeregisterInstance(instanceID string) error
	UpdateHealth(instanceID string, status HealthStatus, latencyMs float64) error
}

// ConsulDiscoveryClient implements in-memory cached Consul DNS service discovery
// with RFC 2782 SRV record semantics and failover routing.
type ConsulDiscoveryClient struct {
	mu           sync.RWMutex
	catalog      map[string][]*ServiceInstance // serviceName -> list of instances
	cacheTTL     time.Duration
	lastResolved map[string]time.Time
	rrIndex      map[string]*uint64 // serviceName -> atomic round robin counter
}

// NewConsulDiscoveryClient initializes a new Consul discovery client.
func NewConsulDiscoveryClient(cacheTTL time.Duration) *ConsulDiscoveryClient {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Second
	}
	return &ConsulDiscoveryClient{
		catalog:      make(map[string][]*ServiceInstance),
		cacheTTL:     cacheTTL,
		lastResolved: make(map[string]time.Time),
		rrIndex:      make(map[string]*uint64),
	}
}

// RegisterInstance registers a new microservice node into the registry.
func (c *ConsulDiscoveryClient) RegisterInstance(inst *ServiceInstance) error {
	if inst == nil || inst.ID == "" || inst.ServiceName == "" || inst.Host == "" || inst.Port <= 0 {
		return errors.New("invalid service instance: mandatory fields missing")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	instances := c.catalog[inst.ServiceName]
	// Check if already exists, update if so
	for i, existing := range instances {
		if existing.ID == inst.ID {
			instances[i] = inst
			return nil
		}
	}

	c.catalog[inst.ServiceName] = append(instances, inst)
	if _, ok := c.rrIndex[inst.ServiceName]; !ok {
		var zero uint64
		c.rrIndex[inst.ServiceName] = &zero
	}
	return nil
}

// DeregisterInstance removes an instance from all registered catalogs.
func (c *ConsulDiscoveryClient) DeregisterInstance(instanceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	found := false
	for svcName, instances := range c.catalog {
		var filtered []*ServiceInstance
		for _, inst := range instances {
			if inst.ID == instanceID {
				found = true
				continue
			}
			filtered = append(filtered, inst)
		}
		c.catalog[svcName] = filtered
	}

	if !found {
		return fmt.Errorf("instance %s not found", instanceID)
	}
	return nil
}

// UpdateHealth updates health and latency metrics for an instance.
func (c *ConsulDiscoveryClient) UpdateHealth(instanceID string, status HealthStatus, latencyMs float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, instances := range c.catalog {
		for _, inst := range instances {
			if inst.ID == instanceID {
				inst.Status = status
				inst.LatencyMs = latencyMs
				inst.LastChecked = time.Now().UTC()
				return nil
			}
		}
	}
	return fmt.Errorf("instance %s not found for health update", instanceID)
}

// Resolve returns all healthy instances for a service.
func (c *ConsulDiscoveryClient) Resolve(serviceName string) ([]*ServiceInstance, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances, ok := c.catalog[serviceName]
	if !ok || len(instances) == 0 {
		return nil, fmt.Errorf("service %s not found in registry", serviceName)
	}

	var healthy []*ServiceInstance
	for _, inst := range instances {
		if inst.Status == HealthPassing {
			healthy = append(healthy, inst)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy instances available for service %s", serviceName)
	}
	return healthy, nil
}

// GetHealthyInstance selects an instance using weighted round-robin among lowest priority (RFC 2782).
func (c *ConsulDiscoveryClient) GetHealthyInstance(serviceName string) (*ServiceInstance, error) {
	healthy, err := c.Resolve(serviceName)
	if err != nil {
		return nil, err
	}

	// Filter by lowest priority number
	minPriority := healthy[0].Priority
	for _, inst := range healthy {
		if inst.Priority < minPriority {
			minPriority = inst.Priority
		}
	}

	var topTier []*ServiceInstance
	for _, inst := range healthy {
		if inst.Priority == minPriority {
			topTier = append(topTier, inst)
		}
	}

	c.mu.RLock()
	counterPtr, ok := c.rrIndex[serviceName]
	c.mu.RUnlock()

	if !ok || counterPtr == nil {
		return topTier[rand.Intn(len(topTier))], nil
	}

	idx := atomic.AddUint64(counterPtr, 1) - 1
	return topTier[idx%uint64(len(topTier))], nil
}

// GetLowestLatencyInstance returns the healthy node with the lowest observed latency.
func (c *ConsulDiscoveryClient) GetLowestLatencyInstance(serviceName string) (*ServiceInstance, error) {
	healthy, err := c.Resolve(serviceName)
	if err != nil {
		return nil, err
	}

	best := healthy[0]
	for _, inst := range healthy[1:] {
		if inst.LatencyMs < best.LatencyMs {
			best = inst
		}
	}
	return best, nil
}

package router

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"api-v2ray/internal/model"
)

type Candidate struct {
	Upstream model.Upstream
	Binding  model.Binding
	Node     model.ProxyNode
}

type CandidateHealth struct {
	Key                 string    `json:"key"`
	UpstreamID          string    `json:"upstream_id"`
	BindingID           string    `json:"binding_id"`
	NodeID              string    `json:"node_id"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	TotalSuccesses      int       `json:"total_successes"`
	TotalFailures       int       `json:"total_failures"`
	LastFailureKind     string    `json:"last_failure_kind,omitempty"`
	LastError           string    `json:"last_error,omitempty"`
	LastUsedAt          time.Time `json:"last_used_at,omitempty"`
	CooldownUntil       time.Time `json:"cooldown_until,omitempty"`
	IsCoolingDown       bool      `json:"is_cooling_down"`
	CooldownSeconds     int       `json:"cooldown_seconds"`
}

type candidateRuntimeState struct {
	ConsecutiveFailures int
	TotalSuccesses      int
	TotalFailures       int
	LastFailureKind     string
	LastError           string
	LastUsedAt          time.Time
	CooldownUntil       time.Time
}

type FailureKind string

const (
	FailureProxyRegistryError FailureKind = "proxy_registry_error"
	FailureProxyConnectError  FailureKind = "proxy_connect_error"
	FailureUpstreamTimeout    FailureKind = "upstream_timeout"
	FailureUpstream429        FailureKind = "upstream_429"
	FailureUpstream5xx        FailureKind = "upstream_5xx"
	FailureUpstream4xx        FailureKind = "upstream_4xx"
	FailureUpstreamAuthError  FailureKind = "upstream_auth_error"
	FailureModelNotFound      FailureKind = "model_not_found"
	FailureUpstreamUnknown    FailureKind = "upstream_unknown_error"
)

type Service struct {
	upstreams map[string]model.Upstream
	bindings  map[string]model.Binding
	nodes     map[string]model.ProxyNode
	failover  model.FailoverConfig

	mu        sync.Mutex
	rrCounter map[string]int
	state     map[string]*candidateRuntimeState
}

func New(cfg *model.Config) *Service {
	s := &Service{
		upstreams: map[string]model.Upstream{},
		bindings:  map[string]model.Binding{},
		nodes:     map[string]model.ProxyNode{},
		failover:  normalizeFailover(cfg.Failover),
		rrCounter: map[string]int{},
		state:     map[string]*candidateRuntimeState{},
	}

	for _, u := range cfg.Upstreams {
		s.upstreams[u.ID] = u
	}
	for _, b := range cfg.Bindings {
		s.bindings[b.ID] = b
	}
	for _, n := range cfg.ProxyNodes {
		s.nodes[n.ID] = n
	}

	return s
}

func (s *Service) ResolveByModel(modelName string) (*model.Upstream, *model.Binding, *model.ProxyNode, error) {
	candidates, err := s.ResolveCandidatesByModel(modelName)
	if err != nil {
		return nil, nil, nil, err
	}
	c := candidates[0]
	uc, bc, nc := c.Upstream, c.Binding, c.Node
	return &uc, &bc, &nc, nil
}

func (s *Service) ResolveCandidatesByModel(modelName string) ([]Candidate, error) {
	var matched []Candidate
	for _, u := range s.upstreams {
		if !u.Enabled {
			continue
		}
		for _, m := range u.Models {
			if !strings.EqualFold(m, modelName) {
				continue
			}
			binding, ok := s.bindings[u.BindingID]
			if !ok {
				return nil, fmt.Errorf("binding not found for upstream %s", u.ID)
			}
			node, ok := s.nodes[binding.NodeID]
			if !ok {
				return nil, fmt.Errorf("proxy node not found for binding %s", binding.ID)
			}
			matched = append(matched, Candidate{Upstream: u, Binding: binding, Node: node})
			break
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no upstream found for model %s", modelName)
	}

	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Upstream.Priority != matched[j].Upstream.Priority {
			return matched[i].Upstream.Priority > matched[j].Upstream.Priority
		}
		if matched[i].Upstream.ID != matched[j].Upstream.ID {
			return matched[i].Upstream.ID < matched[j].Upstream.ID
		}
		return matched[i].Binding.ID < matched[j].Binding.ID
	})

	return s.applyRuntimeOrdering(modelName, matched), nil
}

func (s *Service) applyRuntimeOrdering(modelName string, candidates []Candidate) []Candidate {
	if len(candidates) <= 1 {
		return candidates
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Candidate, 0, len(candidates))
	now := time.Now()
	for i := 0; i < len(candidates); {
		j := i + 1
		for j < len(candidates) && candidates[j].Upstream.Priority == candidates[i].Upstream.Priority {
			j++
		}
		group := append([]Candidate(nil), candidates[i:j]...)

		available := make([]Candidate, 0, len(group))
		cooling := make([]Candidate, 0, len(group))
		for _, c := range group {
			st := s.getStateLocked(candidateKey(c))
			if !st.CooldownUntil.IsZero() && now.Before(st.CooldownUntil) {
				cooling = append(cooling, c)
			} else {
				available = append(available, c)
			}
		}

		if len(available) > 1 {
			key := fmt.Sprintf("%s|p=%d", strings.ToLower(modelName), group[0].Upstream.Priority)
			start := s.rrCounter[key] % len(available)
			s.rrCounter[key] = (s.rrCounter[key] + 1) % len(available)
			available = append(available[start:], available[:start]...)
		}
		sort.SliceStable(cooling, func(a, b int) bool {
			ca := s.getStateLocked(candidateKey(cooling[a])).CooldownUntil
			cb := s.getStateLocked(candidateKey(cooling[b])).CooldownUntil
			return ca.Before(cb)
		})
		out = append(out, available...)
		out = append(out, cooling...)
		i = j
	}
	return out
}

func candidateKey(c Candidate) string {
	return c.Upstream.ID + "|" + c.Binding.ID + "|" + c.Node.ID
}

func (s *Service) getStateLocked(key string) *candidateRuntimeState {
	st, ok := s.state[key]
	if !ok {
		st = &candidateRuntimeState{}
		s.state[key] = st
	}
	return st
}

func (s *Service) MarkFailure(c Candidate, kind FailureKind, errText string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.getStateLocked(candidateKey(c))
	st.TotalFailures++
	st.ConsecutiveFailures++
	st.LastFailureKind = string(kind)
	st.LastError = strings.TrimSpace(errText)
	st.LastUsedAt = time.Now()
	st.CooldownUntil = time.Now().Add(s.cooldownForFailure(kind, st.ConsecutiveFailures))
}

func (s *Service) MarkSuccess(c Candidate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.getStateLocked(candidateKey(c))
	st.TotalSuccesses++
	st.ConsecutiveFailures = 0
	st.LastFailureKind = ""
	st.LastError = ""
	st.LastUsedAt = time.Now()
	st.CooldownUntil = time.Time{}
}

func normalizeFailover(cfg model.FailoverConfig) model.FailoverConfig {
	if len(cfg.CooldownSteps) == 0 {
		cfg.CooldownSteps = []model.CooldownStep{
			{AfterFailures: 1, DurationSeconds: 10},
			{AfterFailures: 2, DurationSeconds: 30},
			{AfterFailures: 3, DurationSeconds: 90},
			{AfterFailures: 5, DurationSeconds: 300},
		}
	}
	sort.Slice(cfg.CooldownSteps, func(i, j int) bool {
		return cfg.CooldownSteps[i].AfterFailures < cfg.CooldownSteps[j].AfterFailures
	})
	return cfg
}

func (s *Service) cooldownForFailures(n int) time.Duration {
	if n <= 0 {
		return 0
	}
	d := 0
	for _, step := range s.failover.CooldownSteps {
		if n >= step.AfterFailures {
			d = step.DurationSeconds
		}
	}
	return time.Duration(d) * time.Second
}

func (s *Service) cooldownForFailure(kind FailureKind, n int) time.Duration {
	switch kind {
	case FailureUpstreamAuthError, FailureModelNotFound, FailureUpstream4xx:
		return 0
	default:
		return s.cooldownForFailures(n)
	}
}

func (s *Service) HealthSnapshot() []CandidateHealth {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	out := make([]CandidateHealth, 0, len(s.upstreams))
	for _, u := range s.upstreams {
		if !u.Enabled {
			continue
		}
		binding, ok := s.bindings[u.BindingID]
		if !ok {
			continue
		}
		node, ok := s.nodes[binding.NodeID]
		if !ok {
			continue
		}
		c := Candidate{Upstream: u, Binding: binding, Node: node}
		st := s.getStateLocked(candidateKey(c))
		cd := 0
		cooling := !st.CooldownUntil.IsZero() && now.Before(st.CooldownUntil)
		if cooling {
			cd = int(time.Until(st.CooldownUntil).Seconds())
			if cd < 0 {
				cd = 0
			}
		}
		out = append(out, CandidateHealth{
			Key:                 candidateKey(c),
			UpstreamID:          u.ID,
			BindingID:           binding.ID,
			NodeID:              node.ID,
			ConsecutiveFailures: st.ConsecutiveFailures,
			TotalSuccesses:      st.TotalSuccesses,
			TotalFailures:       st.TotalFailures,
			LastFailureKind:     st.LastFailureKind,
			LastError:           st.LastError,
			LastUsedAt:          st.LastUsedAt,
			CooldownUntil:       st.CooldownUntil,
			IsCoolingDown:       cooling,
			CooldownSeconds:     cd,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsCoolingDown != out[j].IsCoolingDown {
			return !out[i].IsCoolingDown
		}
		if out[i].ConsecutiveFailures != out[j].ConsecutiveFailures {
			return out[i].ConsecutiveFailures < out[j].ConsecutiveFailures
		}
		return out[i].UpstreamID < out[j].UpstreamID
	})
	return out
}

func (s *Service) Models() []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range s.upstreams {
		if !u.Enabled {
			continue
		}
		for _, m := range u.Models {
			name := strings.TrimSpace(m)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

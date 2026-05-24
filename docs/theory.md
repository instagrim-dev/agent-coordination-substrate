# Theory: Intellectual Lineage

The Agent Coordination Substrate draws from four traditions. This document acknowledges the lineage and explains what we adopted, adapted, and rejected from each.

---

## 1. Stigmergy

**Origin:** Pierre-Paul Grassé (1959), studying termite nest construction. Extended by Theraulaz & Bonabeau (1999) to general multi-agent coordination.

**Core insight:** Agents need not communicate directly. Instead, they modify a shared environment. Other agents perceive those modifications and alter their behavior accordingly. Coordination is _indirect_ — mediated by the environment rather than by messages between agents.

**What we adopted:**
- The fundamental mechanism: deposit state into the environment, observe environment, react
- Mortality of deposits (signals decay, aren't permanent)
- Zone-based locality (not global broadcast — signals exist in named zones)
- Emergent coordination without explicit coupling between agents

**What we adapted:**
- Biological metaphor → engineering vocabulary (`signal` not `pheromone`, `zone` not `nest site`, `deposit` not `emit`)
- Simple quantity accumulation → typed, structured signals with metadata
- Fixed decay → configurable TTL with precise garbage collection
- Passive observation → active induction (threshold → materialized artifact)
- Single modality → two tiers (advisory + enforcement)

**What we rejected:**
- "Pheromone" terminology (imprecise, misleading for engineering use)
- Agent homogeneity assumption (our agents are heterogeneous in capability)
- Purely emergent coordination (we add enforcement for safety-critical boundaries)
- Gradient-following as the only response (our agents have structured readouts, not just intensity signals)

**Key references:**
- Grassé, P-P. (1959). La reconstruction du nid et les coordinations interindividuelles chez _Bellicositermes natalensis_ et _Cubitermes_ sp.
- Theraulaz, G. & Bonabeau, E. (1999). A brief history of stigmergy. _Artificial Life_, 5(2), 97-116.
- Dorigo, M. et al. (2000). Ant algorithms and stigmergy. _Future Generation Computer Systems_, 16(8), 851-871.

---

## 2. Distributed Leases

**Origin:** Gray & Cheriton (1989), "Leases: An Efficient Fault-Tolerant Mechanism for Distributed File Cache Consistency."

**Core insight:** A time-bounded contract between a holder and the environment. The holder may assume exclusive access only while the lease is valid. When the lease expires, the assumption is void — regardless of whether the holder explicitly released.

**What we adopted:**
- Time-bounded contracts (mandatory mortality on all claims)
- Holder identity (actor_id is required on every claim)
- Automatic expiry without holder action (crash-tolerant)
- Renewal as an explicit operation (not automatic)

**What we adapted:**
- File-cache scope → workspace zone scope (claims address named zones)
- Binary grant/deny → typed claims with structured metadata
- Single resource → hierarchical zone addressing (glob-matchable)

**What we rejected:**
- Consensus-based lease election (too heavyweight for single-host workspace)
- Lease transfer between holders (claims are non-transferable; release and re-claim)
- Indefinite leases (every claim MUST have a finite TTL)

**Key references:**
- Gray, C. & Cheriton, D. (1989). Leases: An Efficient Fault-Tolerant Mechanism for Distributed File Cache Consistency. _SOSP_.
- Burrows, M. (2006). The Chubby lock service for loosely-coupled distributed systems. _OSDI_.
- Kleppmann, M. (2016). _Designing Data-Intensive Applications_, Chapter 8 (fencing tokens).
- Antirez (2016). How to do distributed locking. (Redis Redlock analysis).

---

## 3. Resilience Engineering

**Origin:** Nygard (2007) _Release It!_, Envoy proxy patterns, Kubernetes admission controllers.

**Core insight:** Systems need **self-healing boundaries** — mechanisms that automatically engage under pressure and automatically disengage when pressure subsides. Circuit breakers, bulkheads, and rate limiters share a pattern: time-bounded, stateful, and autonomic.

**What we adopted:**
- Self-healing property (enforcement that automatically releases under time pressure)
- Capacity isolation (claims prevent resource exhaustion by containing blast radius)
- Progressive enforcement (advisory → enforcement escalation)
- Transparent state (operators can inspect and override breaker state)

**What we adapted:**
- Service-to-service scope → workspace scope (agent-to-environment)
- Binary open/closed → typed enforcement with structured reason codes
- Automatic reset after timeout → operator-dismissable with explicit override

**What we rejected:**
- Purely reactive patterns (our enforcement can be _proactive_ — claims before conflict)
- Global thresholds (our enforcement is zone-scoped)
- Service mesh sidecar deployment model (our substrate is process-internal)

**Key references:**
- Nygard, M. (2007/2018). _Release It!_ Second Edition. (Circuit breaker, bulkhead patterns).
- Envoy proxy documentation: Circuit breaking, rate limiting.
- Kubernetes: ValidatingAdmissionPolicy, ResourceQuota.
- Netflix Hystrix: Circuit breaker implementation patterns.

---

## 4. Urban Governance

**Origin:** Municipal permit systems, zoning regulations, temporary use agreements.

**Core insight:** Cities manage shared space through **permits** — time-bounded, zone-scoped, inspectable, revocable contracts that gate specific activities. A film permit allows blocking a street on Tuesday 6am-2pm. The permit is scoped (this block), mortal (expires at 2pm), inspectable (public record), and overridable (police can revoke for emergency).

**What we adopted:**
- The governance contract pattern (explicit grant for specific activity in specific zone)
- Time-bounded with hard expiry (permits don't last forever)
- Public inspectability (anyone can see active permits)
- Authority override (equivalent to operator dismiss)
- Zone-scoped (physical location → named workspace zone)

**What we adapted:**
- Human approval workflow → automatic issuance with expiry
- Paper records → structured digital claims with schema
- Fixed zones → hierarchical, dynamically-created zones

**What we rejected:**
- Permanent zoning (all enforcement is temporary by design)
- Complex bureaucratic process (claims are immediate, not request-then-wait)
- Appeal mechanisms (override is immediate, not process-heavy)

**Key references:**
- Municipal permit systems (general pattern, not specific citation)
- Ostrom, E. (1990). _Governing the Commons_. (Principles for managing shared resources)
- Sassen, S. (2001). _The Global City_. (Urban coordination without central planning)

---

## Synthesis

These four traditions converge on a core insight: **coordination of autonomous agents in shared environments requires both advisory and enforcement mechanisms, both must be time-bounded, and both must be transparent to operators.**

No single tradition provides the complete picture:
- Stigmergy gives us indirect coordination but no enforcement
- Distributed leases give us enforcement but no advisory signaling
- Resilience engineering gives us self-healing but no inter-agent coordination
- Urban governance gives us the contract model but no programmatic substrate

The Agent Coordination Substrate synthesizes all four into one coherent layer.

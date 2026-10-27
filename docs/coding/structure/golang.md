# Go Project Structure

Organize by **business capability**, with pragmatic dependency inversion—not mandatory Clean/Hexagonal layers. Packages and interfaces exist only when they express a useful boundary.

## Example: multi-tenant commerce

```text
cmd/                      # Executable entry points (api, worker)
internal/
  app/                    # Composition root: wire dependencies, startup/shutdown
  domain/                 # Business concepts and domain-specific operations
    organization/         # Organizations, memberships, roles
    user/                 # Global identities; users may join several organizations
    product/              # Catalog, variants, prices
    inventory/            # Stock and reservations
    promotion/            # Coupons, eligibility, discounts
    cart/                 # Mutable purchase intent
    order/                # Durable purchase and price snapshots
    payment/              # Payment attempts, captures, refunds
  service/                # Cross-domain orchestration
    checkout/
    ordercancellation/
    paymentprocessing/
  infrastructure/         # Concrete adapters: postgres, payment clients, messaging, config
  transport/              # Incoming entry points: http, grpc, cli, workers
  support/                # Small, business-free reusable utilities
  testutil/               # Shared test fixtures, stubs, test containers
migrations/
```

## Rules and reasons

1. **Prefer cohesive, capability-based packages.** Keep a domain's models, constructors, invariants, model methods, domain services and domain-specific application operations together; do not mandate `model/service/command/query` subfolders. Split packages only when responsibilities or size justify it.
2. **Separate model behavior from orchestration.** `Ticket.Purchase()` enforces entity invariants; `ticket.Purchaser` coordinates domain-specific operations, possibly using repositories. In strict DDD, I/O orchestration is an *application service*, even when colocated in the domain package. Keep models themselves free of I/O and persistence concerns.
3. **Reserve `service/` for workflows spanning domains or lacking a natural single-domain owner.** E.g., checkout coordinates cart, product, promotion, inventory and order. It is not a dumping ground for logging, configuration or every use case.
4. **Accept interfaces for external behavior, return concrete structs.** `domain` and `service` define small interfaces for dependencies such as repositories, gateways and transactions; infrastructure implements them. Constructors such as `NewPurchaser(repo Repository) *Purchaser` expose concrete services. Business inputs/results and entities remain concrete types—no interface per value, service or operation.
5. **Transport may depend on concrete application services.** HTTP/gRPC/CLI/worker handlers translate requests and errors; they can accept `*checkout.Service` directly, without redundant `UseCase` interfaces. They should not implement business policy or wire repositories themselves.
6. **Keep imports directional and acyclic.** `domain` never imports `infrastructure`, `transport` or `app`; `service` may import `domain` but not concrete infrastructure/transport; `infrastructure` may import domain/service contracts; `transport` may import domain/service; `app` imports implementations and assembles them. Domain-to-domain imports must preserve clear ownership and avoid cycles.
7. **Keep infrastructure at the edges.** Database implementations, external HTTP/gRPC clients, auth providers and configuration libraries live in `infrastructure/`. A domain-specific repository contract belongs with its consumer, not in a global `ports/` package.
8. **Enforce tenant isolation beyond routing.** `organization` owns memberships and business rules, not all tenant enforcement. Derive authorized tenant identity from authentication, pass tenant IDs explicitly to operations, scope persistence queries and constraints, and consider database row-level security as defense in depth.
9. **Make cross-domain consistency explicit.** Checkout recalculates authoritative prices, uses concurrency-safe inventory reservation, durable idempotency and an atomic database transaction for order/reservation/outbox writes. External payments run outside that transaction and require retry-safe processing, verified webhooks and reconciliation.
10. **Use shared packages sparingly.** `support/` is for genuinely reusable business-independent code (prefer Go's standard library first). `testutil/` is for tests only; production code must not import it.
11. **Build only what exists.** No mandatory service/repository/interface for every domain, no giant `common` package, no speculative subpackages, no global dependency container passed everywhere. Add abstractions when they solve real coupling or testing needs.

**Example dependency:** `transport/http → service/checkout → domain/{cart,product,promotion,inventory,order}`; `infrastructure/postgres` implements consumer-defined persistence contracts; `app` constructs and connects everything. Arrows indicate **Go imports**, not runtime calls.

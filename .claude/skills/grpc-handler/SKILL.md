---
name: grpc-handler
description: Create a new GRPC (Connect RPC) handler for the Order Service following established project patterns. Covers handler file, server registration, store layer, converter, and mock updates.
user_invocable: true
---

# Create GRPC Handler

You are creating a new Connect RPC handler for the Order Service. Follow every step below precisely — the project has strict conventions.

## Step 0: Gather Information

Before writing any code, determine:
1. **RPC method name** (e.g., `UpdateOrder`, `DeleteOrder`, `CancelOrder`)
2. **Proto request/response types** — check generated code in the contracts module
3. **Required store operations** — does the handler need a new store method?
4. **Business logic** — what validation rules and error scenarios apply?

If the user hasn't specified these, ask before proceeding.

## Step 1: Create Handler File

**File**: `internal/domain/orders/grpc_{method_name_snake}_handler.go`

Use this exact structure (example for `UpdateOrder`):

```go
package orders

import (
	"context"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/store"
)

type updateOrderHandler struct {
	store store.OrderStore
}

func newUpdateOrderHandler(store store.OrderStore) *updateOrderHandler {
	return &updateOrderHandler{store: store}
}

func (h *updateOrderHandler) Handle(
	ctx context.Context,
	req *connect.Request[orderv1.UpdateOrderRequest],
) (*connect.Response[orderv1.UpdateOrderResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}

	// Business logic + store call here

	return connect.NewResponse(&orderv1.UpdateOrderResponse{
		// response fields
	}), nil
}

func (h *updateOrderHandler) validate(req *orderv1.UpdateOrderRequest) error {
	// Validate required fields
	if req.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	return nil
}
```

### Handler rules:
- **One handler per file**, named `grpc_{snake_case_method}_handler.go`
- Struct is **private** with a single `store store.OrderStore` field
- Constructor: `newXxxHandler(store store.OrderStore) *xxxHandler`
- Public method: `Handle(ctx, req) (resp, error)`
- Private method: `validate(req) error`
- Handle method flow: **validate → business logic → store call → convert → response**
- Never log inside handlers — errors propagate via Connect RPC codes

## Step 2: Register in server.go

Edit `internal/domain/orders/server.go`:

1. Add handler field to `Server` struct:
```go
type Server struct {
	// ... existing handlers ...
	updateOrderHandler *updateOrderHandler
}
```

2. Initialize in `NewServer()`:
```go
func NewServer(store store.OrderStore) *Server {
	return &Server{
		// ... existing handlers ...
		updateOrderHandler: newUpdateOrderHandler(store),
	}
}
```

3. Add delegating method:
```go
func (s *Server) UpdateOrder(
	ctx context.Context,
	req *connect.Request[orderv1.UpdateOrderRequest],
) (*connect.Response[orderv1.UpdateOrderResponse], error) {
	return s.updateOrderHandler.Handle(ctx, req)
}
```

The delegating method signature must match the Connect RPC service interface exactly.

## Step 3: Store Layer (if needed)

If the handler requires a new store operation:

### 3a. Add to interface (`internal/store/order.go`)

```go
type OrderStore interface {
	Create(ctx context.Context, order *entity.Order) error
	Get(ctx context.Context, id string) (*entity.Order, error)
	List(ctx context.Context) ([]*entity.Order, error)
	// Add new method:
	Update(ctx context.Context, order *entity.Order) error
	Close() error
}
```

### 3b. Implement in `PostgresStore`

```go
func (s *PostgresStore) Update(ctx context.Context, order *entity.Order) error {
	const query = `UPDATE orders SET item = :item, amount = :amount, status = :status WHERE id = :id`
	result, err := s.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderNotFound
	}
	return nil
}
```

### 3c. Add to `MockOrderStore` (`internal/store/mock_order_store.go`)

```go
type MockOrderStore struct {
	// ... existing fields ...
	UpdateFunc func(ctx context.Context, order *entity.Order) error
}

func (m *MockOrderStore) Update(ctx context.Context, order *entity.Order) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, order)
	}
	return nil
}
```

## Step 4: Converter (if needed)

Existing converter in `internal/domain/orders/converter.go`:
- `entityToProto(e *entity.Order) *orderv1.Order` — already exists, reuse it

If you need a new conversion direction (e.g., proto → entity), add it to the same file:

```go
func protoToEntity(p *orderv1.Order) *entity.Order {
	// conversion logic
}
```

## Step 5: Error Mapping

Use these Connect RPC error codes consistently:

| Scenario | Code | Example |
|----------|------|---------|
| Invalid/missing input fields | `connect.CodeInvalidArgument` | `connect.NewError(connect.CodeInvalidArgument, nil)` |
| Entity not found | `connect.CodeNotFound` | `connect.NewError(connect.CodeNotFound, err)` |
| Authorization/ownership failure | `connect.CodePermissionDenied` | `connect.NewError(connect.CodePermissionDenied, errors.New("..."))` |
| Unexpected/internal errors | `connect.CodeInternal` | `connect.NewError(connect.CodeInternal, err)` |

For NotFound, always check with `errors.Is`:
```go
if errors.Is(err, store.ErrOrderNotFound) {
    return nil, connect.NewError(connect.CodeNotFound, err)
}
return nil, connect.NewError(connect.CodeInternal, err)
```

## Step 6: Verify

After creating the handler:
1. Run `task build` to ensure compilation
2. Run `task lint` to check for lint errors
3. Run `task test` to verify existing tests still pass

## Reference Files

Study these files to understand existing patterns:

| File | Purpose |
|------|---------|
| `internal/domain/orders/grpc_create_order_handler.go` | Create with validation, UUID generation, store.Create |
| `internal/domain/orders/grpc_get_order_handler.go` | Read with NotFound error handling |
| `internal/domain/orders/grpc_check_order_owner_handler.go` | Business logic + PermissionDenied |
| `internal/domain/orders/grpc_list_orders_handler.go` | List with slice conversion |
| `internal/domain/orders/server.go` | Handler registration and delegation |
| `internal/domain/orders/converter.go` | Entity ↔ Proto conversion |
| `internal/store/order.go` | OrderStore interface + PostgresStore |
| `internal/store/mock_order_store.go` | Manual mock for testing |
| `internal/entity/order.go` | Entity with `db:` tags |

## After Handler is Created

Suggest to the user that they write tests using the `order-test-coordinator` agent, which will create both unit and isolation tests in parallel.

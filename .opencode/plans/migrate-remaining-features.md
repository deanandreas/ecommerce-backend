# Migrate remaining features (cart, order, payment, review) and retire DBService

## Goal
Finish the clean-architecture migration: move the last four features out of
`internal/server` into per-feature packages, then delete the temporary
`DBService` interface entirely.

## Status
- Branch: `feat/minio-storage` (ahead of `main`; product/auth/user/home/system/middleware migrated).
- tx logic already lives in `internal/database` behind `*pSQL` methods
  (`user_tx.go` holds cart tx funcs, `order_tx.go`, `payment_tx.go`,
  `product_tx.go` holds `InsertReviewTx`). No DB reshuffling needed.
- `internal/server` currently holds the 4 feature handler files + `DBService`
  in `server.go` + `Handler()` factory + JSON proxy helpers.

## Commit plan (user-approved)
One commit per feature, then a final cleanup commit:
1. cart
2. order (+ fix `InserOrderTx` -> `InsertOrderTx` typo)
3. payment
4. review
5. delete `DBService` + server handler files + `errors.go`; wire a `Closer`
   for shutdown; resolve `user.DeleteUser` leftover

## Pattern (match existing feature packages: user, product)
- `internal/<feat>/repository.go` — `Repository` interface over `db.*` types + `database.*` input structs (implemented by `*pSQL`).
- `internal/<feat>/service.go` — `Service{ repo }`, `NewService(repo)`, business/validation.
- `internal/<feat>/handler.go` — `Handler{ service }`, `NewHandler`, uses `httpx.Write`/`httpx.Read`/`middleware.GetUserID`.
- Wire in `handlers.go` (add field + factory) and `route.go` (repoint `s.<Feat>.*`).
- Keep API responses/messages identical to current handlers.

## Feature 1 — cart (`internal/cart/`, 5 handlers)
Repository: `InsertCartTx(database.CreateCart)`, `GetUserCart(id)`, `UpdateCartItemTx(UpdateCartItemQuantityParams)`, `DeleteCartItem`, `DeleteAllCartItem`.
Service sentinel: `ErrInvalidInput`.
- `CreateCart` validates `ProductID=="" || Quantity<=0` -> 400 "product id and quantity are required"
- `UpdateCartQuantity` validates `Quantity<=0 || ID==""` -> 400 (message preserved)
- `GetUserCarts` keeps pg `NoData` -> 200 zero-value behavior
- `DeleteCart`/`DeleteCarts` keep path-id checks in handler, `rows==0` -> 400
- Handler keeps pgconn `pgerrcode` mapping (`ForeignKeyViolation`, `InvalidTextRepresentation`, `NoData`) and `database.ErrInvalidQuantity`/`ErrNotUpdated` mapping exactly as today.
Routes -> `s.Cart.*`; delete `internal/server/cart.go`.

## Feature 2 — order (`internal/order/`, 3 handlers)
Repository: `InserOrderTx` -> rename to `InsertOrderTx` in BOTH `DBService` and `pSQL` (`internal/database/order_tx.go`), `GetAllUserOrders`, `GetUserOrder`, `GetUnpayedOrder`.
- `CreateOrder` keeps ship-date shipping-price calculation + pg `NoData`/`pgx.ErrNoRows` -> 400 "user does not have cart"
- `GetAllUserOrders` nil -> `[]{}`
Routes -> `s.Order.*`; delete `internal/server/order.go`.

## Feature 3 — payment (`internal/payment/`, 4 handlers)
Repository: `GetUnpayedOrder`, `InsertPayment`, `GetPaymentByID`, `ConfirmPaymentTx`, `CancelPaymentTx`.
- Move `paymentItem` struct + `newTransactionID()` helper into package.
- `InitiatePayment` keeps order/cart-item total calc; `json.Marshal/Unmarshal` of `OrderItems`.
- Ownership checks (`payment.UserID != userID` -> 403) stay in handlers.
- Maps `database.ErrPaymentNotPending` -> 409, `database.ErrNoItem`/`pgx.ErrNoRows` -> 400, `pgx.ErrNoRows` (GetPaymentByID) -> 404.
Routes -> `s.Payment.*`; delete `internal/server/payment.go`.

## Feature 4 — review (`internal/review/`, 3 handlers)
Repository: `InsertReviewTx` (already on `*pSQL` in product_tx.go), `GetUserReview`, `GetProductReviews`.
- `CreateReview` validates `ProductID=="" || Rating==0` -> 400.
- `GetProductReviews` remains a PUBLIC route (no Auth middleware).
Routes -> `s.Review.*`; delete `internal/server/review.go`.

## Feature 5 — retire DBService
- Delete `DBService` interface and dead members (`UpdateAddress`, `DeleteUser`).
- `Handler(dbURL)` no longer needs to return the pool as `DBService`. `pSQL`
  is unexported, so return a minimal `Closer` interface (`Close()`) instead;
  `Server` keeps `db Closer` for `Shutdown`.
- Delete `internal/server/errors.go` (alias vars unused after migration).
- Remove now-unused `WriteJSON`/`ReadJSON`/`FromDataToJSON` proxies from `util.go`.
- `user.DeleteUser` leftover: it is unwired from DBService already; decide to
  route `DELETE /api/v1/users/{id}` to `s.User.DeleteUser` OR delete the dead
  handler/method. (Recommend: wire the route — keeps existing handler working.)

## Leftovers folded in
- `InserOrderTx` -> `InsertOrderTx` (server + database).
- `GetUnpayedOrder` spelling kept (avoid churn).
- `user.DeleteUser` handled in Feature 5.

## Verification (each commit)
`go build ./... && go vet ./... && gofmt -l internal/ cmd/ && go test ./...`
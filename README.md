# E-Commerce API

REST API backend for an e-commerce platform — built with Go and PostgreSQL.

## Tech Stack

- **Language:** Go 1.26.5
- **Database:** PostgreSQL with connection pooling ([pgx](https://github.com/jackc/pgx))
- **Query builder:** [sqlc](https://sqlc.dev) — type-safe SQL from hand-written queries
- **Auth:** JWT access & refresh tokens ([golang-jwt](https://github.com/golang-jwt/jwt))
- **Password hashing:** bcrypt ([golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto/bcrypt))
- **Image processing:** [imaging](https://github.com/disintegration/imaging) (Lanczos resampling, 300px)
- **Dev server:** [air](https://github.com/air-verse/air) for live reload (via [mise](https://mise.jdx.dev))

## Project Structure

```
cmd/api/main.go                 # Entry point
internal/
  server/
    server.go                   # Server setup & graceful shutdown
    route.go                    # All HTTP routes & middleware wiring
    middleware.go                # JWT auth, CORS, request logging
    auth.go                     # Register, login, refresh token handlers
    auth_token.go               # JWT creation & validation helpers
    home.go                     # Home page, products, categories
    product.go                  # CRUD for user products
    cart.go                     # Cart operations
    order.go                    # Order placement & lookup
    payment.go                  # Payment initiation, confirm, cancel
    review.go                   # Product review CRUD
    user.go                     # User profile & address management
    errors.go                   # Shared error types
    util.go                     # JSON helpers, env loading
  database/
    database.go                 # Connection pool & health check
    errors.go                   # Postgres error handling
    util.go                     # File uploads & image resizing
    *_tx.go                     # Transaction logic (users, products, payments, etc.)
    migrations/                 # 14 SQL migrations (up & down)
    queries/                    # sqlc input: SQL queries by domain
    sqlc/                       # sqlc output: generated Go code
```

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL
- [mise](https://mise.jdx.dev) (for air dev server) — optional

### Environment Variables

Create a `.env` file in the project root:

```env
PORT=8080
DBURL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
JWT_KEY=your-256-bit-secret-here
```

### Run

```bash
# with air (live reload)
mise install && mise run air

# or directly
go run cmd/api/main.go
```

### Database Migrations

Migrations are in `internal/database/migrations/`. Apply them using your preferred tool (golang-migrate, goose, etc.).

## API Endpoints

All routes are prefixed with `/api/v1/`. Responses follow a standard shape:

```json
{
  "data": {},
  "message": "human-readable message",
  "success": true
}
```

### Auth (public)

| Method | Endpoint                  | Description            |
| ------ | ------------------------- | ---------------------- |
| POST   | `/register`               | Create new account     |
| POST   | `/login`                  | Login, receive tokens  |
| GET    | `/token/refresh`          | Get new access token   |

### Home (public)

| Method | Endpoint                         | Description                    |
| ------ | -------------------------------- | ------------------------------ |
| GET    | `/products`                      | List products (paginated)      |
| GET    | `/product/categories`            | List product categories        |
| GET    | `/product/star`                  | Popular / top-rated products   |
| GET    | `/product/details/{id}`          | Single product detail view     |
| GET    | `/product/review/{id}`           | Reviews for a product          |

### User Profile (authenticated)

| Method | Endpoint                  | Description              |
| ------ | ------------------------- | ------------------------ |
| GET    | `/user/profiles`          | Get user profile         |
| PATCH  | `/user/profiles`          | Update profile           |
| POST   | `/user/address`           | Add address              |
| PATCH  | `/user/address/{id}`      | Set default address      |
| DELETE | `/user/address/{id}`      | Delete address           |

### Product Management (authenticated)

| Method | Endpoint                         | Description                |
| ------ | -------------------------------- | -------------------------- |
| GET    | `/user/products`                 | List user's products       |
| GET    | `/user/product/{id}`             | Get one of user's products |
| POST   | `/user/products`                 | Create product (+ images)  |
| PATCH  | `/user/products/{id}`            | Update product             |
| DELETE | `/user/products/{id}`            | Soft-delete product        |
| POST   | `/user/product/images/{id}`      | Add images to product      |
| PATCH  | `/user/product/images`           | Set default product image  |
| DELETE | `/user/product/images`           | Delete product image       |

### Cart (authenticated)

| Method | Endpoint                      | Description       |
| ------ | ----------------------------- | ----------------- |
| POST   | `/user/carts`                 | Add item to cart  |
| GET    | `/user/cart/items`            | List cart items   |
| PATCH  | `/user/cart/items`            | Update quantity   |
| DELETE | `/user/cart/items/{id}`       | Remove cart item  |
| DELETE | `/user/carts/{id}`            | Clear entire cart |

### Orders (authenticated)

| Method | Endpoint                      | Description          |
| ------ | ----------------------------- | -------------------- |
| POST   | `/user/orders`                | Place an order       |
| GET    | `/user/orders`                | List all user orders |
| GET    | `/user/orders/{id}`           | Get single order     |

### Payments (authenticated)

| Method | Endpoint                              | Description           |
| ------ | ------------------------------------- | --------------------- |
| POST   | `/user/orders/{id}/payment`           | Initiate payment      |
| POST   | `/user/payments/{id}/confirm`         | Confirm payment       |
| POST   | `/user/payments/{id}/cancel`          | Cancel payment        |
| GET    | `/user/payments/{id}`                 | Get payment details   |

### Reviews (authenticated)

| Method | Endpoint                      | Description       |
| ------ | ----------------------------- | ----------------- |
| POST   | `/user/reviews`               | Create a review   |
| GET    | `/user/review/{id}`           | Get user's review |

### System

| Method | Endpoint          | Description                        |
| ------ | ----------------- | ---------------------------------- |
| GET    | `/health`         | Database health check              |
| GET    | `/`               | Home / welcome message             |

## License

Copyright (C) 2026 Dean Andreas

This project is licensed under the **GNU General Public License v3.0** (GPL-3.0).
See the [LICENSE](LICENSE) file for details.

# 🔐 Authentication Engine (`internal/auth`)

`internal/auth` provides password security and JWT token management for multi-tenant authentication in **PingGopher**.

---

## 📦 Key Exports

### Password Hashing
- **`HashPassword(password string) (string, error)`**: Hashes plaintext passwords using Bcrypt (`bcrypt.DefaultCost`).
- **`CheckPasswordHash(password, hash string) bool`**: Validates a plaintext password against a stored Bcrypt hash.

### JWT Token Management
- **`GenerateJWTToken(userID uuid.UUID, email string, secret string, duration time.Duration) (string, error)`**: Generates signed HMAC-SHA256 JWT tokens containing `UserID` and `Email` claims.
- **`ValidateJWTToken(tokenString string, secret string) (*JWTClaims, error)`**: Parses and validates incoming JWT Bearer tokens.

---

## 🧪 Unit Testing

```bash
go test -v ./internal/auth
```

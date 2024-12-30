# Deepseek 2.5 (Premium)

Here’s a comprehensive guide to creating an authentication service in Go for a web application, covering all the requested aspects with examples and best practices using **JWT**, **bcrypt**, and **PostgreSQL**:

---

### **1. Project Setup**
#### Directory Structure
```
auth-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   ├── services/
│   └── utils/
├── migrations/
├── config/
├── pkg/
├── go.mod
└── go.sum
```

#### Dependencies
Add the following dependencies to your `go.mod` file:
```bash
go get github.com/gin-gonic/gin
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
```

#### Configuration
Create a `config/config.go` file to manage environment variables:
```go
package config

import (
	"os"
)

type Config struct {
	DBConnString string
	JWTSecret    string
}

func LoadConfig() *Config {
	return &Config{
		DBConnString: os.Getenv("DB_CONN_STRING"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
	}
}
```

---

### **2. User Management**
#### User Model
Create a `models/user.go` file:
```go
package models

type User struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}
```

#### Password Hashing with bcrypt
Create a utility function in `utils/password.go`:
```go
package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

---

### **3. Database Operations**
#### PostgreSQL Setup
Create a `repository/user_repository.go` file:
```go
package repository

import (
	"database/sql"
	"auth-service/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(email, password string) error {
	_, err := r.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, password)
	return err
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow("SELECT id, email, password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

---

### **4. API Endpoints**
#### Using Gin Framework
Create a `handlers/auth_handler.go` file:
```go
package handlers

import (
	"net/http"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.userRepo.CreateUser(user.Email, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbUser, err := h.userRepo.FindUserByEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(user.Password, dbUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbUser.ID,
		"email":   dbUser.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
```

---

### **5. Middleware for Authentication**
Create a `middleware/auth_middleware.go` file:
```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"auth-service/config"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}
```

---

### **6. Scaling Strategies**
- **Database Indexing**: Add indexes to frequently queried columns (e.g., `email`).
- **Caching**: Use Redis to cache user sessions or frequently accessed data.
- **Load Balancing**: Deploy multiple instances of the service behind a load balancer.
- **Connection Pooling**: Use `sqlx` or `pgx` for efficient database connection management.

---

### **7. Integration with Web Frameworks**
#### Example with Gin
```go
package main

import (
	"auth-service/config"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/repository"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open("postgres", cfg.DBConnString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)

	r := gin.Default()

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	authGroup := r.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "You are authenticated"})
		})
	}

	r.Run(":8080")
}
```

---

### **8. Best Practices**
- Use HTTPS for secure communication.
- Validate and sanitize all user inputs.
- Implement rate limiting to prevent brute-force attacks.
- Use environment variables for sensitive data (e.g., JWT secret, DB credentials).
- Regularly update dependencies to patch security vulnerabilities.

---

### **9. Testing**

#### Unit Tests

Create a `handlers/auth_handler_test.go` file:
```go
package handlers

import (
	"auth-service/models"
	"auth-service/repository"
	"auth-service/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(email, password string) error {
	args := m.Called(email, password)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("CreateUser", "test@example.com", mock.Anything).Return(nil)

	handler := NewAuthHandler(mockRepo, "secret")

	router := gin.Default()
	router.POST("/register", handler.Register)

	payload := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}
```

#### Integration Tests
Create a `repository/user_repository_test.go` file:
```go
package repository

import (
	"auth-service/models"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://user:password@localhost/auth_test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	err = repo.CreateUser("test@example.com", "hashedpassword")
	assert.NoError(t, err)
}
```

---

### **10. Deployment**
#### Dockerfile
Create a `Dockerfile`:
```dockerfile
FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd/main.go

EXPOSE 8080

CMD ["./auth-service"]
```

#### Docker Compose
Create a `docker-compose.yml` file:
```yaml
version: '3.8'

services:
  auth-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_CONN_STRING: "postgres://user:password@db:5432/auth_db?sslmode=disable"
      JWT_SECRET: "supersecretkey"
    depends_on:
      - db

  db:
    image: postgres:13
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: auth_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

### **11. Monitoring and Logging**
#### Logging with Zap
Add the `zap` logger:
```bash
go get go.uber.org/zap
```

Update `main.go`:
```go
package main

import (
	"auth-service/config"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/repository"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.LoadConfig()

	db, err := sql.Open("postgres", cfg.DBConnString)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)

	r := gin.Default()

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	authGroup := r.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "You are authenticated"})
		})
	}

	logger.Info("Starting server on :8080")
	r.Run(":8080")
}
```

---

### **12. Common Pitfalls**
1. **Weak Passwords**: Enforce strong password policies.
2. **Insecure JWT Storage**: Store JWTs securely (e.g., HttpOnly cookies).
3. **Lack of Input Validation**: Always validate and sanitize user inputs.
4. **No Rate Limiting**: Implement rate limiting to prevent abuse.
5. **Hardcoded Secrets**: Use environment variables for sensitive data.

---

### **13. Extensions**
- Add **OAuth2** support for third-party logins.
- Implement **two-factor authentication (2FA)**.
- Use **gRPC** for internal service communication.
- Add **API documentation** using Swagger.

---

### **14. Example Code Repository**

To make it easier to follow along, here’s a sample repository structure with all the code snippets provided above:

```
auth-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   │   └── auth_handler.go
│   ├── middleware/
│   │   └── auth_middleware.go
│   ├── models/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   ├── services/
│   ├── utils/
│   │   └── password.go
│   └── config/
│       └── config.go
├── migrations/
│   └── 001_create_users_table.sql
├── pkg/
├── go.mod
├── go.sum
├── Dockerfile
└── docker-compose.yml
```

---

### **15. Database Migrations**
Create a `migrations/001_create_users_table.sql` file to set up the `users` table:
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);
```

Run migrations using a tool like [golang-migrate](https://github.com/golang-migrate/migrate):
```bash
migrate -path ./migrations -database "postgres://user:password@localhost/auth_db?sslmode=disable" up
```

---

### **16. Example API Requests**
#### Register User
```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{"email": "test@example.com", "password": "password123"}'
```

#### Login User
```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"email": "test@example.com", "password": "password123"}'
```

#### Access Protected Route
```bash
curl -X GET http://localhost:8080/api/protected \
-H "Authorization: <JWT_TOKEN>"
```

---

### **17. Production Considerations**
1. **HTTPS**: Use a reverse proxy like Nginx or Caddy to enable HTTPS.
2. **Logging**: Centralize logs using tools like ELK Stack or Grafana Loki.
3. **Monitoring**: Use Prometheus and Grafana for monitoring.
4. **Secrets Management**: Use tools like HashiCorp Vault or AWS Secrets Manager.
5. **CI/CD**: Set up a CI/CD pipeline using GitHub Actions, GitLab CI, or Jenkins.

---

### **18. Conclusion**
This guide provides a complete implementation of an authentication service in Go using **JWT**, **bcrypt**, and **PostgreSQL**. It covers everything from project setup to deployment and scaling strategies. By following best practices and incorporating security measures, you can build a robust and scalable authentication service for your web application.

Feel free to extend this service with additional features like **OAuth2**, **2FA**, or **role-based access control (RBAC)** based on your application’s requirements. Happy coding! 🚀



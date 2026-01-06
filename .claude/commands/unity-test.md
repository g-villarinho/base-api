# TASK

Write unit tests for the provided file.

**Required parameter:** file (the file to write tests for)
**Optional parameter:** method name (specific method to test)

Usage:
- `/unity-test <file>` - Write tests for all methods in the file
- `/unity-test <file> <method_name>` - Write tests only for the specified method

If a specific method is provided, focus only on that method. Otherwise, write tests for all public methods in the file.

---

## Structure and Naming

### File Naming
- For a service `auth.go` → test file: `auth_test.go`
- For a repository `user_repository.go` → test file: `user_repository_test.go`
- For a handler `auth_handler.go` → test file: `auth_handler_test.go`

### Test Method Structure
- **One test method for each service/repository method**
- Method name: `Test[MethodName]`
- Example: `TestLogin`, `TestAuthenticate`, `TestSendVerificationCode`

### Using `t.Run()` for Subtests
- Inside each test method, use `t.Run()` with descriptive descriptions
- Format: `t.Run("should [description of expected behavior]", func(t *testing.T) { ... })`
- Examples:
  - `t.Run("should return success when valid credentials are provided", func(t *testing.T) { ... })`
  - `t.Run("should return error when email is invalid", func(t *testing.T) { ... })`
  - `t.Run("should return error when OTP code is expired", func(t *testing.T) { ... })`

## Parallel Test Execution

### Rules for `t.Parallel()`

1. **Always add `t.Parallel()` as the first line of each test function**
2. **Always add `t.Parallel()` as the first line of each subtest (`t.Run`)**
3. This enables tests to run concurrently, improving test execution speed

```go
func TestLogin(t *testing.T) {
    t.Parallel() // First line of test function

    t.Run("should return success", func(t *testing.T) {
        t.Parallel() // First line of each subtest

        // test code...
    })
}
```

## Test Structure Pattern

```go
func TestLogin(t *testing.T) {
    t.Parallel()

    t.Run("should return success when valid credentials are provided", func(t *testing.T) {
        t.Parallel()

        // Arrange
        service, deps := setupAuthService(t)
        // Setup test data...

        // Act
        result, err := service.Login(ctx, ...)

        // Assert
        require.NoError(t, err)
        assert.NotNil(t, result)
    })

    t.Run("should return error when email is invalid", func(t *testing.T) {
        t.Parallel()

        // Arrange
        service, deps := setupAuthService(t)
        // ...

        // Act
        // Assert
    })
}
```

## Dependencies Struct Pattern

### Purpose
Use a dedicated struct to hold all mock dependencies for a component's tests. This makes it easy to access any mock in your tests and keeps the setup function signature clean.

### Naming Convention
- Format: `[component]TestDeps`
- Examples: `authTestDeps`, `userTestDeps`, `sessionTestDeps`

### Structure Definition

```go
// authTestDeps holds all mock dependencies for auth service tests.
type authTestDeps struct {
    store               *mocks.StoreMock
    verificationService *mocks.VerificationServiceMock
    sessionService      *mocks.SessionServiceMock
    emailService        *mocks.EmailServiceMock
    captchaService      *mocks.CaptchaServiceMock
}
```

## Setup Helpers with Dependencies Struct

### Rules for Setup Helpers

1. **Always use `t.Helper()` as the first line**
2. **Return the service/component AND the deps struct**
3. **Use `mocks.New[Interface]Mock(t)` to create mocks**

### Example: Service Setup Helper

```go
func setupAuthService(t *testing.T) (AuthService, *authTestDeps) {
    t.Helper()

    deps := &authTestDeps{
        store:               mocks.NewStoreMock(t),
        verificationService: mocks.NewVerificationServiceMock(t),
        sessionService:      mocks.NewSessionServiceMock(t),
        emailService:        mocks.NewEmailServiceMock(t),
        captchaService:      mocks.NewCaptchaServiceMock(t),
    }

    service := NewAuthService(
        deps.store,
        deps.verificationService,
        deps.sessionService,
        deps.emailService,
        deps.captchaService,
    )

    return service, deps
}
```

### Example: Service with Configuration

```go
// sessionTestDeps holds all mock dependencies for session service tests.
type sessionTestDeps struct {
    store  *mocks.StoreMock
    signer *mocks.SignerMock
}

func setupSessionService(t *testing.T) (SessionService, *sessionTestDeps) {
    t.Helper()

    deps := &sessionTestDeps{
        store:  mocks.NewStoreMock(t),
        signer: mocks.NewSignerMock(t),
    }

    cfg := &config.Config{
        Session: config.Session{
            Duration: 24 * time.Hour,
        },
    }

    service := NewSessionService(deps.store, deps.signer, cfg)

    return service, deps
}
```

## Helper Functions for Test Data

### Purpose
Create helper functions to build common test data structures. This reduces boilerplate and makes tests more readable.

### Naming Convention
- Format: `createTest[EntityName]`
- Examples: `createTestUser`, `createTestSQLCUser`, `createTestSession`

### Example: SQLC User Helper

```go
func createTestSQLCUser(id uuid.UUID, name, email string, status domain.UserStatus, emailConfirmedAt *time.Time) sqlc.User {
    var emailConfirmedAtNull sql.NullTime
    if emailConfirmedAt != nil {
        emailConfirmedAtNull = sql.NullTime{Time: *emailConfirmedAt, Valid: true}
    }

    return sqlc.User{
        ID:               id,
        Name:             sql.NullString{String: name, Valid: true},
        Email:            email,
        Status:           string(status),
        CreatedAt:        time.Now().UTC(),
        EmailConfirmedAt: emailConfirmedAtNull,
    }
}
```

### Example: Domain Entity Helper

```go
func createTestVerification(userID uuid.UUID, flow domain.VerificationFlow) *domain.Verification {
    return &domain.Verification{
        ID:        uuid.New(),
        Token:     "verification-token",
        UserID:    userID,
        Flow:      flow,
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }
}
```

## Mock Expectations with EXPECT()

### Using EXPECT() Pattern
- Use `EXPECT()` method for setting up mock expectations
- Chain method calls for readable expectations
- Use `mock.AnythingOfType()` for dynamic values

```go
deps.captchaService.EXPECT().
    Validate(ctx, captchaToken, remoteIP).
    Return(nil)

deps.store.EXPECT().
    FindUserByEmail(ctx, email).
    Return(sqlc.User{}, sql.ErrNoRows)

deps.store.EXPECT().
    CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
    Return(nil)
```

## Complete Example

```go
// authTestDeps holds all mock dependencies for auth service tests.
type authTestDeps struct {
    store               *mocks.StoreMock
    verificationService *mocks.VerificationServiceMock
    sessionService      *mocks.SessionServiceMock
    emailService        *mocks.EmailServiceMock
    captchaService      *mocks.CaptchaServiceMock
}

func setupAuthService(t *testing.T) (AuthService, *authTestDeps) {
    t.Helper()

    deps := &authTestDeps{
        store:               mocks.NewStoreMock(t),
        verificationService: mocks.NewVerificationServiceMock(t),
        sessionService:      mocks.NewSessionServiceMock(t),
        emailService:        mocks.NewEmailServiceMock(t),
        captchaService:      mocks.NewCaptchaServiceMock(t),
    }

    service := NewAuthService(
        deps.store,
        deps.verificationService,
        deps.sessionService,
        deps.emailService,
        deps.captchaService,
    )

    return service, deps
}

func createTestSQLCUser(id uuid.UUID, name, email string, status domain.UserStatus, emailConfirmedAt *time.Time) sqlc.User {
    var emailConfirmedAtNull sql.NullTime
    if emailConfirmedAt != nil {
        emailConfirmedAtNull = sql.NullTime{Time: *emailConfirmedAt, Valid: true}
    }

    return sqlc.User{
        ID:               id,
        Name:             sql.NullString{String: name, Valid: true},
        Email:            email,
        Status:           string(status),
        CreatedAt:        time.Now().UTC(),
        EmailConfirmedAt: emailConfirmedAtNull,
    }
}

func TestRegisterAccount(t *testing.T) {
    t.Parallel()

    t.Run("should return error when captcha validation fails", func(t *testing.T) {
        t.Parallel()

        // Arrange
        service, deps := setupAuthService(t)
        ctx := context.Background()
        email := "john@example.com"
        captchaToken := "invalid-token"
        remoteIP := "192.168.1.1"
        captchaErr := errors.New("captcha validation failed")

        deps.captchaService.EXPECT().
            Validate(ctx, captchaToken, remoteIP).
            Return(captchaErr)

        // Act
        err := service.RegisterAccount(ctx, email, captchaToken, remoteIP)

        // Assert
        require.Error(t, err)
        assert.Contains(t, err.Error(), "validate captcha")
    })

    t.Run("should create new user when user does not exist", func(t *testing.T) {
        t.Parallel()

        // Arrange
        service, deps := setupAuthService(t)
        ctx := context.Background()
        email := "john@example.com"
        captchaToken := "valid-token"
        remoteIP := "192.168.1.1"

        verification := &domain.Verification{
            ID:    uuid.New(),
            Token: "verification-token",
            Flow:  domain.VerificationEmailFlow,
        }

        deps.captchaService.EXPECT().
            Validate(ctx, captchaToken, remoteIP).
            Return(nil)

        deps.store.EXPECT().
            FindUserByEmail(ctx, email).
            Return(sqlc.User{}, sql.ErrNoRows)

        deps.store.EXPECT().
            CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
            Return(nil)

        deps.verificationService.EXPECT().
            CreateVerification(ctx, mock.AnythingOfType("uuid.UUID"), domain.VerificationEmailFlow, "").
            Return(verification, nil)

        deps.emailService.EXPECT().
            SendVerifyEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
            Return()

        // Act
        err := service.RegisterAccount(ctx, email, captchaToken, remoteIP)

        // Assert
        require.NoError(t, err)
    })

    t.Run("should return ErrUserBlocked when user is blocked", func(t *testing.T) {
        t.Parallel()

        // Arrange
        service, deps := setupAuthService(t)
        ctx := context.Background()
        email := "blocked@example.com"
        captchaToken := "valid-token"
        remoteIP := "192.168.1.1"
        userID := uuid.New()

        blockedUser := createTestSQLCUser(userID, "Blocked User", email, domain.BlockedStatus, nil)

        deps.captchaService.EXPECT().
            Validate(ctx, captchaToken, remoteIP).
            Return(nil)

        deps.store.EXPECT().
            FindUserByEmail(ctx, email).
            Return(blockedUser, nil)

        // Act
        err := service.Login(ctx, email, captchaToken, remoteIP, "", "")

        // Assert
        require.Error(t, err)
        assert.ErrorIs(t, err, domain.ErrUserBlocked)
    })
}
```

## Additional Best Practices

### 1. Test Organization (AAA Pattern)
- **Arrange**: Prepare data, mocks, and dependencies
- **Act**: Execute the method being tested
- **Assert**: Verify expected results

### 2. Using Mocks
- Use the generated mocks in the `internal/mocks/` folder
- Use `EXPECT()` method for setting expectations
- Use `mock.AnythingOfType()` for dynamic parameter matching

### 3. Descriptive Names
- Use names that describe the scenario and expected result
- Avoid generic names like "test1", "test2"
- Good examples:
  - `"should return user when valid ID is provided"`
  - `"should return error when user not found"`
  - `"should return ErrUserBlocked when user is blocked"`

### 4. Scenario Coverage
- Test success cases
- Test error cases
- Test edge cases (empty values, nulls, etc.)
- Test domain error mappings (e.g., `ErrUserNotFound`, `ErrUserBlocked`)

### 5. Assertions
- Use `assert` and `require` from testify
- Use `require` for conditions that should stop the test if they fail
- Use `assert` for verifications that can continue the test
- Use `assert.ErrorIs()` for domain error comparisons
- Use `assert.Contains()` for error message validation

## Useful Commands

### Running Tests
```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/service/...

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -run TestRegisterAccount ./internal/service/...
```

### Generating Mocks
```bash
# Generate mocks for interfaces
make mocks
```

Follow these guidelines to maintain consistency and quality in project tests.
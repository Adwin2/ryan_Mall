# 🧪 TDD测试体系实现指南

## 📋 目标概述

通过TDD（测试驱动开发）方法为Ryan Mall建立完整的测试体系，包括单元测试、集成测试和性能测试。

**简历亮点**: "建立完整的测试体系，单元测试覆盖率达到85%，集成测试覆盖核心业务流程"

## 🎯 TDD实现计划

### 阶段一：测试环境搭建（第1天）
### 阶段二：Service层单元测试（第2-4天）
### 阶段三：Repository层单元测试（第5-6天）
### 阶段四：Handler层集成测试（第7-9天）
### 阶段五：性能测试和覆盖率报告（第10天）

## 🚀 阶段一：测试环境搭建

### 1.1 安装测试依赖

```bash
# 在项目根目录执行
go mod tidy

# 添加测试相关依赖
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-redis/redismock/v8
```

### 1.2 创建测试目录结构

```bash
mkdir -p tests/{unit,integration,performance}
mkdir -p tests/unit/{service,repository,handler}
mkdir -p tests/integration/{api}
mkdir -p tests/mocks
mkdir -p tests/fixtures
```

### 1.3 创建测试配置文件

**tests/config/test_config.go**:
```go
package config

import (
    "ryan-mall/internal/config"
)

func GetTestConfig() *config.Config {
    return &config.Config{
        Database: config.DatabaseConfig{
            Host:     "localhost",
            Port:     "3306",
            Username: "test_user",
            Password: "test_password",
            Database: "ryan_mall_test",
        },
        JWT: config.JWTConfig{
            SecretKey:   "test_secret_key_for_testing_only",
            ExpireHours: 24,
        },
        Server: config.ServerConfig{
            Port: "8081",
            Mode: "test",
        },
    }
}
```

## 🧪 阶段二：Service层单元测试（TDD实践）

### 2.1 UserService测试实现

**第一步：编写失败的测试**

**tests/unit/service/user_service_test.go**:
```go
package service_test

import (
    "testing"
    "ryan-mall/internal/model"
    "ryan-mall/internal/service"
    "ryan-mall/tests/mocks"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
    suite.Suite
    userService service.UserService
    mockUserRepo *mocks.MockUserRepository
    mockJWTManager *mocks.MockJWTManager
}

func (suite *UserServiceTestSuite) SetupTest() {
    suite.mockUserRepo = &mocks.MockUserRepository{}
    suite.mockJWTManager = &mocks.MockJWTManager{}
    suite.userService = service.NewUserService(suite.mockUserRepo, suite.mockJWTManager)
}

// 测试用户注册成功场景
func (suite *UserServiceTestSuite) TestRegister_Success() {
    // Arrange - 准备测试数据
    req := &model.UserRegisterRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    // Mock期望调用
    suite.mockUserRepo.On("ExistsByUsername", req.Username).Return(false, nil)
    suite.mockUserRepo.On("ExistsByEmail", req.Email).Return(false, nil)
    suite.mockUserRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)
    suite.mockJWTManager.On("GenerateToken", mock.AnythingOfType("uint"), req.Username, req.Email).Return("mock_token", nil)
    
    // Act - 执行测试
    result, err := suite.userService.Register(req)
    
    // Assert - 验证结果
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    assert.Equal(suite.T(), "mock_token", result.Token)
    assert.Equal(suite.T(), req.Username, result.User.Username)
    assert.Equal(suite.T(), req.Email, result.User.Email)
    
    // 验证Mock调用
    suite.mockUserRepo.AssertExpectations(suite.T())
    suite.mockJWTManager.AssertExpectations(suite.T())
}

// 测试用户名已存在的场景
func (suite *UserServiceTestSuite) TestRegister_UsernameExists() {
    // Arrange
    req := &model.UserRegisterRequest{
        Username: "existinguser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    suite.mockUserRepo.On("ExistsByUsername", req.Username).Return(true, nil)
    
    // Act
    result, err := suite.userService.Register(req)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Contains(suite.T(), err.Error(), "用户名已存在")
    
    suite.mockUserRepo.AssertExpectations(suite.T())
}

func TestUserServiceTestSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}
```

**第二步：创建Mock对象**

**tests/mocks/mock_user_repository.go**:
```go
package mocks

import (
    "ryan-mall/internal/model"
    "github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*model.User, error) {
    args := m.Called(id)
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*model.User, error) {
    args := m.Called(username)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
    args := m.Called(id)
    return args.Error(0)
}

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
    args := m.Called(username)
    return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(email string) (bool, error) {
    args := m.Called(email)
    return args.Bool(0), args.Error(1)
}
```

**第三步：运行测试（应该失败）**

```bash
cd tests/unit/service
go test -v ./...
```

**第四步：修复代码使测试通过**

检查UserService实现，确保逻辑正确。

**第五步：重构和优化**

添加更多测试用例，覆盖边界情况。

### 2.2 ProductService测试实现

**tests/unit/service/product_service_test.go**:
```go
package service_test

import (
    "testing"
    "ryan-mall/internal/model"
    "ryan-mall/internal/service"
    "ryan-mall/tests/mocks"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
    suite.Suite
    productService service.ProductService
    mockProductRepo *mocks.MockProductRepository
    mockCategoryRepo *mocks.MockCategoryRepository
}

func (suite *ProductServiceTestSuite) SetupTest() {
    suite.mockProductRepo = &mocks.MockProductRepository{}
    suite.mockCategoryRepo = &mocks.MockCategoryRepository{}
    suite.productService = service.NewCachedProductService(suite.mockProductRepo, suite.mockCategoryRepo)
}

func (suite *ProductServiceTestSuite) TestGetByID_Success() {
    // Arrange
    productID := uint(1)
    expectedProduct := &model.Product{
        ID:   productID,
        Name: "Test Product",
        Price: 99.99,
        Stock: 10,
    }
    
    suite.mockProductRepo.On("GetByID", productID).Return(expectedProduct, nil)
    
    // Act
    result, err := suite.productService.GetByID(productID)
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    assert.Equal(suite.T(), expectedProduct.ID, result.ID)
    assert.Equal(suite.T(), expectedProduct.Name, result.Name)
    
    suite.mockProductRepo.AssertExpectations(suite.T())
}

func TestProductServiceTestSuite(t *testing.T) {
    suite.Run(t, new(ProductServiceTestSuite))
}
```

## 🔧 阶段三：Repository层单元测试

### 3.1 使用sqlmock测试数据库操作

**tests/unit/repository/user_repository_test.go**:
```go
package repository_test

import (
    "database/sql/driver"
    "testing"
    "time"
    "ryan-mall/internal/model"
    "ryan-mall/internal/repository"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
    suite.Suite
    db   *gorm.DB
    mock sqlmock.Sqlmock
    repo repository.UserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
    var err error
    var sqlDB *sql.DB
    
    sqlDB, suite.mock, err = sqlmock.New()
    assert.NoError(suite.T(), err)
    
    suite.db, err = gorm.Open(mysql.New(mysql.Config{
        Conn:                      sqlDB,
        SkipInitializeWithVersion: true,
    }), &gorm.Config{})
    assert.NoError(suite.T(), err)
    
    suite.repo = repository.NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
    // Arrange
    user := &model.User{
        Username:     "testuser",
        Email:        "test@example.com",
        PasswordHash: "hashed_password",
    }
    
    suite.mock.ExpectBegin()
    suite.mock.ExpectExec("INSERT INTO `users`").
        WithArgs(user.Username, user.Email, user.PasswordHash, sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    suite.mock.ExpectCommit()
    
    // Act
    err := suite.repo.Create(user)
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func TestUserRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(UserRepositoryTestSuite))
}
```

## 🌐 阶段四：Handler层集成测试

### 4.1 API端到端测试

**tests/integration/api/user_api_test.go**:
```go
package api_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "ryan-mall/internal/model"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type UserAPITestSuite struct {
    suite.Suite
    router *gin.Engine
    server *httptest.Server
}

func (suite *UserAPITestSuite) SetupSuite() {
    gin.SetMode(gin.TestMode)
    // 初始化测试路由
    suite.router = setupTestRouter()
    suite.server = httptest.NewServer(suite.router)
}

func (suite *UserAPITestSuite) TearDownSuite() {
    suite.server.Close()
}

func (suite *UserAPITestSuite) TestUserRegister_Success() {
    // Arrange
    registerReq := model.UserRegisterRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    jsonData, _ := json.Marshal(registerReq)
    
    // Act
    resp, err := http.Post(
        suite.server.URL+"/api/v1/register",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
    
    var response map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&response)
    
    assert.Equal(suite.T(), float64(200), response["code"])
    assert.NotNil(suite.T(), response["data"])
}

func TestUserAPITestSuite(t *testing.T) {
    suite.Run(t, new(UserAPITestSuite))
}
```

## 📊 阶段五：测试覆盖率和报告

### 5.1 生成测试覆盖率报告

**scripts/test_coverage.sh**:
```bash
#!/bin/bash

echo "🧪 运行测试并生成覆盖率报告..."

# 创建覆盖率目录
mkdir -p coverage

# 运行所有测试并生成覆盖率
go test -v -coverprofile=coverage/coverage.out ./...

# 生成HTML报告
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# 显示覆盖率统计
go tool cover -func=coverage/coverage.out

echo "✅ 测试完成！覆盖率报告已生成到 coverage/coverage.html"
```

### 5.2 创建测试运行脚本

**scripts/run_tests.sh**:
```bash
#!/bin/bash

echo "🚀 开始运行Ryan Mall测试套件..."

# 运行单元测试
echo "📝 运行单元测试..."
go test -v ./tests/unit/...

# 运行集成测试
echo "🔗 运行集成测试..."
go test -v ./tests/integration/...

# 运行性能测试
echo "⚡ 运行性能测试..."
go test -v -bench=. ./tests/performance/...

echo "✅ 所有测试完成！"
```

## 🎯 TDD实践要点

### 1. 红-绿-重构循环
1. **红**: 编写失败的测试
2. **绿**: 编写最少代码使测试通过
3. **重构**: 优化代码结构

### 2. 测试命名规范
```go
func TestMethodName_Scenario_ExpectedResult(t *testing.T)
// 例如：TestRegister_UsernameExists_ReturnsError
```

### 3. 测试结构（AAA模式）
- **Arrange**: 准备测试数据
- **Act**: 执行被测试的方法
- **Assert**: 验证结果

### 4. Mock使用原则
- 只Mock外部依赖
- 验证Mock调用
- 保持Mock简单

## 📈 预期成果

完成后您将获得：
1. **85%+的测试覆盖率**
2. **完整的测试套件**
3. **自动化测试流程**
4. **测试报告和文档**
5. **TDD开发经验**

这将成为您简历上的重要亮点，展示您对代码质量的重视和企业级开发经验。

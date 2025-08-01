# 📅 Day 1: 测试环境搭建实施指南

## 🎯 今日目标

搭建完整的测试环境，为后续的TDD开发做好准备。

**预期成果**: 
- 测试依赖安装完成
- 测试目录结构创建
- 测试配置文件就绪
- 第一个示例测试运行成功

## 🚀 实施步骤

### 步骤1: 安装测试依赖包（15分钟）

在项目根目录执行以下命令：

```bash
# 进入项目目录
cd /home/raymond/桌面/ryan_Mall

# 安装测试相关依赖
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock  
go get github.com/stretchr/testify/suite
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-redis/redismock/v8

# 更新go.mod
go mod tidy

# 验证依赖安装
go list -m all | grep testify
```

### 步骤2: 创建测试目录结构（10分钟）

```bash
# 创建测试目录结构
mkdir -p tests/{unit,integration,performance}
mkdir -p tests/unit/{service,repository,handler}
mkdir -p tests/integration/{api}
mkdir -p tests/mocks
mkdir -p tests/fixtures
mkdir -p tests/config
mkdir -p coverage
mkdir -p scripts

# 验证目录结构
tree tests/
```

预期目录结构：
```
tests/
├── unit/
│   ├── service/
│   ├── repository/
│   └── handler/
├── integration/
│   └── api/
├── performance/
├── mocks/
├── fixtures/
└── config/
```

### 步骤3: 创建测试配置文件（20分钟）

**创建 tests/config/test_config.go**:
```go
package config

import (
    "ryan-mall/internal/config"
)

// GetTestConfig 获取测试环境配置
func GetTestConfig() *config.Config {
    return &config.Config{
        Database: config.DatabaseConfig{
            Host:     "localhost",
            Port:     "3306", 
            Username: "root",
            Password: "123456",
            Database: "ryan_mall_test",
        },
        JWT: config.JWTConfig{
            SecretKey:   "test_secret_key_for_testing_only_do_not_use_in_production",
            ExpireHours: 24,
        },
        Server: config.ServerConfig{
            Port: "8081",
            Mode: "test",
        },
        Redis: config.RedisConfig{
            Host:     "localhost",
            Port:     "6379",
            Password: "",
            Database: 1, // 使用不同的数据库避免冲突
        },
    }
}

// SetupTestDatabase 设置测试数据库
func SetupTestDatabase() error {
    // 这里可以添加测试数据库初始化逻辑
    // 例如：创建测试数据库、运行迁移等
    return nil
}

// CleanupTestDatabase 清理测试数据库
func CleanupTestDatabase() error {
    // 这里可以添加测试数据库清理逻辑
    return nil
}
```

### 步骤4: 创建基础Mock文件（25分钟）

**创建 tests/mocks/mock_user_repository.go**:
```go
package mocks

import (
    "ryan-mall/internal/model"
    "github.com/stretchr/testify/mock"
)

// MockUserRepository 用户仓储层Mock
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*model.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
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

**创建 tests/mocks/mock_jwt_manager.go**:
```go
package mocks

import (
    "ryan-mall/pkg/jwt"
    "github.com/stretchr/testify/mock"
)

// MockJWTManager JWT管理器Mock
type MockJWTManager struct {
    mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID uint, username, email string) (string, error) {
    args := m.Called(userID, username, email)
    return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*jwt.Claims, error) {
    args := m.Called(tokenString)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*jwt.Claims), args.Error(1)
}
```

### 步骤5: 创建第一个示例测试（20分钟）

**创建 tests/unit/service/example_test.go**:
```go
package service_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestExample 示例测试，验证测试环境是否正常工作
func TestExample(t *testing.T) {
    // Arrange
    expected := "Hello, Testing!"
    
    // Act
    actual := "Hello, Testing!"
    
    // Assert
    assert.Equal(t, expected, actual, "示例测试应该通过")
}

// TestExampleWithSubtests 带子测试的示例
func TestExampleWithSubtests(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"正数", 5, 5},
        {"零", 0, 0},
        {"负数", -3, -3},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.input)
        })
    }
}
```

### 步骤6: 创建测试运行脚本（15分钟）

**创建 scripts/run_tests.sh**:
```bash
#!/bin/bash

echo "🧪 Ryan Mall 测试套件"
echo "===================="

# 设置测试环境变量
export GIN_MODE=test
export GO_ENV=test

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 运行示例测试
echo -e "${YELLOW}📝 运行示例测试...${NC}"
go test -v ./tests/unit/service/example_test.go

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 示例测试通过！${NC}"
else
    echo -e "${RED}❌ 示例测试失败！${NC}"
    exit 1
fi

# 运行所有单元测试（目前只有示例测试）
echo -e "${YELLOW}📝 运行所有单元测试...${NC}"
go test -v ./tests/unit/...

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 所有单元测试通过！${NC}"
else
    echo -e "${RED}❌ 单元测试失败！${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 测试环境搭建成功！${NC}"
```

**创建 scripts/test_coverage.sh**:
```bash
#!/bin/bash

echo "📊 生成测试覆盖率报告..."

# 创建覆盖率目录
mkdir -p coverage

# 运行测试并生成覆盖率
go test -v -coverprofile=coverage/coverage.out ./tests/unit/...

# 生成HTML报告
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# 显示覆盖率统计
echo "📈 覆盖率统计："
go tool cover -func=coverage/coverage.out

echo "✅ 覆盖率报告已生成到 coverage/coverage.html"
```

### 步骤7: 设置执行权限并运行测试（10分钟）

```bash
# 设置脚本执行权限
chmod +x scripts/run_tests.sh
chmod +x scripts/test_coverage.sh

# 运行示例测试
./scripts/run_tests.sh

# 生成覆盖率报告
./scripts/test_coverage.sh
```

### 步骤8: 验证环境搭建（5分钟）

运行以下命令验证环境：

```bash
# 检查测试依赖
go list -m github.com/stretchr/testify

# 检查目录结构
ls -la tests/

# 运行测试
go test ./tests/unit/...

# 检查覆盖率文件
ls -la coverage/
```

## ✅ 完成检查清单

- [ ] 测试依赖包安装成功
- [ ] 测试目录结构创建完成
- [ ] 测试配置文件创建完成
- [ ] Mock文件创建完成
- [ ] 示例测试创建并运行成功
- [ ] 测试脚本创建并可执行
- [ ] 覆盖率报告生成成功

## 🎯 今日成果

完成今日任务后，您将拥有：

1. **完整的测试环境** - 所有必要的依赖和配置
2. **标准的目录结构** - 便于后续测试开发
3. **基础的Mock框架** - 支持单元测试隔离
4. **自动化测试脚本** - 提升测试效率
5. **覆盖率报告工具** - 监控测试质量

## 📝 明日预告

明天我们将开始Day 2-4的Service层单元测试开发：

1. **UserService测试** - 用户注册、登录逻辑测试
2. **ProductService测试** - 商品管理逻辑测试
3. **TDD实践** - 红-绿-重构循环
4. **Mock使用** - 依赖隔离和测试独立性

准备好开始真正的TDD之旅了吗？🚀

# EntityDB Testing Guide

This directory contains comprehensive testing documentation and methodologies for EntityDB.

## 🧪 Testing Categories

### Production Testing
- **[Production Battle Testing Guide](./production-battle-testing-guide.md)** - Real-world scenario testing methodology
- **[E2E Test Suite](./e2e-testing-guide.md)** - End-to-end testing procedures
- **[Performance Testing](./performance-testing-guide.md)** - Load and stress testing

### Unit Testing
- **[Unit Test Guidelines](./unit-testing-guide.md)** - Writing effective unit tests
- **[Mock Strategies](./mocking-guide.md)** - Testing with mocks and stubs
- **[Test Coverage](./coverage-guide.md)** - Achieving comprehensive coverage

### Integration Testing  
- **[API Testing](./api-testing-guide.md)** - Testing REST endpoints
- **[Storage Testing](./storage-testing-guide.md)** - Binary format testing
- **[Temporal Testing](./temporal-testing-guide.md)** - Time-travel query testing

## 📋 Testing Standards

### Test Organization
```
src/
├── api/
│   ├── entity_handler.go
│   └── entity_handler_test.go
├── storage/binary/
│   ├── entity_repository.go
│   └── entity_repository_test.go
└── tests/
    ├── integration/
    ├── e2e/
    └── performance/
```

### Test Naming Conventions
```go
// Unit tests
func TestEntityRepository_Create(t *testing.T)
func TestEntityRepository_Create_WithLargeContent(t *testing.T)
func TestEntityRepository_Create_ConcurrentAccess(t *testing.T)

// Table-driven tests
func TestEntityValidation(t *testing.T) {
    tests := []struct {
        name    string
        entity  *Entity
        wantErr bool
    }{
        // test cases...
    }
}
```

### Performance Benchmarks
```go
func BenchmarkEntityCreate(b *testing.B) {
    // benchmark code
}

func BenchmarkMultiTagQuery(b *testing.B) {
    // benchmark code
}
```

## 🎯 Testing Best Practices

### 1. Test Isolation
- Each test should be independent
- Use fresh test databases
- Clean up after tests
- No shared state between tests

### 2. Test Data
- Use realistic test data
- Test edge cases
- Test with various entity sizes
- Test temporal boundaries

### 3. Concurrent Testing
- Test concurrent operations
- Verify race conditions
- Test lock contention
- Validate data consistency

### 4. Error Testing
- Test error conditions
- Verify error messages
- Test recovery scenarios
- Validate error handling

## 📊 Test Coverage Goals

| Component | Target Coverage | Current |
|-----------|----------------|---------|
| API Handlers | 90% | ✅ |
| Storage Layer | 95% | ✅ |
| RBAC System | 100% | ✅ |
| Temporal Queries | 90% | ✅ |
| Error Paths | 80% | ✅ |

## 🔧 Testing Tools

### Required Tools
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
make bench

# Run E2E tests
make test-e2e
```

### Continuous Integration
- All tests run on every commit
- Coverage reports generated
- Performance regression detection
- Automated security scanning

## 📚 Related Documentation

### Implementation
- [Implementation Guides](../implementation/)
- [Code Standards](../03-logging-standards.md)
- [Git Workflow](../02-git-workflow.md)

### Architecture
- [Architecture Decisions](../../architecture/adr/)
- [Technical Specs](../../reference/technical-specs/)

---

**Last Updated**: 2025-06-23  
**Purpose**: Comprehensive testing guidance for contributors  
**Maintainers**: EntityDB Core Team
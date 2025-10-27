# Phase C2: Real-World Validation - Planning

## Objectives

Validate AI Distiller against real-world codebases to ensure:
1. Parsers handle complex, production-grade code
2. Performance is acceptable on large projects
3. Output quality meets practical requirements
4. Edge cases are handled gracefully

## Test Strategy

### Approach 1: Representative Examples (Fast)
Create small but realistic code samples that mirror production patterns:
- Django-style Python (models, views, services)
- React-style TypeScript (components, hooks, context)
- Go microservice patterns (handlers, middleware, repositories)

**Pros**: Fast, controlled, reproducible
**Cons**: May miss real-world edge cases

### Approach 2: Actual Open Source (Comprehensive)
Test with subsets of real projects:
- Django/Flask subset (50-100 files)
- React/Next.js subset (50-100 files)
- Kubernetes Go code subset (50-100 files)

**Pros**: Real edge cases, authentic complexity
**Cons**: Slower, harder to reproduce, external dependencies

### Chosen Approach: Hybrid

**Phase C2.1** (Current): Representative examples with production patterns
- Quick validation of common patterns
- Establish performance baselines
- Identify obvious issues

**Phase C2.2** (If needed): Actual open-source testing
- Deep validation with real codebases
- Find subtle bugs
- Stress test at scale

## C2.1: Representative Examples

### 1. Django-Style Python App
```
testdata/real-world/django-app/
├── models.py          # SQLAlchemy/Django ORM models
├── views.py           # View functions with decorators
├── services.py        # Business logic layer
├── serializers.py     # API serializers
└── tests.py           # pytest tests
```

**Patterns to test**:
- Class-based views
- Function decorators (@login_required, @api_view)
- ORM relationships (ForeignKey, ManyToMany)
- Async views (async def)
- Type hints with complex types

### 2. React-Style TypeScript App
```
testdata/real-world/react-app/
├── components/
│   ├── UserProfile.tsx    # React component with hooks
│   ├── DataTable.tsx      # Complex component with generics
│   └── HOC.tsx            # Higher-order component
├── hooks/
│   ├── useAuth.ts         # Custom hook
│   └── useFetch.ts        # Async data fetching
├── context/
│   └── AppContext.tsx     # React context
└── utils/
    └── api.ts             # API utilities
```

**Patterns to test**:
- React components (functional + class)
- Hooks (useState, useEffect, custom)
- Generics with constraints
- Async/await patterns
- Type inference

### 3. Go Microservice
```
testdata/real-world/go-service/
├── main.go              # Entry point
├── handlers/
│   ├── user.go          # HTTP handlers
│   └── middleware.go    # Middleware chain
├── services/
│   └── user_service.go  # Business logic
├── repository/
│   └── user_repo.go     # Data access
└── models/
    └── user.go          # Domain models
```

**Patterns to test**:
- HTTP handlers with method receivers
- Interface satisfaction (implicit)
- Error wrapping (%w in fmt.Errorf)
- Context propagation
- Goroutine patterns

## Performance Benchmarks

### Metrics to Track
1. **Processing Time**
   - Single file: target < 10ms
   - Small project (10 files): target < 100ms
   - Medium project (100 files): target < 1s
   - Large project (1000 files): target < 10s

2. **Memory Usage**
   - Peak memory: track with `time -v`
   - Memory per file: should be bounded
   - No memory leaks: run multiple times

3. **Accuracy**
   - Classes/functions found: 100%
   - Parameter types: 95%+
   - Return types: 95%+
   - Visibility: 98%+

### Benchmark Scenarios
1. Cold start (first run)
2. Warm start (subsequent runs)
3. Parallel processing (default)
4. Serial processing (workers=1)
5. Different option combinations

## Success Criteria

- [ ] All representative examples parse without errors
- [ ] Performance meets targets above
- [ ] Output quality is high (spot-check manually)
- [ ] No crashes or panics
- [ ] Memory usage is reasonable
- [ ] Results are deterministic

## Implementation Plan

1. Create Django-style Python examples
2. Create React-style TypeScript examples
3. Create Go microservice examples
4. Run comprehensive tests
5. Performance benchmarking
6. Document findings
7. Fix any issues discovered

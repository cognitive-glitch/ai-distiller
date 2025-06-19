# 🧪 Testing Guide

AI Distiller používá **gotestsum** pro krásný a přehledný výstup testů s barvičkami, fajfkami a progress indikátory.

## 🚀 Quick Start

```bash
# Nainstaluj gotestsum (automaticky při make dev-init)
go install gotest.tools/gotestsum@latest

# Spusť testy s výchozím hezčím výstupem
make test
```

## 📋 Dostupné formáty testů

### ✨ Doporučené formáty

**`make test`** - Výchozí formát s názvami testů:
```
PASS command-line-arguments.TestBasicFunctionality/should_pass_simple_test (0.00s)
PASS command-line-arguments.TestBasicFunctionality/should_handle_strings (0.00s)
PASS command-line-arguments.TestBasicFunctionality (0.00s)
DONE 12 tests in 0.103s
```

**`make test-pretty`** - Nejhezčí formát s package summary:
```
✓  command-line-arguments (103ms)
DONE 12 tests in 0.103s
```

**`make test-dots`** - Progress s tečkami:
```
[command-line-arguments]············
DONE 12 tests in 0.103s
```

### 🔧 Další užitečné formáty

**`make test-short`** - Stručný verbose:
```
PASS command-line-arguments.TestBasicFunctionality/should_pass_simple_test (0.00s)
PASS command-line-arguments.TestBasicFunctionality (0.00s)
DONE 12 tests in 0.103s
```

**`make test-standard`** - Standardní formát s progress:
```
=== RUN   TestBasicFunctionality
=== RUN   TestBasicFunctionality/should_pass_simple_test
=== RUN   TestBasicFunctionality/should_handle_strings
--- PASS: TestBasicFunctionality (0.00s)
```

**`make test-github`** - GitHub Actions formát pro CI/CD:
```
::group::command-line-arguments
=== RUN   TestBasicFunctionality
--- PASS: TestBasicFunctionality (0.00s)
::endgroup::
```

**`make test-watch`** - Watch mode (rerun při změnách souborů):
```
gotestsum --watch --format testname -- -race ./...
```

**`make test-basic`** - Klasický Go test výstup (bez gotestsum):
```
=== RUN   TestBasicFunctionality
=== RUN   TestBasicFunctionality/should_pass_simple_test
    example_test.go:9: Test message
--- PASS: TestBasicFunctionality/should_pass_simple_test (0.00s)
```

## 🎯 Failure Output 

Při neúspěšných testech gotestsum krásně zvýrazní chyby:

```
✖  command-line-arguments (2ms)

=== Failed
=== FAIL: command-line-arguments TestFailingScenarios/should_fail_intentionally (0.00s)
    example_failing_test.go:11: Expected 5, got 4

DONE 3 tests, 2 failures in 0.002s
```

## 📊 Coverage & Race Detection

Všechny formáty zahrnují:
- **Race detection** (`-race`)
- **Coverage report** (`-coverprofile=coverage.txt`)
- **Atomic coverage mode** (`-covermode=atomic`)

## ⚡ Performance Tips

- **`test-pretty`** - Nejrychlejší, jen summary
- **`test-dots`** - Dobrý pro dlouhé test suite
- **`test-short`** - Balans mezi detaily a rychlostí
- **`test-watch`** - Pro vývoj, automatický rerun

## 🔧 Installation

gotestsum se automaticky nainstaluje při:
```bash
make dev-init
```

Nebo manuálně:
```bash
go install gotest.tools/gotestsum@latest
```

## 📁 Test Organization

```
internal/
├── example_test.go           # ✅ Ukázkové úspěšné testy
├── example_failing_test.go   # ❌ Ukázkové neúspěšné testy
└── testrunner/              # 🧪 Integration test runner
    └── integration_test.go
```

## 🎨 Další možnosti

Pro ještě více možností customizace, viz [gotestsum documentation](https://github.com/gotestyourself/gotestsum).

Další dostupné formáty:
- `testdox` - BDD-style output
- `quiet` - Pouze chyby
- `silent` - Bez výstupu (jen exit kód)
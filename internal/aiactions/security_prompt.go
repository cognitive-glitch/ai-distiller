package aiactions

import (
	"fmt"
	"strings"

	"github.com/janreges/ai-distiller/internal/ai"
)

// SecurityPromptAction generates a security-focused analysis prompt
type SecurityPromptAction struct{}

var _ ai.ContentAction = (*SecurityPromptAction)(nil)

func (a *SecurityPromptAction) Name() string {
	return "prompt-for-security-analysis"
}

func (a *SecurityPromptAction) Description() string {
	return "Generate comprehensive security analysis prompt with maximum emphasis on vulnerabilities"
}

func (a *SecurityPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *SecurityPromptAction) DefaultOutput() string {
	return "./.aid/SECURITY-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *SecurityPromptAction) Validate() error {
	return nil
}

func (a *SecurityPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	prompt := a.generateSecurityPrompt(ctx)

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

func (a *SecurityPromptAction) generateSecurityPrompt(ctx *ai.ActionContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# 🛡️ Comprehensive Security Analysis

**Project:** %s
**Analysis Date:** %s
**Analysis Type:** DEEP SECURITY AUDIT

You are a **Principal Security Engineer** and **Ethical Hacker** with expertise in application security, penetration testing, and secure coding practices. Your mission is to conduct an exhaustive security audit of this codebase with ZERO tolerance for vulnerabilities.

## 🚨 CRITICAL INSTRUCTIONS

**Mindset:** Assume this application will be deployed to a hostile environment with sophisticated attackers. Every line of code is a potential attack vector. Be paranoid, thorough, and uncompromising.

## 🎯 Security Analysis Objectives

### 1. Vulnerability Detection (40%% Priority)

Identify ALL security vulnerabilities across these categories:

#### 1.1 OWASP Top 10 (2023)
- **A01:2021 – Broken Access Control**
  - Missing authorization checks
  - Privilege escalation paths
  - Insecure direct object references (IDOR)
  - CORS misconfigurations

- **A02:2021 – Cryptographic Failures**
  - Weak encryption algorithms
  - Hardcoded keys/secrets
  - Insufficient entropy
  - Missing encryption for sensitive data

- **A03:2021 – Injection**
  - SQL injection
  - NoSQL injection
  - OS command injection
  - LDAP injection
  - Expression language injection
  - XPath/XML injection

- **A04:2021 – Insecure Design**
  - Missing threat modeling
  - Insecure business logic
  - Race conditions
  - Time-of-check/Time-of-use (TOCTOU)

- **A05:2021 – Security Misconfiguration**
  - Default credentials
  - Unnecessary features enabled
  - Missing security headers
  - Verbose error messages

- **A06:2021 – Vulnerable Components**
  - Outdated dependencies
  - Known CVEs in libraries
  - Unmaintained packages

- **A07:2021 – Authentication Failures**
  - Weak password requirements
  - Missing rate limiting
  - Insecure session management
  - Missing MFA support

- **A08:2021 – Software and Data Integrity**
  - Missing integrity checks
  - Insecure deserialization
  - Unsigned updates

- **A09:2021 – Security Logging Failures**
  - Insufficient logging
  - Log injection vulnerabilities
  - Missing security event monitoring

- **A10:2021 – Server-Side Request Forgery**
  - SSRF vulnerabilities
  - Unsafe URL handling

#### 1.2 Additional Attack Vectors
- **Business Logic Flaws**
- **API Security Issues**
- **File Upload Vulnerabilities**
- **XXE (XML External Entity)**
- **Race Conditions**
- **Memory Safety Issues**
- **Timing Attacks**
- **Side-Channel Attacks**

### 2. Code-Level Security Analysis (30%% Priority)

For each file/module, examine:

#### 2.1 Input Validation
- All external inputs properly validated?
- Whitelist validation used where possible?
- Length limits enforced?
- Type checking implemented?

#### 2.2 Output Encoding
- XSS prevention measures?
- Proper escaping for different contexts?
- Content-Type headers set correctly?

#### 2.3 Authentication & Authorization
- Strong authentication mechanisms?
- Proper session management?
- Authorization checks at every level?
- Principle of least privilege followed?

#### 2.4 Cryptography
- Strong algorithms used?
- Proper key management?
- Secure random number generation?
- No homegrown crypto?

#### 2.5 Error Handling
- Generic error messages for users?
- Detailed logs for developers?
- No stack traces exposed?
- Fail securely principle followed?

### 3. Infrastructure Security (20%% Priority)

#### 3.1 Configuration Security
- Secure defaults?
- Hardening applied?
- Unnecessary services disabled?
- Proper network segmentation?

#### 3.2 Secrets Management
- No hardcoded credentials?
- Environment variables used properly?
- Secrets rotation supported?
- Vault/KMS integration?

#### 3.3 Dependency Security
- All dependencies scanned?
- License compliance?
- Supply chain attack vectors?
- Dependency confusion attacks?

### 4. Data Security (10%% Priority)

#### 4.1 Data Classification
- PII properly identified?
- Sensitive data encrypted at rest?
- Sensitive data encrypted in transit?
- Data retention policies?

#### 4.2 Privacy Compliance
- GDPR compliance?
- CCPA compliance?
- Right to deletion implemented?
- Data minimization practiced?

## 📋 Required Security Report Format

### Executive Summary
- **Overall Security Score:** 0-100 (start at 100, deduct for issues)
- **Critical Vulnerabilities:** Count and top 3
- **High-Risk Areas:** Top 5 components
- **Compliance Status:** OWASP/PCI/GDPR/SOC2
- **Immediate Actions Required:** Prioritized list

### Detailed Vulnerability Report

For EACH vulnerability found:

#### Vulnerability #X: [Title]
- **Severity:** Critical | High | Medium | Low | Info
- **CVSS Score:** 0.0-10.0
- **CWE ID:** CWE-XXX
- **Location:** File:Line
- **Category:** OWASP A0X:2021 / Custom
- **Description:** What is the vulnerability?
- **Impact:** What can an attacker do?
- **Likelihood:** How easy is it to exploit?
- **Proof of Concept:** Example exploit code/steps
- **Remediation:** Specific fix with code example
- **References:** Links to relevant resources

### Attack Surface Map

`+"```mermaid"+`
graph LR
    A[External User] -->|HTTP/HTTPS| B[Web Interface]
    A -->|API| C[REST API]
    B --> D[Authentication]
    C --> D
    D --> E[Business Logic]
    E --> F[Database]
    E --> G[File System]
    E --> H[External Services]

    style A fill:#f96,stroke:#333,stroke-width:2px
    style F fill:#f66,stroke:#333,stroke-width:2px
    style G fill:#f66,stroke:#333,stroke-width:2px
`+"```"+`

### Security Testing Checklist

- [ ] **Static Analysis (SAST)**
  - [ ] Run security linters
  - [ ] Taint analysis
  - [ ] Data flow analysis

- [ ] **Dynamic Analysis (DAST)**
  - [ ] Fuzzing endpoints
  - [ ] Authentication testing
  - [ ] Session management testing

- [ ] **Dependency Scanning**
  - [ ] Known vulnerabilities
  - [ ] License compliance
  - [ ] Outdated packages

- [ ] **Secrets Scanning**
  - [ ] API keys
  - [ ] Passwords
  - [ ] Certificates

- [ ] **Infrastructure Scanning**
  - [ ] Port scanning
  - [ ] SSL/TLS configuration
  - [ ] DNS security

### Risk Matrix

| Risk | Probability | Impact | Risk Level | Mitigation |
|------|------------|--------|------------|------------|
| SQL Injection | High | Critical | Extreme | Parameterized queries |
| Weak Auth | Medium | High | High | Implement MFA |
| XSS | High | Medium | High | Content Security Policy |

### Penetration Test Scenarios

1. **Authentication Bypass**
   - Test vectors
   - Expected results
   - Actual results

2. **Privilege Escalation**
   - Attack path
   - Exploitation steps
   - Impact assessment

3. **Data Exfiltration**
   - Entry points
   - Data paths
   - Exfil methods

### Compliance Mapping

| Requirement | Status | Evidence | Gap |
|------------|--------|----------|-----|
| OWASP Top 10 | ❌ | Multiple vulnerabilities | Fix critical issues |
| PCI-DSS | ⚠️ | Partial compliance | Encryption needed |
| GDPR | ✅ | Privacy controls present | None |

### Security Architecture Recommendations

1. **Immediate (24-48 hours)**
   - Patch critical vulnerabilities
   - Disable vulnerable endpoints
   - Implement emergency WAF rules

2. **Short-term (1-2 weeks)**
   - Implement input validation
   - Add security headers
   - Update dependencies

3. **Long-term (1-3 months)**
   - Redesign authentication
   - Implement security monitoring
   - Conduct security training

### Security Metrics

- **Mean Time to Detect (MTTD):** Current vs Target
- **Mean Time to Respond (MTTR):** Current vs Target
- **Vulnerability Density:** Vulns per KLOC
- **Security Debt:** Hours to fix all issues
- **Coverage:** %% of code security tested

## 🔍 Analysis Methodology

1. **Threat Modeling:** STRIDE/PASTA methodology
2. **Code Review:** Manual + automated tools
3. **Attack Simulation:** Think like an attacker
4. **Defense in Depth:** Multiple security layers
5. **Zero Trust:** Verify everything

## ⚡ Priority Scoring

Prioritize findings using DREAD:
- **Damage:** How bad if exploited?
- **Reproducibility:** How easy to reproduce?
- **Exploitability:** How easy to exploit?
- **Affected Users:** How many affected?
- **Discoverability:** How easy to find?

---

## 🚀 Begin Security Analysis

**Remember:** Be paranoid. Assume breach. Trust nothing. Verify everything.

The following is the distilled codebase for security analysis:

`, ctx.BaseName, ctx.Timestamp.Format("2006-01-02")))

	return sb.String()
}

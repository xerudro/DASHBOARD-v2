# AI-Assisted Development - Quick Reference Card
## For Claude & GitHub Copilot Usage

**Print this and keep at your desk!**

---

## QUICK PROMPTING TEMPLATES

### Template 1: Implementation Request
```
[Paste relevant section from ai-dev-system-prompts.md]

Implement [feature]:
- Accept: [inputs]
- Return: [outputs]
- Must handle: [errors]
- Security: [specific concerns]

Include: Tests, docs, error handling, examples
```

### Template 2: Code Review
```
[Paste CODE QUALITY section]

Review this code for:
✓ Security (input validation, auth, encryption)
✓ Performance (N+1 queries, indexes, caching)
✓ Testing (coverage, edge cases)
✓ Error handling (all cases covered)

Code:
[paste code]

Issues found & fixes?
```

### Template 3: Performance Optimization
```
[Paste PERFORMANCE section]

Optimize for:
- Latency: [target]
- Throughput: [target]
- Resource limits: [CPU/RAM/disk]

Current approach:
[describe]

Provide analysis + optimized implementation
```

### Template 4: Security Audit
```
[Paste SECURITY STANDARDS]

Audit this for vulnerabilities:
- OWASP Top 10
- Common auth flaws
- Data protection issues
- Error leakage

Code:
[paste]

Found issues + fixes?
```

---

## TECH STACK CHECKLIST

### Backend
- [ ] Rust 1.75+ with Tokio
- [ ] Actix-web 4.x
- [ ] PostgreSQL 14+
- [ ] Redis 7.x
- [ ] RabbitMQ 3.12+

### Frontend
- [ ] React 18.x
- [ ] TypeScript 5.x (strict mode)
- [ ] Vite 5.x
- [ ] Redux Toolkit 1.9.x
- [ ] TailwindCSS 3.x

### DevOps
- [ ] Docker 24.x
- [ ] Kubernetes 1.28.x (or Docker Compose)
- [ ] GitHub Actions CI/CD
- [ ] Prometheus + Grafana
- [ ] Terraform 1.6.x

---

## SECURITY CHECKLIST

### Every API Endpoint
- [ ] Authentication required
- [ ] Authorization checked
- [ ] Input validated
- [ ] SQL injection prevented (parameterized)
- [ ] Error messages don't leak info
- [ ] Rate limiting applied

### Every Database Query
- [ ] Parameterized (not string concat)
- [ ] Indexes used
- [ ] No N+1 problems
- [ ] Results pagination

### Every Secret
- [ ] NOT hardcoded
- [ ] NOT in git
- [ ] NOT in env files
- [ ] Use Vault/Secrets Manager
- [ ] Rotated regularly

### Every Deployment
- [ ] Tests pass
- [ ] Security scan passed
- [ ] No warnings/errors
- [ ] Staged to staging first
- [ ] Monitored for 24 hours

---

## CODE QUALITY CHECKLIST

### Before Commit
- [ ] Code compiles/runs
- [ ] `cargo fmt` and `prettier` run
- [ ] `cargo clippy` passes (no warnings)
- [ ] Tests pass locally
- [ ] No console.log/debug statements
- [ ] No hardcoded values

### Before PR
- [ ] Unit tests added
- [ ] Integration tests pass
- [ ] > 80% code coverage
- [ ] Error cases tested
- [ ] Documentation added
- [ ] Examples provided

### Before Merge
- [ ] Code review approved
- [ ] All CI checks pass
- [ ] Security scan passed
- [ ] Performance impact assessed
- [ ] Database migrations safe
- [ ] Backward compatible

---

## TESTING PYRAMID

```
      /\
     /  \      E2E (10%) - Full user workflows
    /____\
   /      \
  / Integ  \   Integration (20%) - Components together
 /  Tests  \
/___________\
/           \
/ Unit Tests \ Unit (70%) - Individual functions
/____________\
```

**Rules:**
- ✓ 100% coverage for critical paths
- ✓ >80% overall coverage
- ✓ Test happy path + all error cases
- ✓ Test edge cases
- ✓ Test performance (load test at 10x)

---

## ERROR HANDLING PATTERN

```rust
// Define error type
#[derive(Debug)]
pub enum ServiceError {
    NotFound,
    Unauthorized,
    ValidationFailed(String),
    DatabaseError(String),
}

// Convert to HTTP response
impl Into<HttpResponse> for ServiceError {
    fn into(self) -> HttpResponse {
        match self {
            ServiceError::NotFound => 
                HttpResponse::NotFound().json(...),
            // ... other cases
        }
    }
}

// Use in handler
#[get("/resource/{id}")]
async fn get(id: i32) -> Result<Json<Resource>, ServiceError> {
    let resource = db.get(id).await
        .map_err(|e| {
            error!("DB error: {}", e);
            ServiceError::DatabaseError(...)
        })?;
    Ok(Json(resource))
}
```

---

## AUTHENTICATION PATTERN

```rust
// Password: Use Argon2
let hash = Argon2::default()
    .hash_password(pwd.as_bytes(), &salt)?
    .to_string();

// Token: Use JWT
let claims = Claims { sub: user_id, exp: now() + 3600, ... };
let token = jsonwebtoken::encode(&header, &claims, &key)?;

// 2FA: Use TOTP
// Base32 secret, 30-second windows, 6-digit codes

// Session: Use Redis, not memory
redis.set_ex(session_id, user_data, 3600)?;
```

---

## DATABASE PATTERN

```sql
-- Always use:
CREATE TABLE websites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    domain VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,  -- Soft delete
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Migrations: Always include UP and DOWN
-- Queries: Always parameterized (never concat strings)
-- Updates: Always include deleted_at for soft deletes
```

---

## API RESPONSE FORMAT

```json
// Success
{
  "status": "success",
  "data": { /* resource */ },
  "meta": {
    "timestamp": "2025-11-02T15:45:23Z",
    "version": "1.0"
  }
}

// Error
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "User-friendly message",
    "details": { "field": "email", "reason": "invalid" }
  }
}

// Paginated
{
  "status": "success",
  "data": [/* items */],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 1500,
    "has_next": true
  }
}
```

---

## GIT WORKFLOW

```bash
# Branch naming
feature/user-auth              # New feature
bugfix/email-validation        # Bug fix
hotfix/security-patch          # Emergency fix
release/v1.0.0                 # Release

# Commit message format
<type>: <subject>

<body>

<footer>

# Types: feat, fix, refactor, docs, test, perf, security, ci, chore

# Example:
feat: implement two-factor authentication

Add TOTP-based 2FA support
- Generate base32 secrets
- Validate 6-digit codes
- Generate backup codes

Closes #123
```

---

## DEPLOYMENT CHECKLIST

**24 Hours Before:**
- [ ] Test in staging
- [ ] Database backup created
- [ ] Rollback plan documented

**1 Hour Before:**
- [ ] All tests passing
- [ ] Security scan passed
- [ ] Performance baseline known

**Deployment:**
- [ ] Deploy to staging (if not already)
- [ ] Smoke tests pass
- [ ] Deploy canary (5% of servers)
- [ ] Monitor for errors
- [ ] Gradually increase (25%, 50%, 100%)
- [ ] Full deployment
- [ ] Monitor for 24 hours

**If Issues:**
- [ ] Automatic rollback if error rate > 5%
- [ ] Manual rollback available (< 5 min)
- [ ] Incident post-mortem within 24 hours

---

## MONITORING ALERTS

- CPU > 80%: WARNING
- CPU > 95%: CRITICAL (respond in 5 min)
- Memory > 85%: WARNING
- Disk > 90%: WARNING
- Error rate > 1%: WARNING
- Error rate > 5%: CRITICAL
- API latency > 500ms: WARNING
- API latency > 2s: CRITICAL

---

## COMMON GOTCHAS

### ❌ NEVER:
- Hardcode secrets
- Use `panic!()` in production code
- Deploy on Friday
- Skip testing
- Ignore compiler warnings
- Mix encrypted + unencrypted data
- Use `unwrap()` without fallback
- Deploy without staging test
- Skip database backup
- Commit `.env` file

### ✅ ALWAYS:
- Use `Result<T, E>`
- Validate all input
- Log important events
- Test error cases
- Review all code
- Use prepared queries
- Lock dependencies
- Monitor in production
- Backup data
- Have rollback plan

---

## RESPONSE TIMES

**Target Performance:**
- API response: < 100ms (95th percentile)
- Database query: < 50ms (average)
- Page load: < 200ms
- Throughput: 10,000 req/sec per server
- Error rate: < 0.1%

**Benchmark Command:**
```bash
# Load test
cargo bench

# k6 load test (if using)
k6 run script.js --vus 100 --duration 30s
```

---

## DOCUMENTATION MUST-HAVES

Every public function:
- [ ] Purpose statement
- [ ] Parameter descriptions
- [ ] Return type & description
- [ ] Error cases & reasons
- [ ] Usage example

Every API endpoint:
- [ ] Authorization level
- [ ] Request format
- [ ] Response format
- [ ] Error responses
- [ ] Rate limits
- [ ] Examples

---

## COPILOT KEYBOARD SHORTCUTS

**VSCode:**
- `Ctrl+I` (Cmd+I on Mac) - Inline chat
- `Ctrl+K` (Cmd+K on Mac) - Open chat panel
- `/explain` - Explain selected code
- `/fix` - Fix code issues
- `/tests` - Generate tests
- `/doc` - Add documentation

---

## CLAUDE BEST PRACTICES

**Effective:**
- Use section references: "Following section 3.2..."
- Chain complex tasks: "First design, then implement"
- Ask for alternatives: "Show two approaches"
- Request explanations: "Why this pattern?"
- Provide context: [Paste standards]

**Less Effective:**
- "Write a function" (no context)
- "Make it faster" (no constraints)
- "Add security" (no specific threats)
- Ignoring suggestions without explanation
- Deploying without review

---

## ESCALATION CONTACTS

**Technical Lead:** [email]  
**Security Lead:** [email]  
**DevOps Lead:** [email]  
**Database Admin:** [email]  

**Slack:** #development-standards  
**Wiki:** https://wiki.internal.example.com/development

---

## QUICK LINKS

- OpenAPI 3.1: https://spec.openapis.org/
- Rust Book: https://doc.rust-lang.org/book/
- React Hooks: https://react.dev/reference/react
- PostgreSQL: https://www.postgresql.org/docs/
- Kubernetes: https://kubernetes.io/docs/
- OWASP Top 10: https://owasp.org/www-project-top-ten/

---

**Remember:**
✨ Quality over speed  
🔒 Security first  
🧪 Test everything  
📚 Document well  
👥 Review code  
📊 Monitor constantly  
🚀 Deploy confidently  

---

**Print & Laminate!**
Keep this at your desk while developing.

**Version:** 1.0 | **Date:** November 2, 2025 | **Review:** May 2026
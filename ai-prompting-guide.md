# How to Use These System Prompts Effectively
## Guidelines for AI-Assisted Development with Claude & GitHub Copilot

**Document Version:** 1.0  
**Date:** November 2, 2025  

---

## QUICK START GUIDE

### For Claude (Chat Interface)

**When asking Claude for code/architectural help:**

```
[Copy the relevant section from ai-dev-system-prompts.md]

Now, I need you to [specific request]:

[Your specific question or requirement]

Please ensure:
- Latest stack versions (Rust 1.75+, React 18+, PostgreSQL 14+)
- Enterprise-grade security practices
- Full error handling
- Complete test coverage
- Documentation included
```

**Example:**

```
[Paste: TECHNOLOGY STACK REQUIREMENTS & SECURITY STANDARDS SECTIONS]

I need you to implement a user authentication endpoint in Rust/Actix-web that:
- Accepts email and password
- Uses Argon2 for password hashing
- Returns JWT token on success
- Implements rate limiting
- Includes comprehensive error handling
- Has full test coverage
- Follows all security best practices from the document

Generate production-ready code with documentation.
```

### For GitHub Copilot (VSCode Extension)

**Setup Copilot with System Context:**

1. Create file: `.copilot/system-prompt.md` in project root
2. Paste content from `ai-dev-system-prompts.md`
3. Configure VSCode settings:

```json
{
  "github.copilot.enable": {
    "markdown": true,
    "plaintext": true
  },
  "copilot.advanced": {
    "debug.overrideEngine": "gpt-4",
    "debug.testOverrideProxyUrl": "",
    "debug.overrideProxy": ""
  }
}
```

4. Use Copilot Chat with `@` mentions for context:

```
@workspace Implement the authentication service following:
- System prompt standards in .copilot/system-prompt.md
- Rust with Actix-web
- Argon2 password hashing
- JWT tokens
- Rate limiting
- Full test coverage
```

---

## PROMPT ENGINEERING TIPS

### 1. Chain-of-Thought for Complex Requests

Instead of:
```
"Build me a provisioning system"
```

Use:
```
"Build a provisioning system that:

1. DESIGN PHASE:
   - What components are needed?
   - What are the dependencies?
   - What are failure modes?

2. IMPLEMENTATION PHASE:
   - Start with database schema
   - Then service layer
   - Then API endpoints
   - Then error handling

3. TESTING PHASE:
   - Unit tests
   - Integration tests
   - Error scenarios

4. DOCUMENTATION PHASE:
   - API docs
   - Error codes
   - Examples

Please follow this structure and explain your approach at each step."
```

### 2. Be Specific About Requirements

**Vague:**
```
"Write a database migration"
```

**Specific:**
```
"Write a PostgreSQL migration that:
- Creates websites table with proper indexing
- Adds created_at, updated_at, deleted_at timestamps
- Includes audit fields (created_by, updated_by)
- Adds CHECK constraints for status enum
- Includes rollback migration
- Follows the schema design standards from section 10.1
- Is safe for online deployment"
```

### 3. Provide Examples of Your Codebase

```
"This is our current error handling pattern:

[Paste your error type definitions]

Please implement [new feature] following this exact pattern.
"
```

### 4. Ask for Alternatives & Trade-offs

```
"Implement X in two ways:

Approach 1: [your preference]
- Pros:
- Cons:

Approach 2: [alternative]
- Pros:
- Cons:

Which is better for a system that needs to scale to 10,000 websites?"
```

---

## WORKFLOW INTEGRATION

### GitHub Copilot in VSCode

**When Writing Code:**

1. **Start with comments describing what you want:**
   ```rust
   // Validate user email format using regex
   // Check against blacklist of disposable domains
   // Return custom error with reason
   // Handle database errors gracefully
   fn validate_email(email: &str) -> Result<(), ValidationError> {
   ```

2. **Let Copilot suggest implementation**

3. **Review the suggestion against checklist:**
   - ✓ Follows security standards?
   - ✓ Has error handling?
   - ✓ Uses latest APIs?
   - ✓ Has comments for "why"?
   - ✓ Matches coding style?

4. **If not perfect, refine:**
   - Ask Copilot to add tests
   - Ask to add error handling
   - Ask to optimize performance

**Example in VSCode:**

```typescript
// Step 1: Write comment describing intent
// Fetch websites for authenticated user with pagination
// Support sorting by domain, created_at, status
// Return paginated response with metadata
// Handle database errors gracefully

// Step 2: Copilot autocompletes
// Accept suggestions if good, reject if not

// Step 3: Request improvements
// @copilot Add TypeScript types and error handling
```

### Claude for Design & Architecture

**Best Used For:**

- ✓ Architectural decisions (monolith vs microservices)
- ✓ Performance optimization strategies
- ✓ Security review of design
- ✓ Error handling patterns
- ✓ Testing strategies
- ✓ Refactoring advice
- ✓ Integration approaches

**Example:**

```
[Paste security standards section]

We're designing the provisioning system. Here's our approach:

[Describe your design]

Please review this against:
1. Security best practices
2. Performance considerations  
3. Failure modes
4. Scalability to 10,000 websites

Should we change anything?
```

---

## SECURITY-FIRST PROMPTING

### Template: Security Review Request

```
[Paste relevant security section]

Please review this code for security vulnerabilities:

[Paste code]

Specifically check:
- Input validation
- SQL injection risks
- XSS vulnerabilities
- Authorization checks
- Error information leakage
- Cryptographic correctness
- Rate limiting implementation
- Session management
```

### Template: Secure Implementation Request

```
[Paste SECURITY STANDARDS section]

Implement [feature] with security-first approach:

Requirements:
- Accept [inputs]
- Return [outputs]
- Must prevent: [OWASP Top 10 vectors]
- Must validate: [specific validations]
- Must encrypt: [sensitive data]
- Must log: [audit events]

Provide:
1. Threat model
2. Implementation
3. Security tests
4. Documentation of security measures
```

---

## PERFORMANCE-FOCUSED PROMPTING

### Template: Performance Optimization

```
[Paste relevant performance section]

Optimize this operation for performance:

Current approach:
[Describe current implementation]

Constraints:
- Must scale to [volume]
- Target latency: [milliseconds]
- Resource limits: [CPU/memory]
- Can change architecture? [yes/no]

Provide:
1. Analysis of bottlenecks
2. Optimization strategies
3. Trade-offs for each approach
4. Benchmarking approach
5. Implementation
```

---

## TESTING-FOCUSED PROMPTING

### Template: Test Coverage Request

```
[Paste CODE QUALITY section on testing]

Write comprehensive tests for this function:

Function signature:
[Paste function]

Requirements:
- Happy path tests
- Error scenarios: [list error cases]
- Edge cases: [specific edge cases]
- Performance tests: target [latency/throughput]
- Integration tests: [what to integrate with]
- Target coverage: [percentage]

Include:
1. Unit tests
2. Integration tests  
3. Test data fixtures
4. Mock setups
5. Assertion guidelines
```

---

## PROMPTS FOR SPECIFIC TASKS

### 1. Refactoring Existing Code

```
[Paste ARCHITECTURE & DESIGN PATTERNS]

Refactor this code to follow these standards:

Current code:
[Paste code]

Refactoring goals:
- Extract service layer
- Add dependency injection
- Improve error handling
- Add comprehensive logging
- Add tests

Provide:
1. Before/after comparison
2. Explanation of changes
3. Migration steps
4. New tests
```

### 2. Performance Debugging

```
[Paste PERFORMANCE BENCHMARKING section]

Debug this performance issue:

Symptom: [what's slow]
Expected performance: [target latency/throughput]
Current performance: [actual performance]

Environment:
- Server specs: [CPU, RAM, disk]
- Load: [requests/sec, data size]
- Database: [type, version, size]

Investigation so far:
[what you've checked]

Provide:
1. Likely bottlenecks (ranked by probability)
2. Debugging steps for each
3. Optimization strategies
4. Expected improvement
5. Implementation code
```

### 3. Security Audit

```
[Paste SECURITY STANDARDS section]

Perform security audit of this codebase:

Areas to review:
1. Authentication/Authorization - [what's implemented]
2. Input validation - [where validation occurs]
3. Data protection - [what's encrypted]
4. API security - [endpoint protection]
5. Error handling - [how errors are handled]

Provide:
1. Vulnerabilities found (with severity)
2. Proof of concept for critical issues
3. Remediation steps with code
4. Testing to verify fixes
5. Preventive measures
```

### 4. Dependency Updates

```
[Paste TECHNOLOGY STACK section]

Our project has these dependencies:
[Paste Cargo.toml or package.json]

Update to latest stable versions following:
- Rust 1.75+ with Tokio 1.35+
- React 18.x with TypeScript 5.x
- All security patches

Provide:
1. List of updates with versions
2. Breaking changes for each
3. Migration steps
4. Testing checklist
5. Deployment plan
```

### 5. New Feature Implementation

```
[Paste all relevant standard sections]

Implement new feature: [feature name]

Requirements:
- [requirement 1]
- [requirement 2]
- [requirement 3]

Must follow:
- Tech stack: Rust/React/PostgreSQL
- Security standards: [list specific sections]
- Test coverage: >80%
- Documentation: API + examples

Provide:
1. Design & architecture
2. Backend implementation (Rust)
3. Frontend implementation (React/TS)
4. Database schema (if needed)
5. API specifications
6. Tests (unit + integration)
7. Documentation
8. Migration steps
```

---

## CLAUDE-SPECIFIC TIPS

### Use Multiple Messages for Complex Tasks

Instead of one long message:
```
Message 1: [Context + standards + architecture request]
Response 1: [Architecture design]

Message 2: [Paste architecture] Now implement the database schema
Response 2: [Schema + explanation]

Message 3: [Paste schema] Now implement the service layer
Response 3: [Service code]

Message 4: [Paste service] Now implement the API endpoints
Response 4: [API code]
```

### Use Formatting for Clarity

```
## Backend Implementation

### Phase 1: Database Schema
[Your request]

### Phase 2: Data Models  
[Your request]

### Phase 3: Service Layer
[Your request]

### Phase 4: API Layer
[Your request]

### Phase 5: Error Handling
[Your request]

### Phase 6: Tests
[Your request]
```

### Ask for Explanation

```
"Here's my implementation:
[code]

Explain:
1. Why did you make these architectural choices?
2. How does this handle [specific requirement]?
3. What are potential failure modes?
4. How would we scale this to [size]?
5. What tests would you recommend?"
```

---

## GITHUB COPILOT-SPECIFIC TIPS

### Context Window Management

Copilot has limited context. To maximize effectiveness:

1. **Keep files focused** - One responsibility per file
2. **Use clear names** - Descriptive file/variable names
3. **Add comments** - Explain intent clearly
4. **Reference standards** - Link to patterns in codebase

Example:

```rust
// Following database.rs repository pattern
// See: src/infrastructure/database.rs for similar implementation
// Authentication follows auth_service.rs patterns

pub async fn get_website(id: i32, db: &Database) -> Result<Website, DbError> {
    // Implementation
}
```

### Command Usage

**Command: `/explain`**
```
/explain [highlight code]
"Explain this function and how to test it"
```

**Command: `/fix`**
```
/fix [highlight code]
"Fix security issues and add error handling"
```

**Command: `/tests`**
```
/tests [highlight function]
"Write unit tests with 100% coverage"
```

**Command: `/doc`**
```
/doc [highlight code]
"Add comprehensive documentation with examples"
```

### Inline Chat Shortcuts

In VSCode, use keyboard shortcuts:
- `Ctrl+I` or `Cmd+I` - Open inline chat
- Type your request
- Review and accept/modify

Example:
```
Ctrl+I
"Add error handling using Result type pattern from this codebase"
```

---

## COMMON MISTAKES TO AVOID

### ❌ DON'T:

1. **Ask without context**
   ```
   "Write a provisioning function"  ❌
   ```
   Instead:
   ```
   [Paste PROVISIONING AUTOMATION section]
   "Write provisioning function that..." ✓
   ```

2. **Accept first suggestion**
   ```
   Copilot generates code → Accept immediately ❌
   ```
   Instead:
   ```
   Copilot generates code → Review checklist → Ask for fixes ✓
   ```

3. **Skip testing**
   ```
   "Generate function"  ❌
   ```
   Instead:
   ```
   "Generate function WITH unit + integration tests" ✓
   ```

4. **Ignore security**
   ```
   "Implement login" ❌
   ```
   Instead:
   ```
   [Paste AUTHENTICATION section]
   "Implement login using Argon2, JWT, rate limiting..." ✓
   ```

5. **Use outdated versions**
   ```
   "Update dependencies to latest" ❌
   ```
   Instead:
   ```
   [Paste TECHNOLOGY STACK]
   "Update all dependencies to versions specified..." ✓
   ```

---

## VALIDATION CHECKLIST

After AI generates code, verify:

### Security ✓
- [ ] Input validation present
- [ ] Authentication/authorization checks
- [ ] No hardcoded secrets
- [ ] Proper error handling (no stack traces)
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)

### Code Quality ✓
- [ ] Follows style guide (formatting, naming)
- [ ] No code duplication
- [ ] Clear error types
- [ ] Comprehensive comments
- [ ] No compiler warnings

### Testing ✓
- [ ] Unit tests included
- [ ] Error cases tested
- [ ] Edge cases covered
- [ ] Test coverage > 80%
- [ ] Tests pass locally

### Performance ✓
- [ ] No N+1 queries
- [ ] Indexes used
- [ ] Efficient algorithms
- [ ] Caching considered
- [ ] No unnecessary allocations

### Documentation ✓
- [ ] Function documented
- [ ] Parameters explained
- [ ] Returns documented
- [ ] Errors documented
- [ ] Examples provided

### Maintainability ✓
- [ ] Follows project patterns
- [ ] Matches existing code style
- [ ] Uses established abstractions
- [ ] Clear variable names
- [ ] Logical structure

---

## ESCALATION PATH

If AI-generated code doesn't meet standards:

1. **First Try:** Ask Copilot to fix specific issues
   ```
   "Add comprehensive error handling"
   "Add unit tests with 100% coverage"
   "Add security validation for inputs"
   ```

2. **Second Try:** Provide more specific constraints
   ```
   [Paste relevant section from standards]
   "Regenerate following these specific requirements..."
   ```

3. **Third Try:** Use Claude for architectural guidance
   ```
   "This implementation doesn't work. Let's discuss:
   1. What's the issue?
   2. What's the right approach?
   3. Show me the corrected implementation"
   ```

4. **Final Try:** Manual implementation + review
   - Write code yourself
   - Ask AI to review against standards

---

## BEST PRACTICES SUMMARY

✅ **DO:**
- Provide context from standards document
- Be specific about requirements
- Request tests along with code
- Review all generated code
- Ask for explanations
- Request security review
- Use modern AI features (Claude's long context)
- Chain complex tasks into phases
- Validate against checklists

❌ **DON'T:**
- Copy/paste without review
- Use AI code as-is in production
- Ignore security guidelines
- Skip testing
- Use outdated tech stack
- Accept first suggestion
- Deploy without staging
- Ignore error handling
- Store secrets in code
- Use without attribution/review

---

## EXAMPLE: END-TO-END WORKFLOW

### Step 1: Create Context File
Save system prompts to `.copilot/context.md` in project root

### Step 2: Plan Feature with Claude
```
Claude: [Full context]
"I need to implement user billing system. 
Here are the requirements: [paste user stories]
What's the architecture?"

Claude responds with architecture design
```

### Step 3: Implement with Copilot
```
// In VSCode, write comments
// Create invoice and charge user payment
// Handle payment failures with retry logic
// Log all billing events
// Return structured response

#[post("/api/billing/charge")]
async fn charge_user(...
```

Copilot fills in implementation following standards

### Step 4: Review with Claude
```
Claude: [Paste generated code]
"Review this for:
1. Security vulnerabilities
2. Test coverage
3. Error handling
4. Performance

What should we improve?"
```

### Step 5: Refine with Copilot
```
"@copilot Add unit tests, improve error handling, 
add rate limiting middleware"
```

### Step 6: Deploy
- Verify all checks pass
- Deploy to staging
- Test thoroughly
- Deploy to production
- Monitor closely

---

**Remember:** AI is a tool for productivity, not a replacement for critical thinking.  
Always review generated code, verify it meets standards, and take responsibility for what you deploy.

---

**For Questions:**  
Email: tech-leads@example.com  
Slack: #development-standards  
Wiki: https://wiki.internal.example.com/ai-development-standards

---

**Version:** 1.0  
**Last Updated:** November 2, 2025  
**Next Review:** May 2, 2026
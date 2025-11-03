# AI-Assisted Development Documentation Index
## Complete System Prompts for Claude & GitHub Copilot

**Project:** Next-Generation Hosting Control Panel  
**Date:** November 2, 2025  
**Audience:** Development Team, AI Assistants (Claude & GitHub Copilot)  

---

## DOCUMENT STRUCTURE

This documentation package contains comprehensive instructions for using Claude AI and GitHub Copilot to develop enterprise-grade code for the hosting control panel project.

### Included Documents

1. **ai-dev-system-prompts.md** (Main Document - 15+ sections)
   - Primary development principles
   - Technology stack requirements (Rust, React, PostgreSQL)
   - Security standards (comprehensive)
   - Code quality standards
   - Architecture & design patterns
   - Development workflow
   - Testing requirements
   - Deployment & DevOps
   - API development standards
   - Database standards
   - Frontend standards
   - Infrastructure & SysAdmin standards
   - Documentation requirements
   - Code review checklist
   - Emergency procedures

2. **ai-prompting-guide.md** (Usage Guide)
   - How to use system prompts with Claude
   - How to use system prompts with GitHub Copilot
   - Prompt engineering tips
   - Workflow integration
   - Security-focused prompting templates
   - Performance-focused prompting templates
   - Task-specific prompts
   - Claude-specific tips
   - GitHub Copilot-specific tips
   - Common mistakes to avoid
   - Validation checklist
   - Example end-to-end workflow

3. **quick-reference-card.md** (Cheat Sheet)
   - Quick prompting templates
   - Tech stack checklist
   - Security checklist
   - Code quality checklist
   - Testing pyramid
   - Error handling pattern
   - Authentication pattern
   - Database pattern
   - API response format
   - Git workflow
   - Deployment checklist
   - Performance targets
   - Common gotchas
   - Keyboard shortcuts
   - Quick links

---

## QUICK START (5 MINUTES)

### For Claude

1. Copy `ai-dev-system-prompts.md` content
2. Paste into Claude conversation
3. Ask specific question about your feature
4. Include requirements from project scope

**Example:**
```
[Paste ai-dev-system-prompts.md]

I need to implement a user authentication endpoint in Rust.
Requirements:
- Email and password login
- Argon2 password hashing
- JWT tokens
- Rate limiting
- Full test coverage

Generate production-ready code with all best practices.
```

### For GitHub Copilot

1. Create `.copilot/context.md` in project root
2. Paste `ai-dev-system-prompts.md` content
3. Reference in prompts with `@` mentions
4. Use keyboard shortcuts (Ctrl+I for inline chat)

**Example:**
```typescript
// Implement user authentication following:
// @workspace standards in .copilot/context.md
// - Argon2 password hashing
// - JWT token generation
// - Rate limiting middleware
// - Comprehensive error handling

async function authenticateUser(email: string, password: string) {
    // Copilot generates implementation
}
```

---

## DOCUMENT USAGE GUIDE

### Section 1: Primary Development Principles
**When to reference:**
- Starting new feature development
- Making architectural decisions
- Evaluating technology choices
- Setting team standards

**For Claude:** "I'm making a decision about [X]. Following section 1.3 decision framework..."

**For Copilot:** Add comment: `// Following primary development principles from context`

### Section 2: Technology Stack Requirements
**When to reference:**
- Choosing libraries/frameworks
- Updating dependencies
- Onboarding new developers
- Code reviews for tech choices

**For Claude:** "Update all dependencies to match section 2 stack requirements..."

**For Copilot:** Use for autocomplete suggestions - Copilot will match your declared stack

### Section 3: Security Standards
**When to reference:**
- Every API endpoint development
- Database access code
- Authentication/authorization implementation
- Security review of code
- Vulnerability response

**For Claude:** "Review this code against section 3 security standards for:"
- Input validation
- SQL injection prevention
- XSS prevention
- CSRF protection
- Error handling

**For Copilot:** Use `/fix` command after pasting code

### Section 4: Code Quality Standards
**When to reference:**
- Code style enforcement
- Documentation writing
- Error handling patterns
- Testing strategy

**For Claude:** "Check this code against section 4 standards - does it meet:"
- Code style requirements?
- Documentation standards?
- Error handling requirements?

**For Copilot:** Use for style suggestions and documentation generation

### Section 5: Architecture & Design Patterns
**When to reference:**
- System design decisions
- Component structure
- Layer separation
- Design pattern selection

**For Claude:** "Design [component] following section 5 patterns. Should I use:"
- Repository pattern for data access?
- Service layer for business logic?
- Middleware for cross-cutting concerns?

**For Copilot:** Copilot will generate code matching your architecture

### Section 6: Development Workflow & Git
**When to reference:**
- Committing code
- Creating pull requests
- Code review process
- Branch management

**For Claude:** "Review my Git workflow. Did I follow section 6 standards?"

**For Copilot:** Reference in commit messages and PR descriptions

### Section 7: Testing & QA
**When to reference:**
- Writing test cases
- Setting coverage targets
- Performance testing
- Security testing

**For Claude:** "Write comprehensive tests following section 7 test pyramid for:"
- Unit tests (70%)
- Integration tests (20%)
- E2E tests (10%)

**For Copilot:** Use `/tests` command to generate test code

### Section 8: Deployment & DevOps
**When to reference:**
- Deployment planning
- Infrastructure setup
- Monitoring configuration
- Disaster recovery

**For Claude:** "Plan deployment of [service] following section 8 checklist..."

**For Copilot:** Reference in deployment scripts and Terraform code

### Section 9: API Development Standards
**When to reference:**
- Building REST API endpoints
- Response format specification
- API documentation
- Versioning strategy

**For Claude:** "Design API for [feature] following section 9 standards with:"
- RESTful conventions
- Standard response format
- Error response format
- Pagination

**For Copilot:** Copilot generates endpoints matching specified format

### Section 10: Database Standards
**When to reference:**
- Schema design
- Query optimization
- Migration creation
- Index planning

**For Claude:** "Design database schema following section 10 standards including:"
- Proper normalization
- Constraints and checks
- Audit fields
- Indexes

**For Copilot:** Use for SQL generation and optimization

### Section 11: Frontend Standards
**When to reference:**
- React component development
- TypeScript configuration
- State management
- Form handling

**For Claude:** "Implement React component following section 11 standards with:"
- TypeScript strict mode
- Hooks best practices
- Redux patterns
- Full type safety

**For Copilot:** Copilot generates React code matching your stack

### Section 12: Infrastructure & SysAdmin
**When to reference:**
- Server configuration
- Security hardening
- Monitoring setup
- Disaster recovery

**For Claude:** "Configure [service] following section 12 standards for:"
- Security hardening
- Monitoring and alerting
- Backup strategy

**For Copilot:** Reference in IaC (Terraform, Docker) files

### Section 13: Documentation Requirements
**When to reference:**
- Writing code comments
- API documentation
- Runbook creation
- Architecture documentation

**For Claude:** "Write documentation for [API endpoint] following section 13 standards"

**For Copilot:** Use `/doc` command for automatic documentation

### Section 14: Code Review Checklist
**When to reference:**
- Reviewing code before committing
- PR review process
- Quality gate verification

**For Claude:** "Review this code against the section 14 checklist for:"
- Security vulnerabilities
- Code quality
- Test coverage
- Performance
- Documentation

**For Copilot:** Use during code review to catch issues

### Section 15: Emergency Procedures
**When to reference:**
- Security incidents
- Service outages
- Performance degradation
- Database problems

**For Claude:** "We have [incident]. According to section 15 procedures, what should we do?"

---

## WORKFLOW INTEGRATION

### Claude + GitHub Copilot Workflow

**Phase 1: Design (Claude)**
```
[Paste ai-dev-system-prompts.md sections 1,5]
"Design the billing system architecture"
→ Get architectural guidance
```

**Phase 2: Implementation (GitHub Copilot)**
```
// Create files following Claude's design
// Copilot provides implementation suggestions
// Based on project's declared tech stack
```

**Phase 3: Review (Claude)**
```
[Paste generated code]
"Review against section 4 (quality) and section 3 (security)"
→ Get improvement suggestions
```

**Phase 4: Refinement (GitHub Copilot)**
```
// Use /fix command
// Use @workspace references
// Get improved implementation
```

**Phase 5: Deploy (Both)**
```
Claude: "Review deployment plan against section 8"
Copilot: "Generate deployment script"
```

---

## REFERENCE BY TECHNOLOGY

### For Rust Development

**Key Sections:**
- Section 2.1: Backend Stack (Rust, Actix-web, Tokio)
- Section 3: Security Standards (all Rust-specific practices)
- Section 4: Code Quality (Rust patterns)
- Section 5: Architecture & Design Patterns
- Section 9: API Development (with Actix examples)
- Section 10: Database (Diesel ORM)

**Claude Prompt:**
```
[Paste sections 2.1, 3, 4, 5, 9]
"Implement [feature] in Rust/Actix-web following all standards"
```

**Copilot:** Start with comments describing intent, let Copilot generate

### For React/TypeScript Development

**Key Sections:**
- Section 2.2: Frontend Stack (React 18, TypeScript 5)
- Section 4: Code Quality (TypeScript strict mode)
- Section 5: Architecture & Design Patterns
- Section 11: Frontend Development Standards
- Section 7: Testing (React Testing Library)

**Claude Prompt:**
```
[Paste sections 2.2, 4, 11]
"Build [component] in React/TypeScript following standards"
```

**Copilot:** Works excellently with TypeScript for type safety

### For Database Development

**Key Sections:**
- Section 10: Database Standards
- Section 3: Security (data protection)
- Section 8: Deployment (migrations, backups)

**Claude Prompt:**
```
[Paste sections 10, 3]
"Design [table] schema following PostgreSQL best practices"
```

**Copilot:** Excellent at SQL generation with comments

### For DevOps/Infrastructure

**Key Sections:**
- Section 8: Deployment & DevOps
- Section 12: Infrastructure & SysAdmin
- Section 2.3: DevOps & Infrastructure Stack

**Claude Prompt:**
```
[Paste sections 8, 12, 2.3]
"Create deployment pipeline for [service]"
```

**Copilot:** Great for Terraform, Docker, CI/CD configuration

---

## COMMON SCENARIOS

### Scenario 1: Build New Feature

**Claude:**
1. Paste sections: 1, 2, 3, 4, 5
2. Ask: "Design new feature [X]"
3. Get architecture, tech choices, security approach

**Copilot:**
1. Create files following architecture
2. Write comments describing intent
3. Let Copilot generate implementation
4. Review with validation checklist

**Claude (Review):**
1. Paste generated code
2. Reference section 4 (quality) + section 3 (security)
3. Request improvements

### Scenario 2: Security Audit

**Claude:**
1. Paste section 3 (Security Standards)
2. Paste your code
3. Ask: "Find security vulnerabilities in this code"

**Outcome:**
- List of vulnerabilities
- Severity ratings
- Remediation code
- Testing approach

### Scenario 3: Performance Optimization

**Claude:**
1. Paste section: 8 (Performance benchmarking)
2. Describe performance problem
3. Ask: "Optimize this following performance standards"

**Outcome:**
- Root cause analysis
- Optimization strategies
- Performance testing approach
- Optimized implementation

### Scenario 4: Code Review Before PR

**Claude:**
1. Paste section 14 (Code Review Checklist)
2. Paste your code
3. Ask: "Review against this checklist"

**Outcome:**
- Issues found
- Priority (blocking vs nice-to-have)
- How to fix each issue

### Scenario 5: Deployment Planning

**Claude:**
1. Paste section 8 (Deployment & DevOps)
2. Describe what you're deploying
3. Ask: "Create deployment plan"

**Outcome:**
- Pre-deployment checklist
- Deployment steps
- Monitoring approach
- Rollback procedure

---

## VALIDATION CHECKLIST (BEFORE DEPLOYING)

Use this checklist after AI generates code:

### Security ✓
- [ ] Paste code into Claude with section 3
- [ ] Ask: "Find security issues"
- [ ] Fix all found issues
- [ ] Verify input validation
- [ ] Verify auth/authz checks
- [ ] Verify no secrets leaked

### Quality ✓
- [ ] Code runs without errors
- [ ] `cargo fmt` passes (Rust)
- [ ] `cargo clippy` passes (Rust)
- [ ] `prettier` passes (TypeScript)
- [ ] ESLint passes (TypeScript)
- [ ] No compiler warnings

### Testing ✓
- [ ] Unit tests written
- [ ] Integration tests written
- [ ] All tests pass
- [ ] Coverage > 80%
- [ ] Error cases tested
- [ ] Edge cases tested

### Performance ✓
- [ ] No N+1 queries
- [ ] Indexes used
- [ ] Efficient algorithms
- [ ] Caching considered
- [ ] Load tested

### Documentation ✓
- [ ] Function documented
- [ ] API documented
- [ ] Examples provided
- [ ] Error codes documented

---

## TEAM SETUP INSTRUCTIONS

### For Solo Developers

1. Save all 3 documents locally
2. When starting feature: Copy main section to Claude
3. Use Copilot for inline code generation
4. Validate against quick-reference-card

### For Small Teams (2-5 developers)

1. **Repository Setup:**
   - Add `.copilot/context.md` to project (main prompts file)
   - Add `docs/development-standards.md` (for reference)
   - Add `docs/quick-reference.md` (cheat sheet)

2. **Onboarding:**
   - Each dev reads main document
   - Keep quick-reference-card at desk
   - Review section 6 (Git workflow)

3. **Code Review:**
   - Use section 14 (Code Review Checklist)
   - Require AI review before human review
   - Use Claude for architecture review

### For Larger Teams (5+ developers)

1. **Standardization:**
   - Store in Wiki/Confluence
   - Update quarterly with lessons learned
   - Maintain version history

2. **CI/CD Integration:**
   - Add linting checks (cargo clippy, eslint)
   - Add security scanning (SAST)
   - Add test coverage gates (>80%)

3. **Developer Onboarding:**
   - Week 1: Read main document
   - Week 2: Pair program with senior dev
   - Week 3: First PR with senior review
   - Week 4: Solo development with code review

4. **Training Sessions:**
   - Monthly: Security deep-dive (section 3)
   - Monthly: Performance optimization (section 8)
   - Quarterly: Architecture decisions (section 5)

---

## METRICS TO TRACK

### Code Quality Metrics
- Test coverage: Target > 80%
- Code review time: Target < 24 hours
- Build time: Target < 5 minutes
- Deploy time: Target < 10 minutes

### Security Metrics
- Vulnerabilities found: Target 0 critical/high
- Security review rate: 100% of PRs
- Incident response time: < 5 minutes (critical)
- Patch time: < 48 hours (critical)

### Performance Metrics
- API latency: Target < 100ms (p95)
- Throughput: Target > 10,000 req/sec
- Database query time: Target < 50ms
- Error rate: Target < 0.1%

### Team Metrics
- Developer productivity: Stories completed/sprint
- Code review efficiency: Avg comments per PR
- Onboarding time: Days to first solo PR
- Knowledge sharing: Cross-team code reviews

---

## TROUBLESHOOTING

### "AI Generated Code That Doesn't Compile"

**Solution:**
1. Copy error message
2. Ask Claude: "Fix this compilation error"
3. Or use Copilot `/fix` command

### "AI Code Doesn't Follow Our Standards"

**Solution:**
1. Identify which standard violated (use quick-reference-card)
2. Paste relevant section to Claude
3. Ask: "This code doesn't follow this standard. Fix it."

### "Generated Code Passes Tests But Feels Wrong"

**Solution:**
1. Paste code to Claude
2. Ask: "Review this against section 5 (architecture). Does it follow our patterns?"
3. Get architectural review

### "Performance is Slow"

**Solution:**
1. Describe the issue
2. Paste code to Claude
3. Reference section 8 (Performance)
4. Ask: "Optimize this"

### "Security Vulnerability Found"

**Solution:**
1. Report issue to Claude with section 3 (Security)
2. Get fix recommendations
3. Claude generates patch code
4. Deploy after testing

---

## MAINTENANCE & UPDATES

### Document Update Schedule
- **Quarterly:** Review and update based on lessons learned
- **Annually:** Major review of all standards
- **As needed:** Security updates (immediately)

### Version Control
- Keep version number in each document
- Document date of last update
- Maintain changelog in README
- Archive previous versions

### Team Feedback
- Monthly feedback collection from developers
- Quarterly standards review meeting
- Annual comprehensive audit
- Suggestion process for improvements

---

## FINAL CHECKLIST

Before using these documents in production:

- [ ] All team members have read main document
- [ ] Quick-reference-card printed and at desk
- [ ] Copilot system prompts configured in VSCode
- [ ] Claude context ready for pasting
- [ ] CI/CD checks configured
- [ ] Code review process using section 14
- [ ] Emergency procedures understood (section 15)
- [ ] Git workflow practiced (section 6)
- [ ] Security practices reinforced (section 3)
- [ ] First PR review completed by senior dev

---

## CONTACT & SUPPORT

**Technical Lead:** [name] - [email]  
**Security Lead:** [name] - [email]  
**DevOps Lead:** [name] - [email]  
**Team Slack:** #development-standards  

**Questions About:**
- Architecture/Design → Technical Lead
- Security issues → Security Lead
- Deployment/DevOps → DevOps Lead
- Anything else → Slack channel

---

## DOCUMENT SUMMARY

| Document | Purpose | Length | Review |
|----------|---------|--------|--------|
| ai-dev-system-prompts.md | Main standards reference | 25+ pages | Quarterly |
| ai-prompting-guide.md | How to use the prompts | 20+ pages | Annually |
| quick-reference-card.md | Cheat sheet for desk | 4 pages | Print & laminate |
| INDEX (this document) | Navigation guide | 5+ pages | Annually |

---

## VERSION HISTORY

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | Nov 2, 2025 | Initial release |
| 1.1 | [TBD] | [TBD] |

---

**You are now ready to use AI-assisted development with enterprise-grade standards.**

✨ Quality over speed  
🔒 Security first  
🧪 Test everything  
📚 Document well  
👥 Review code  
🚀 Deploy confidently  

**Good luck with your hosting control panel project!**

---

**Last Updated:** November 2, 2025  
**Next Review:** May 2, 2026  
**Maintained By:** Development Team  
**License:** Internal Use Only
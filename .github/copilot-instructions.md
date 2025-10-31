# Copilot Instructions for DASHBOARD-v2

## Project Overview
This is a dashboard web application project currently in early setup phase. The codebase follows security-first practices with comprehensive CodeQL analysis and strict file security conventions.

## Architecture & Technology Stack
- **Frontend**: HTML-based dashboard (technology stack to be determined)
- **Security**: CodeQL-enabled with multi-language support (JavaScript/TypeScript, Python ready)
- **Environment**: Windows development environment with Git Bash as default terminal
- **Repository**: GitHub with `master` as working branch, `main` as default branch

## Development Environment Setup
- Use Git Bash terminal (configured in `.vscode/settings.json`)
- All terminal commands should be bash-compatible
- VSCode is the primary development environment

## Security & Compliance Patterns
- **CodeQL Integration**: Comprehensive security analysis runs on push/PR to master/main branches
  - Supports JavaScript/TypeScript and Python out of the box
  - Weekly scheduled scans (Mondays 16:44 UTC)
  - Manual workflow dispatch available
- **File Security**: Extensive `.gitignore` covers secrets, API keys, credentials, and environment files
- **Security Policy**: Follow the template in `SECURITY.md` for vulnerability reporting

## Key File Patterns
- **Main Entry**: `index.html` serves as the application entry point
- **Configuration**: Project uses VSCode workspace settings for terminal configuration
- **Security**: Never commit files matching patterns in `.gitignore`, especially:
  - Environment files (`.env*`, `*.env`)
  - Credentials (`credentials.json`, `secrets.json`, `*.key`, `*.pem`)
  - API keys and certificates

## Development Workflows
### Security Analysis
```bash
# Trigger manual CodeQL analysis
gh workflow run "CodeQL Advanced"
```

### Branch Strategy
- **Working branch**: `master`
- **Default branch**: `main` 
- PRs should target appropriate branch based on deployment strategy

## Code Quality Standards
- CodeQL security-and-quality query suite enforced
- All workflows include comprehensive permissions for security scanning
- Build processes should accommodate both interpreted (JS/Python) and compiled languages

## Future Architecture Considerations
The CodeQL configuration suggests this project may expand to include:
- Node.js frontend framework
- Python backend services
- Potential multi-language microservices architecture

## When Adding New Technologies
1. Update CodeQL matrix in `.github/workflows/codeql.yml` to include new languages
2. Uncomment relevant setup steps in the workflow
3. Configure manual build steps for compiled languages
4. Update `.gitignore` for language-specific ignore patterns

## Critical Files to Understand
- `.github/workflows/codeql.yml`: Security analysis configuration and supported languages
- `.gitignore`: Comprehensive security-focused file exclusions
- `.vscode/settings.json`: Development environment terminal configuration
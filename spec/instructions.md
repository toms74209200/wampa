# GitHub Copilot Instructions for Wampa Project

This document provides custom instructions for GitHub Copilot when working with the Wampa project.

## Important Note on Documentation

- All documentation and code comments MUST be written entirely in English for better token efficiency and consistent documentation.

## Important Files to Reference

When working on this project, always reference these key files for context and requirements:

### Technical Requirements and Standards
- `/spec/requirements.md` - Primary technical requirements, architecture design, and implementation guidelines
  - **When to reference**: Before starting new feature design, when deciding implementation approach, when checking coding standards
  - **Key information**: Architecture constraints, dependency restrictions, testing requirements
- `/spec/spec.md` - Detailed specifications and use cases
  - **When to reference**: Before detailed implementation, when understanding use cases, when designing interfaces
  - **Key information**: User scenarios, input/output formats, expected behaviors

### Feature Specifications
- `/features/config_file_handling.feature` - Configuration file handling scenarios
  - **When to reference**: When implementing configuration file related features, creating related tests
  - **Key information**: Acceptance criteria, expected config file processing behavior
- `/features/local_file_monitoring.feature` - Local file monitoring scenarios
  - **When to reference**: When implementing file monitoring features, creating related tests
  - **Key information**: File change detection specifications, output file update requirements
- `/features/remote_file_handling.feature` - Remote file handling scenarios
  - **When to reference**: When implementing remote file processing features, creating related tests
  - **Key information**: Remote file retrieval and monitoring requirements, error handling conditions

### Project Status
- `/TODO.md` - Current development status, issues, and progress tracking
  - **When to reference**: 
    - Before starting work (checking current status)
    - During work (recording progress)
    - When completing work (marking completed items)
    - When discovering issues (recording problems)
  - **What to check/update**: Current work context, error and bug tracking, implementation status

## Development Process and Guidelines

### General Development Workflow
1. Always start with the latest main branch
2. Review and agree on implementation strategy before starting work:
   - Check `/spec/requirements.md` for architectural constraints and technical requirements
   - Review `/spec/spec.md` for relevant use cases and functionality
   - Verify feature specifications in appropriate `.feature` files
   - Consider test size implications (small/medium/large) based on implementation needs
   - Document the proposed implementation strategy
   - Get team agreement on the implementation approach through:
     - Share the documented strategy with team members
     - Discuss potential trade-offs and alternatives
     - Address concerns and incorporate feedback
     - Obtain explicit approval before proceeding
   - **Important Note on Using edit_file Tool**:
     - All edit_file tool executions performed without prior agreement will be rejected
     - Rejected edit_file operations are permanently discarded and cannot be recovered
     - The edit_file tool can only be used with proper agreement or upon specific request

3. Create a feature branch for each task/bugfix
4. Follow TDD approach: test → implementation → refactoring
5. Update `TODO.md` with EVERY operation (see [Project Tracking](#project-tracking))
6. Run `make all` before committing to check for any issues
7. Create a Pull Request after ensuring all checks pass

### Working with Code
1. Use absolute paths for all file operations to avoid execution failures
2. Run commands from the project root when possible
3. Always verify dependency availability before using tools

### Version Control and Commits
1. Always use [gitmoji](https://gitmoji.dev/) for commit messages
   - ✨ (`:sparkles:`) - New features
   - 🐛 (`:bug:`) - Bug fixes
   - ♻️ (`:recycle:`) - Code refactoring
   - 📝 (`:memo:`) - Documentation updates
   - ✅ (`:white_check_mark:`) - Adding or updating tests
   - 🎨 (`:art:`) - Improving code structure/format
2. Format commits as: `emoji Short description in English`
3. **Always verify changes before committing using `git status`**
4. Before committing, run tests and linting to ensure quality

### Testing Strategy
1. Test Types and Size Classifications (refer to [Google Testing Blog](https://testing.googleblog.com/2010/12/test-sizes.html))
   
   **Important Principle**: All unit tests MUST be small tests, while acceptance tests can be medium or large tests depending on their implementation requirements.
   
   - **Small Tests**:
     - For unit testing only
     - Tagged with `//go:build small`
     - Run with `go test -tags=small ./...`
     - Must follow strict Small Test requirements (see below)
     
   - **Medium Tests**:
     - For integration testing and simpler acceptance tests
     - Include acceptance tests that use file system but no external services
     - Located in `tests/acceptance` with appropriate features
     - Tagged with `//go:build medium`
     - Run with `go test -tags=medium ./tests/acceptance/...`
     - Allowed to use file system, concurrency, and reasonable timeouts
     
   - **Large Tests**:
     - For end-to-end acceptance tests with external dependencies
     - Can use network, external services, longer timeouts
     - Tagged with `//go:build large`
     - Run with `go test -tags=large ./tests/acceptance/...`
     - Should be kept to a minimum due to resource requirements

2. Test size Constraints:
    | Feature                  | Small                 | Medium                     | Large                           |
    |--------------------------|-----------------------|----------------------------|---------------------------------|
    | Time limit(seconds)      | 60                    | 300                        | 900+                            |
    | Networks                 | No                    | localhost only             | Yes                             |
    | Database                 | No                    | Yes                        | Yes                             |
    | File system access       | No                    | Discouraged                | Yes                             |
    | Multiple threads         | No                    | Yes                        | Yes                             |
    | Sleep statements         | No                    | Yes                        | Yes                             |
    | System properties        | No                    | Yes                        | Yes                             |

3. Test Coverage Requirements:
   - Overall coverage must be at least 80%
   - Managed through `scripts/coverage_pkgs.txt`
   - Check coverage using `make cover`

4. Acceptance Test Process:
   - Acceptance tests are implemented at appropriate test sizes (medium or large)
   - The test size should match the resource needs of the test scenario
   - Feature files are in `/features` directory with `.feature` extension
   - Test implementations are in `/tests/acceptance`
   - Run acceptance tests from the project root with:
     ```bash
     # For medium tests
     cd /workspaces/wampa/tests/acceptance && go test -tags=medium ./...
     
     # For large tests (if applicable)
     cd /workspaces/wampa/tests/acceptance && go test -tags=large ./...
     ```
   - Features can be filtered by tags in the test files

### Pull Request Process
1. Run `make all` to verify all tests pass and no lint errors exist
2. Push changes to feature branch
3. Create PR using GitHub CLI: `gh pr create --base main --head feature-branch`
4. Ensure PR description includes:
   - Description of changes
   - Reference to requirements implemented
   - Checklist of completed items

## Code Quality Tools

For development tools configuration and usage:
- Refer to `/.golangci.yml` for linter rules and settings
- Refer to `/Makefile` for all development, test, and code quality commands
- Refer to `/scripts/coverage_pkgs.txt` for test coverage configuration

### Formatters and Linters
1. Code formatting: `gofmt` and `goimports`
   - Run via: `make fmt`
   - Standard style: tab indentation, no trailing newlines

2. Linting: `golangci-lint`
   - Configuration in `.golangci.yml`
   - Run via: `make lint`

### Build and Test Tools
For build and test commands, refer to `/Makefile`

## Project Tracking

The `TODO.md` file must be updated with EVERY operation to maintain an accurate record of development status. Update it when:

1. Starting to edit a file (record the target filename)
2. Encountering compilation/test errors (copy the error message)
3. When tests fail (record test name and error message)
4. Finding new issues/concerns (details and mitigation strategy)
5. Completing feature implementation (summary of implementation)

Format the file according to the example in `requirements.md`.

## Definition of Done

Before considering any task complete, ensure:
1. All small tests are passing
2. All acceptance tests (Gherkin scenarios) are passing
3. Test coverage is at least 80%
4. No lint errors are present

These requirements are non-negotiable and must be verified using the appropriate tools:
- Run `make test` to verify small tests
- Run acceptance tests from the appropriate directory with correct tags
- Use `make cover` to check coverage
- Run `make lint` to verify code style and quality
- Finally, run `make all` to perform all checks
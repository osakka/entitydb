# Documentation Maintenance Guidelines

> **Category**: Development | **Target Audience**: Maintainers | **Technical Level**: Advanced
> **Version**: v2.32.0-dev | **Last Updated**: 2025-06-15 | **Status**: AUTHORITATIVE

This document defines the standards, processes, and guidelines for maintaining the EntityDB documentation library to ensure consistency, accuracy, and professional quality.

## 📋 Documentation Standards

### 1. File Naming Conventions

#### File Names
```
01-topic-name.md          # Numbered for sequence
README.md                 # Section overview
CHANGELOG.md              # Version history
```

#### Directory Structure
```
docs/
├── getting-started/      # New user onboarding
├── architecture/         # Technical system design
├── api-reference/        # Complete API documentation
├── user-guide/          # Task-oriented guides
├── admin-guide/         # Administrative documentation
├── developer-guide/     # Development and contributing
└── reference/           # Technical references and troubleshooting
```

### 2. Document Format Standards

#### Document Header Format
```markdown
# Document Title (Title Case, No Emojis)

> **Category**: Document Type | **Target Audience**: Users | **Technical Level**: Beginner/Intermediate/Advanced
> **Version**: vX.Y.Z | **Last Updated**: YYYY-MM-DD | **Status**: DRAFT/REVIEW/AUTHORITATIVE

Brief description of document purpose and scope.
```

#### Section Headers
- **H1 (`#`)**: Document title only
- **H2 (`##`)**: Major sections
- **H3 (`###`)**: Subsections
- **H4 (`####`)**: Sub-subsections (avoid deeper nesting)

#### Code Block Standards
```markdown
# Shell commands - always use 'bash'
```bash
curl -X GET https://localhost:8085/api/v1/entities/list
```

# Go code - use 'go'
```go
func Example() {
    // Implementation
}
```

# Configuration - use appropriate format
```json
{"key": "value"}
```

```yaml
key: value
```
```

#### Cross-Reference Format
```markdown
## See Also

- [Related Topic](./relative-path.md) - Brief description
- [External Resource](../other-section/file.md) - Brief description
```

### 3. Content Guidelines

#### Writing Style
- **Clear and concise**: Avoid unnecessary words
- **Active voice**: Use active voice when possible
- **Consistent terminology**: Use established EntityDB terminology
- **Professional tone**: Formal but accessible

#### Technical Accuracy
- All code examples must be tested and functional
- API documentation must match current implementation
- Version information must be current and accurate
- Cross-references must be valid and up-to-date

## 🔄 Review and Update Process

### 1. Quarterly Documentation Review

**Schedule**: Every 3 months (March, June, September, December)

**Checklist**:
- [ ] Update version references to current release
- [ ] Verify all cross-references are valid
- [ ] Test all code examples and API calls
- [ ] Review technical accuracy against codebase
- [ ] Update performance metrics and benchmarks
- [ ] Check for broken external links
- [ ] Validate installation and setup instructions

### 2. Release Documentation Updates

**When**: With each EntityDB release

**Required Updates**:
- [ ] Update CHANGELOG.md with new features and changes
- [ ] Update version references in all documents
- [ ] Add new API endpoints and deprecate old ones
- [ ] Update system requirements if changed
- [ ] Review and update performance characteristics
- [ ] Update screenshots and UI documentation

### 3. Content Validation Process

#### Before Publishing
1. **Technical Review**: Verify technical accuracy
2. **Editorial Review**: Check grammar, style, and consistency
3. **Link Validation**: Ensure all cross-references work
4. **Format Check**: Verify adherence to standards

#### Documentation Status Levels
- **DRAFT**: Work in progress, not ready for users
- **REVIEW**: Ready for technical and editorial review
- **AUTHORITATIVE**: Approved, current, and accurate

### 4. Change Management

#### Major Changes (Breaking Changes, New Features)
- Require technical review by maintainer
- Update affected cross-references
- Add changelog entry
- Update version metadata

#### Minor Changes (Corrections, Clarifications)
- Can be made directly
- Update "Last Updated" metadata
- Consider impact on related documents

## 🛠️ Maintenance Tools and Automation

### 1. Link Validation
```bash
# Check for broken internal links
find docs/ -name "*.md" -exec grep -l "(\.\./\|(\./\)" {} \; | \
    xargs -I {} bash -c 'echo "Checking: {}"; grep -n "(\.\./\|(\./\)" "{}"'
```

### 2. Format Validation
```bash
# Check for missing metadata
find docs/ -name "*.md" ! -name "README.md" -exec grep -L "Category\|Target Audience" {} \;

# Check for inconsistent headers  
find docs/ -name "*.md" -exec grep -n "^# " {} \; | grep -v "# [A-Z]"
```

### 3. Content Validation
```bash
# Find outdated version references
grep -r "v2\.[0-9][0-9]" docs/ | grep -v "v2.32"

# Find potential SQLite references (should be eliminated)
grep -ri "sqlite\|\.db" docs/ --exclude-dir=archive
```

## 📊 Quality Metrics

### Documentation Health Indicators

#### Completeness
- [ ] All API endpoints documented
- [ ] All major features covered
- [ ] Installation and setup complete
- [ ] Troubleshooting section comprehensive

#### Accuracy
- [ ] Code examples tested and working
- [ ] Version information current
- [ ] Performance metrics up-to-date
- [ ] Architecture diagrams accurate

#### Usability
- [ ] Clear navigation structure
- [ ] Consistent cross-references
- [ ] Appropriate difficulty progression
- [ ] Comprehensive search capability

### Monthly Health Check

**Metrics to Track**:
- Number of broken links
- Percentage of documents with current metadata
- Age of oldest "Last Updated" timestamp
- Number of DRAFT vs AUTHORITATIVE documents

**Quality Thresholds**:
- Broken links: 0 in active documentation
- Current metadata: 100% of active documents
- Stale content: No documents older than 6 months
- Draft ratio: <10% of total documents

## 🔧 Troubleshooting Documentation Issues

### Common Problems and Solutions

#### 1. Broken Cross-References
**Symptoms**: Links that don't resolve or return 404
**Solution**: Update paths to match current directory structure
```bash
# Find and fix common patterns
sed -i 's|../[0-9][0-9]-\([^/]*\)/|../\1/|g' docs/**/*.md
```

#### 2. Outdated Code Examples
**Symptoms**: API calls that fail or return unexpected results
**Solution**: Test all examples against current API
```bash
# Test API endpoints
curl -k https://localhost:8085/health
curl -k https://localhost:8085/api/v1/status
```

#### 3. Inconsistent Formatting
**Symptoms**: Mixed header styles, missing metadata
**Solution**: Apply standardized templates
```bash
# Check for formatting issues
grep -n "^#" docs/**/*.md | grep -v "^# [A-Z]"
```

#### 4. Version Drift
**Symptoms**: References to old versions
**Solution**: Global update with verification
```bash
# Update version references (with manual verification)
grep -r "v2\.[0-9][0-9]" docs/ | grep -v "v2.32"
```

## 📚 Resources and References

### Style Guides
- [EntityDB Terminology Guide](./reference/terminology.md)
- [API Documentation Standards](./developer-guide/04-api-documentation.md)
- [Markdown Style Guide](./developer-guide/markdown-style.md)

### Tools
- **Markdown Linter**: markdownlint for consistency
- **Link Checker**: markdown-link-check for validation  
- **Spell Check**: aspell for proofreading
- **Grammar Check**: LanguageTool for grammar

### External Resources
- [Markdown Guide](https://www.markdownguide.org/)
- [GitHub Flavored Markdown](https://github.github.com/gfm/)
- [Technical Writing Best Practices](https://developers.google.com/tech-writing)

---

## Implementation Notes

This maintenance guide was created as part of the v2.32.0-dev documentation overhaul that:
- Eliminated 200+ duplicate files
- Fixed 20+ broken cross-references
- Standardized format across all documents
- Created authoritative single source of truth

**Next Review Due**: September 2025

## See Also

- [Contributing Guide](./developer-guide/01-contributing.md) - Development contribution process
- [Git Workflow](./developer-guide/02-git-workflow.md) - Version control standards
- [Configuration Management](./developer-guide/05-configuration-alignment-action-plan.md) - Configuration standards
# Quick Documentation Maintenance Checklist

> **Category**: Development | **Target Audience**: Maintainers | **Technical Level**: Intermediate
> **Version**: v2.32.0-dev | **Last Updated**: 2025-06-15 | **Status**: AUTHORITATIVE

Fast reference for common documentation maintenance tasks.

## 🚀 Pre-Release Checklist (5 minutes)

### Version Update
- [ ] Update version in main README.md
- [ ] Update CHANGELOG.md with new features
- [ ] Update version metadata in key documents
- [ ] Check for hardcoded version references

```bash
# Quick version check
grep -r "v2\.[0-9][0-9]" docs/ | grep -v "v2.32" | head -10
```

### Link Validation
- [ ] Test major navigation paths
- [ ] Verify API endpoint examples work
- [ ] Check that installation instructions are current

```bash
# Quick link check
curl -k https://localhost:8085/health
curl -k https://localhost:8085/api/v1/status
```

## 🔍 Monthly Health Check (15 minutes)

### Content Accuracy
- [ ] Test API examples in getting-started
- [ ] Verify installation steps work
- [ ] Check that troubleshooting solutions are current
- [ ] Review recent issues for documentation gaps

### Format Consistency
- [ ] Check for missing metadata headers
- [ ] Verify cross-reference consistency
- [ ] Ensure code blocks have language tags

```bash
# Format validation
find docs/ -name "*.md" ! -name "README.md" -exec grep -L "Category\|Target Audience" {} \;
```

### Broken Links
- [ ] Test internal navigation
- [ ] Verify external links work
- [ ] Check for old archive references

```bash
# Find old archive references
grep -r "../[0-9][0-9]-" docs/ | grep -v archive/
```

## 🛠️ Quarterly Deep Review (45 minutes)

### Technical Accuracy
- [ ] Review all API endpoint documentation
- [ ] Test all code examples
- [ ] Verify system requirements
- [ ] Update performance metrics
- [ ] Check architecture accuracy

### Content Organization
- [ ] Review navigation structure
- [ ] Assess document flow and progression
- [ ] Identify gaps in coverage
- [ ] Consolidate overlapping content

### Quality Assurance
- [ ] Proofread for grammar and clarity
- [ ] Ensure consistent terminology
- [ ] Verify document status levels
- [ ] Update "Last Updated" timestamps

## 🚨 Emergency Fixes (1 minute)

### Critical Issues
If users report documentation problems:

1. **Broken Link**: Fix immediately
```bash
# Quick fix pattern
sed -i 's|old-path|new-path|g' docs/path/to/file.md
```

2. **Wrong API Example**: Test and correct
```bash
# Test the API call first
curl -k -X GET https://localhost:8085/api/v1/entities/list
```

3. **Outdated Information**: Mark as outdated until fixed
```markdown
> ⚠️ **OUTDATED**: This section is being updated for v2.32.0
```

## 📊 Quality Gates

### Before Any Release
- **Zero broken internal links** in active documentation
- **All code examples tested** and working
- **Version information current** across all documents
- **Installation guide verified** on clean system

### Health Indicators
- 🟢 **Good**: All checks pass, <5% draft documents
- 🟡 **Warning**: 1-2 broken links, some stale content
- 🔴 **Critical**: >3 broken links, outdated API examples

## 🔧 Common Quick Fixes

### Update Cross-References
```bash
# Fix old archive structure references
find docs/ -name "*.md" -not -path "*/archive/*" -exec sed -i 's|../[0-9][0-9]-getting-started/|../getting-started/|g' {} \;
find docs/ -name "*.md" -not -path "*/archive/*" -exec sed -i 's|../[0-9][0-9]-architecture/|../architecture/|g' {} \;
```

### Add Missing Metadata
```bash
# Find files missing metadata
find docs/ -name "*.md" ! -name "README.md" -exec grep -L "Target Audience" {} \;
```

### Update Version References
```bash
# Update version (manual verification recommended)
grep -l "v2\.3[0-1]" docs/**/*.md | xargs sed -i 's/v2\.3[0-1]/v2.32/g'
```

## 📞 Quick Support

### Documentation Issues
- **File a bug**: [Documentation Issues](https://git.home.arpa/itdlabs/entitydb/issues?labels=documentation)
- **Quick fix needed**: Contact maintainer directly
- **Major overhaul**: Schedule quarterly review

### Resources
- [Full Maintenance Guide](./DOCUMENTATION_MAINTENANCE.md)
- [Contributing Guidelines](./developer-guide/01-contributing.md)
- [Git Workflow](./developer-guide/02-git-workflow.md)

---

*Last comprehensive review: June 2025 | Next review due: September 2025*
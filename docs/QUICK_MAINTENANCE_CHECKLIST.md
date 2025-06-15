# EntityDB Documentation Quick Maintenance Checklist

> **Fast Reference Guide** for daily documentation maintenance tasks.  
> Keep EntityDB documentation accurate and up-to-date with minimal effort.

## ⚡ Daily Checks (5 minutes)

### 🔍 Quick Validation
- [ ] **Link Check**: Run `markdown-link-check docs/**/*.md` to find broken links
- [ ] **Code Syntax**: Verify code blocks have proper language tags
- [ ] **Version References**: Spot-check version numbers in recently changed files
- [ ] **New File Review**: Ensure new files follow naming conventions

### 📝 Content Updates
- [ ] **Recent Changes**: Update docs for any API/config changes from yesterday
- [ ] **Issue Review**: Check GitHub issues for documentation problems
- [ ] **Quick Fixes**: Fix any obvious typos or formatting issues
- [ ] **Status Update**: Update "Last Updated" dates on modified files

## 📅 Weekly Tasks (30 minutes)

### 🤖 Automated Checks
```bash
# Run comprehensive validation
cd /opt/entitydb
make docs-validate

# Regenerate API documentation
make swagger-docs

# Check for outdated content
find docs/ -name "*.md" -mtime +90 -exec echo "Review: {}" \;
```

### 📊 Quality Review
- [ ] **API Documentation**: Verify swagger docs match actual endpoints
- [ ] **Code Examples**: Test 3-5 random code examples for correctness
- [ ] **Cross-References**: Check internal links in 10 random documents
- [ ] **New Feature Coverage**: Ensure recent features are documented

## 🔄 Monthly Review (2 hours)

### 📋 Content Audit
- [ ] **Accuracy Check**: Verify 20% of documentation against current codebase
- [ ] **Completeness Review**: Check for missing documentation on new features
- [ ] **User Feedback**: Review support tickets and user feedback for doc issues
- [ ] **Metrics Analysis**: Check documentation usage analytics

### 🎯 Improvement Tasks
- [ ] **Content Gaps**: Identify and plan fixes for missing content
- [ ] **User Journey**: Test getting-started guide end-to-end
- [ ] **SEO Check**: Ensure proper headings and meta descriptions
- [ ] **Mobile Review**: Check documentation on mobile devices

## 🚨 Emergency Response

### 🔥 Critical Issues (Fix within 2 hours)
- **Security Documentation Error**: Incorrect security instructions
- **Data Loss Risk**: Wrong configuration that could cause data loss
- **API Breaking Change**: Undocumented breaking change in API

**Response**:
1. Add urgent warning notice to affected pages
2. Create emergency fix PR
3. Get expedited technical review
4. Deploy fix immediately
5. Notify users through appropriate channels

### ⚠️ High Priority (Fix within 24 hours)
- **New Feature**: Major feature released without documentation
- **Configuration Error**: Wrong configuration examples
- **Broken Getting Started**: Getting started guide doesn't work

**Response**:
1. Create high-priority documentation ticket
2. Assign to appropriate team member
3. Complete fix with normal review process
4. Deploy with next scheduled release

## 🛠️ Quick Tools

### 📝 Useful Commands
```bash
# Quick link validation
npx markdown-link-check docs/**/*.md

# Find recent changes
find docs/ -name "*.md" -mtime -7

# Check for TODO items
grep -r "TODO\|FIXME\|XXX" docs/

# Find files without recent updates
find docs/ -name "*.md" -mtime +180

# Word count for section
wc -w docs/getting-started/*.md

# Check for duplicate content
fdupes -r docs/
```

### 🔍 Quality Shortcuts
```bash
# Validate all code examples
extract-code-blocks docs/ | test-examples

# Check internal links only
find docs/ -name "*.md" -exec grep -l "\]\(\.\./" {} \; | xargs validate-internal-links

# Spell check with technical dictionary
cspell "docs/**/*.md" --config .cspell.json

# Style guide validation
textlint docs/**/*.md --config .textlintrc
```

## 📊 Quick Metrics

### ✅ Green Light Indicators
- All internal links resolve (>99.5%)
- Code examples execute successfully (>99%)
- No critical documentation issues open
- Getting started guide completes in <5 minutes
- Recent content has been updated within 30 days

### 🟡 Yellow Warning Signs
- 1-3 broken internal links
- 1-2 code examples failing
- Documentation updates >30 days behind code changes
- 1-2 minor accuracy issues reported
- Getting started takes 5-10 minutes

### 🔴 Red Alert Conditions
- >3 broken internal links
- >2 code examples failing
- Critical security/data loss documentation errors
- Major features undocumented for >1 week
- Getting started guide broken or >10 minutes

## 🎯 Weekly Metrics Tracking

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Link Validity | >99.5% | ___ | 🟢🟡🔴 |
| Code Example Success | >99% | ___ | 🟢🟡🔴 |
| Content Freshness | <30 days avg | ___ | 🟢🟡🔴 |
| User Task Success | >95% | ___ | 🟢🟡🔴 |
| Documentation Coverage | 100% | ___ | 🟢🟡🔴 |

## 📋 Responsibility Quick Reference

| Issue Type | Immediate Owner | Escalation |
|------------|----------------|------------|
| Broken Link | Any team member | Tech Writer |
| Code Example Failure | Engineer | Engineering Lead |
| API Documentation Error | Engineer | Product Manager |
| Getting Started Issue | Product Team | Technical Writer |
| Security Documentation | DevOps/Security | CTO |

## 🏆 Quality Mantras

> **"Documentation is code"** - Treat docs with same quality standards as source code

> **"User first"** - Every change should improve user experience

> **"Single source of truth"** - Eliminate contradictions and duplicates

> **"Accuracy over beauty"** - Correct information is more valuable than perfect formatting

> **"Test everything"** - If users will run it, we must test it

---

## 📞 Quick Support

- **Documentation Issues**: Create GitHub issue with `documentation` label
- **Urgent Fixes**: Slack #docs-emergency or email docs-team@entitydb.io
- **Questions**: Slack #docs-help or GitHub Discussions
- **Tool Issues**: Slack #dev-tools or create DevOps ticket

*This checklist ensures consistent, high-quality documentation maintenance with minimal time investment.*
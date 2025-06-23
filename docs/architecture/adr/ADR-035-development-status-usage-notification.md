# ADR-035: Development Status Disclaimer and Usage Notification License

**Status**: Accepted  
**Date**: 2025-06-23  
**Commit**: `4bae4b1`

## Context

Following production readiness certification in ADR-034, EntityDB required clear communication about its development status and a mechanism to track real-world usage. While the platform has been extensively tested and validated, it represents an actively developed project that would benefit from user feedback and community building.

Additionally, as an open-source project with significant architectural innovations, understanding how EntityDB is being used in practice would enable better support, feature development, and community engagement.

## Decision

1. **Add Development Status Disclaimer**: Clearly communicate that EntityDB is under heavy development and has not been tested in production environments, while encouraging its use for development, testing, and evaluation purposes.

2. **Implement Usage Notification License**: Create a modified MIT License requiring users to notify the author when using EntityDB in production environments, commercial settings, or significant projects.

## Implementation

### Development Status Disclaimer

Added prominent disclaimer to README.md:

```markdown
## ⚠️ DEVELOPMENT STATUS DISCLAIMER

> **🚧 UNDER HEAVY DEVELOPMENT**: This codebase is under active development and **has not been tested in production environments**.  
> **Use at your own risk** - suitable for development, testing, and evaluation purposes only.  
> **Production deployment is not recommended** without thorough testing and validation in your specific environment.
```

### Usage Notification License

Created "MIT License with Usage Notification" replacing standard MIT License:

```
MIT License with Usage Notification

Copyright (c) 2025 ITDLabs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

1. The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

2. USAGE NOTIFICATION REQUIREMENT: Any person or organization using this Software
   in a production environment, commercial setting, or for any significant project
   SHALL notify the original author by sending an email to: licensing@itdlabs.co.uk
   
   The notification shall include:
   - Brief description of the intended use case
   - Organization/project name (if applicable)
   - Contact information for ongoing communication
   
   This notification requirement enables the author to:
   - Understand how the software is being utilized
   - Provide support and updates when needed
   - Build a community around the project
   - Potentially collaborate on improvements

3. DEVELOPMENT STATUS ACKNOWLEDGMENT: Users acknowledge that this software is
   under heavy development and has not been tested in production environments.
   Use in production is at the user's own risk.
```

### Documentation Updates

Updated badges and references throughout documentation:

```markdown
[![License](https://img.shields.io/badge/license-MIT%20with%20Usage%20Notification-green)](./LICENSE)
```

## Rationale

### Development Status Transparency

**Honest Communication**: Despite extensive testing and production readiness certification (ADR-034), honest disclosure about development status maintains user trust and sets appropriate expectations.

**Risk Management**: Users can make informed decisions about deployment based on clear understanding of the project's maturity level.

**Encourages Evaluation**: Disclaimer specifically encourages use for development, testing, and evaluation, promoting adoption for appropriate use cases.

### Usage Notification Benefits

**Community Building**: Understanding real-world usage enables building a community around actual use cases rather than theoretical applications.

**Support Enhancement**: Knowledge of how EntityDB is being used allows for targeted support and documentation improvements.

**Feature Development**: User feedback from production environments informs feature prioritization and development roadmap.

**Collaboration Opportunities**: Direct communication with users can lead to collaboration, contributions, and partnerships.

**Quality Improvement**: Production usage reports help identify edge cases and improvement opportunities not covered in testing.

### License Choice Rationale

**MIT Foundation**: Maintains permissive open-source licensing ensuring wide adoption and compatibility.

**Minimal Burden**: Simple email notification requirement with clear purpose and benefit explanation.

**Legal Clarity**: Clear language about notification requirements and development status acknowledgment.

**Author Rights**: Preserves author's ability to understand project impact while maintaining user freedom.

## Consequences

### Positive

**Enhanced Community Engagement**:
- Direct communication channel with production users
- Opportunity for feedback and collaboration
- Understanding of real-world use cases and requirements
- Building relationships with enterprise and commercial users

**Better Project Development**:
- Informed feature development based on actual usage patterns
- Targeted documentation and support improvements
- Quality feedback from production environments
- Partnership and contribution opportunities

**User Trust and Transparency**:
- Honest communication about development status builds trust
- Clear expectations prevent disappointed users
- Appropriate use case guidance helps successful implementations
- Legal clarity provides certainty for commercial users

### Negative

**Potential Adoption Friction**:
- Additional license requirements may deter some users
- Development status disclaimer might discourage production adoption
- Notification requirement adds administrative overhead for users

**Support Expectations**:
- Users contacting for support may expect immediate responses
- Production usage notifications could create implicit support obligations
- Success stories might create pressure for continued development

### Neutral

**License Complexity**:
- More complex than standard MIT but simpler than many commercial licenses
- Legal review required for some organizations
- Clear termination of obligations if development status changes

## Implementation Timeline

**Immediate (2025-06-23)**:
- ✅ README.md updated with development disclaimer
- ✅ LICENSE file replaced with usage notification version
- ✅ Documentation badges updated
- ✅ Repository metadata reflects new license

**Short Term (Next 30 days)**:
- Monitor for user notifications and questions
- Document any support requests or feedback received
- Evaluate effectiveness of disclaimer and notification system

**Medium Term (Next 90 days)**:
- Analyze usage patterns from notifications received
- Consider updates to disclaimer based on actual production adoption
- Evaluate if notification system provides expected benefits

## Success Metrics

**Community Engagement**:
- Number of usage notifications received
- Quality and diversity of use cases reported
- Collaboration opportunities created
- User feedback and contributions

**Project Improvement**:
- Feature requests based on production usage
- Bug reports from real-world environments
- Documentation improvements identified
- Partnership and contribution opportunities

## References

- **Commit**: `4bae4b1` - Add development disclaimer and usage notification license
- **ADR-034**: Production readiness certification (prerequisite)
- **MIT License**: Foundation for usage notification license
- **README.md**: Development status disclaimer implementation
- **LICENSE**: Complete usage notification license text

## Future Considerations

**License Evolution**:
- Consider transitioning to standard MIT if project reaches full production maturity
- Evaluate effectiveness of notification requirement after 6-12 months
- Potentially add contributor license agreement (CLA) for significant contributions

**Development Status Updates**:
- Update disclaimer as project matures and gains production validation
- Consider production readiness certification levels
- Document transition criteria from development to production status

**Community Development**:
- Establish formal support channels based on user feedback
- Create contribution guidelines based on collaboration opportunities
- Develop user community platforms if significant adoption occurs
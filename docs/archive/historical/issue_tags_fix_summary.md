# Issue Tags Fix Summary

This update addresses a bug where tags were not being saved properly for epic and story issue types. The root cause was that the code was only updating issues with optional fields (including tags) for regular issues and subissues, but not for epics and stories.

## Changes Made

1. Modified `/opt/entitydb/src/api/issue.go` to always update issues with optional fields, regardless of issue type.
2. Updated `/opt/entitydb/src/api/test_endpoints_fix.go` to handle tags in test endpoints.
3. Added a test script (`/opt/entitydb/share/tests/api/issue/test_issue_tags.sh`) to verify the fix.
4. Created documentation explaining the issue and solution.

## Verification

The fix has been tested with the following:

1. Creating issues of each type (workspace, epic, story, issue) with tags
2. Verifying that the tags are saved correctly for all issue types
3. Running the test script to validate the behavior

This update ensures consistent behavior across all issue types and fixes the regression in tag handling.
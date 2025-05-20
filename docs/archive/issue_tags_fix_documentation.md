# Issue Tags Fix Documentation

## Problem Description

The system had an issue handling tags when creating certain types of issues (epics and stories). While the `tags` field was properly defined in the `CreateIssueRequest` struct and the database schema included an `issue_tags` table, the implementation had a flaw:

1. For regular issues and subissues:
   - Tags were set to the issue object after it was created
   - Then an update call was made to save these tags to the database

2. For epics and stories:
   - Tags were set to the issue object after it was created
   - But no update call was made to save these tags to the database
   - This resulted in tags being lost for epics and stories

3. In the test endpoints:
   - The mock issue creation endpoint didn't handle tags at all, leading to inconsistent behavior in tests

## Solution

The following changes were made to fix the issue:

1. In `/opt/entitydb/src/api/issue.go`:
   - Modified the code to always perform an update after setting optional fields (including tags), regardless of issue type.
   - Removed the conditional check that was only updating regular issues and subissues, but not epics and stories.

2. In `/opt/entitydb/src/api/test_endpoints_fix.go`:
   - Added code to handle tags in the test endpoint for issue creation
   - Ensured tags are included in the mock response when provided

## Affected Code

### Before:
```go
// For epics and stories, we've already saved the issue
// For regular issues and subissues, we need to update with the optional fields
if issueType != models.IssueTypeEpic && issueType != models.IssueTypeStory {
    if updateErr := h.repo.Update(issue); updateErr != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to update issue with optional fields")
        return
    }
}
```

### After:
```go
// Always update the issue to save tags and other optional fields
// regardless of issue type (epic, story, issue, subissue)
if updateErr := h.repo.Update(issue); updateErr != nil {
    RespondError(w, http.StatusInternalServerError, "Failed to update issue with optional fields")
    return
}
```

## Testing

To test this fix:

1. Create an epic with tags:
```bash
./bin/entitydbc.sh issue create \
  --title="Test Epic with Tags" \
  --description="Testing tags on epics" \
  --type=epic \
  --priority=medium \
  --tags="tag1,tag2,epic"
```

2. Create a story with tags:
```bash
./bin/entitydbc.sh issue create \
  --title="Test Story with Tags" \
  --description="Testing tags on stories" \
  --type=story \
  --priority=medium \
  --parent_id=<epic_id> \
  --tags="tag1,tag2,story"
```

3. View both issues to verify tags are present:
```bash
./bin/entitydbc.sh issue view <issue_id>
```

This update ensures that tags work consistently across all issue types (workspace, epic, story, issue, subissue) and fixes the regression in epic and story tag handling.
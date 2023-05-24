## Describe your changes and why you are making these changes

## Related issue number (if any)

## Test Plans
Reviewer note: by approving the PR, reviewer agree that the test plans are solid enough to merge, with risk low enough for daily tests / release eng to take over.
### Tests specific to this PR
Include tests specific to this PR's context here:
### Tests based on components touched by this PR
Check the components touched by this PR and ensures all items under that component are checked. If certain item is not checked, explicitly state why and agree with PR reviewers.
- [ ] This PR involves SDK changes
  - [ ] Have integration test that covers my change
  - [ ] Added 'run_integration_test' label to this PR
  - [ ] If you are changing user-facing APIs, the change is backward compatible
- [ ] This PR involves backend (go and executor changes)
  - [ ] Have integration test that covers my change
  - [ ] If changing REST endpoint, have backend test that covers this change (under `integration_tests/backend`)
  - [ ] Added 'run_integration_test' label to this PR
  - [ ] This PR involves changes in database layer
    - [ ] There's a database unit test covering this change
    - [ ] If the change affects how data is stored, has carefully examined if migration is required.
- [ ] This PR involves engine-specific changes
  - [ ] Performed integration tests (manually or automatically) backed by this engine.
- [ ] This PR involves UI changes
  - [ ] All manual_qa_tests workflows works with this feature.
  - [ ] Critical integration registration works with this feature. (k8s, s3, and snowflakes) Refer to https://www.notion.so/aqueducthq/Compute-Resource-Setup-Guide-c83e25e1bc6847efbb226f6fc86fa5cd on connecting to integrations.
  - [ ] If changing a shared component, ensure all callers to this component are updated.
  - [ ] Included a loom demo / screenshot for this feature.


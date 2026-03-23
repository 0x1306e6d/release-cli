## REMOVED Requirements

### Requirement: Freeform convention treats all commits as releasable
**Reason**: The freeform behavior is already the default when `changes.commits` is absent. An explicit convention value is redundant.
**Migration**: Remove the `changes.commits` section (or at minimum `convention: freeform`) from `.release.yaml`. The system defaults to the same behavior.

### Requirement: Freeform convention always produces patch bumps
**Reason**: Covered by default behavior when `changes.commits` is absent.
**Migration**: See above.

### Requirement: Freeform convention uses full subject as description
**Reason**: Covered by default behavior when `changes.commits` is absent.
**Migration**: See above.

### Requirement: Freeform convention does not detect breaking changes
**Reason**: Covered by default behavior when `changes.commits` is absent.
**Migration**: See above.

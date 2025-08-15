# Environment Variable Refactoring Plan

## Overview
Simplify the Performer CRD environment variable handling by consolidating multiple environment definitions into a single field using standard Kubernetes types.

## Current Problems
- [ ] Multiple environment definitions (`Environment` map and `EnvironmentFrom` array)
- [ ] Custom types duplicating Kubernetes native types
- [ ] Complex translation layers between different type systems
- [ ] Confusing API surface for developers

## Refactoring Checklist

### Phase 1: CRD Definition Updates
**File:** `/hourglass-operator/api/v1alpha1/performerTypes.go`

- [ ] Add new `Env []corev1.EnvVar` field to `PerformerConfig`
- [ ] Add deprecation comments to existing `Environment` and `EnvironmentFrom` fields
- [ ] Remove custom `EnvVarSource` and `EnvValueFrom` type definitions
- [ ] Update CRD validation markers for new field
- [ ] Run `make manifests` to regenerate CRD YAML

### Phase 2: Ponos Type Updates
**File:** `/ponos/pkg/kubernetesManager/types.go`

- [ ] Add `Env []corev1.EnvVar` to `CreatePerformerRequest`
- [ ] Mark `Environment` and `EnvironmentFrom` as deprecated
- [ ] Remove custom `EnvVarSource`, `EnvValueFrom`, and `KeySelector` types
- [ ] Update `DeepCopyInto` methods to handle new field

**File:** `/ponos/pkg/kubernetesManager/crd.go`

- [ ] Update `CreatePerformer` to use new `Env` field
- [ ] Add backward compatibility logic for old fields
- [ ] Update validation functions for new structure
- [ ] Add deprecation warnings when old fields are used

### Phase 3: Kubernetes Performer Updates
**File:** `/ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/kubernetesPerformer.go`

- [ ] Refactor `buildEnvironmentFromImage` to return `[]corev1.EnvVar`
- [ ] Remove separation between direct values and references
- [ ] Update `createPerformerResource` to pass `[]corev1.EnvVar` directly
- [ ] Simplify environment variable processing logic

**File:** `/ponos/pkg/config/config.go`

- [ ] Add new `Env []corev1.EnvVar` field to `AVSPerformer`
- [ ] Update `AVSPerformerEnv` struct with deprecation notice
- [ ] Add migration helper to convert old format to new
- [ ] Update validation logic

### Phase 4: Controller Updates
**File:** `/hourglass-operator/internal/controller/performerController.go`

- [ ] Update `reconcilePod` to use new `Env` field directly
- [ ] Remove separate handling for `Environment` and `EnvironmentFrom`
- [ ] Add backward compatibility for existing CRDs
- [ ] Simplify environment variable assignment to container

### Phase 5: Test Updates

**Files to Update:**
- [ ] `/ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/kubernetesPerformer_test.go`
  - [ ] Update all test cases to use new `Env` structure
  - [ ] Add migration tests for backward compatibility
  - [ ] Remove tests for old environment handling

- [ ] `/hourglass-operator/internal/controller/performerController_test.go` (if exists)
  - [ ] Update controller tests for new structure
  - [ ] Add tests for backward compatibility

### Phase 6: Documentation Updates

- [ ] Update API documentation in `/hourglass-operator/docs/operator/api-reference.md`
- [ ] Update examples in `/hourglass-operator/docs/operator/examples.md`
- [ ] Add migration guide section
- [ ] Update any helm chart values that reference environment variables

### Phase 7: Migration Support

- [ ] Create migration script for existing CRDs
- [ ] Add logging to detect and warn about deprecated field usage
- [ ] Document breaking changes in CHANGELOG
- [ ] Plan removal timeline for deprecated fields (e.g., v2.0.0)

## Target Structure

### Before (Current)
```go
type PerformerConfig struct {
    Environment     map[string]string  // Direct values
    EnvironmentFrom []EnvVarSource     // References to secrets/configmaps
}
```

### After (Target)
```go
type PerformerConfig struct {
    Env []corev1.EnvVar  // Single field using k8s native type
    
    // Deprecated: Use Env instead
    Environment     map[string]string
    // Deprecated: Use Env instead  
    EnvironmentFrom []EnvVarSource
}
```

## Usage Example (After Refactoring)

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
spec:
  config:
    env:
    - name: DATABASE_URL
      value: "postgres://localhost/db"
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-secrets
          key: api-key
    - name: CONFIG_DATA
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: config.json
```

## Testing Checklist

- [ ] Unit tests pass for all modified components
- [ ] Integration tests pass with new structure
- [ ] Backward compatibility tests pass
- [ ] E2E tests pass in demo environment
- [ ] Manual testing of environment variable injection
- [ ] Test secret and configmap references work correctly
- [ ] Test that old CRDs continue to work

## Rollout Plan

1. **Version 1.x.x** (Current + Compatibility)
   - [ ] Add new `Env` field
   - [ ] Maintain backward compatibility
   - [ ] Log deprecation warnings
   - [ ] Update documentation

2. **Version 1.x+1.x** (Transition)
   - [ ] Make new field the primary recommendation
   - [ ] Provide migration tooling
   - [ ] Increase deprecation warning visibility

3. **Version 2.0.0** (Clean)
   - [ ] Remove deprecated fields
   - [ ] Clean up backward compatibility code
   - [ ] Final documentation update

## Success Criteria

- [ ] Single, clear API for environment variables
- [ ] Full compatibility with Kubernetes EnvVar specification
- [ ] No custom type definitions for environment handling
- [ ] Simplified codebase with fewer translation layers
- [ ] Zero breaking changes for existing deployments (until v2.0.0)
- [ ] Clear migration path documented

## Notes

- Ensure CRD versioning is properly handled
- Consider using conversion webhooks for seamless migration
- Monitor for any performance implications
- Coordinate with team on deprecation timeline
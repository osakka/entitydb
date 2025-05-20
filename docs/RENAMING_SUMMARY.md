# EntityDB v2.10.0 Professional Naming Update

## Overview

This major update replaces all "turbo" and buzzword terminology with professional naming throughout the EntityDB codebase.

## Changes Made

### Code Refactoring
- `TurboEntityRepository` → `HighPerformanceRepository`
- `TemporalTurboRepository` → `TemporalRepository`
- All method signatures and references updated
- File renames:
  - `turbo_repository.go` → `high_performance_repository.go`
  - `temporal_turbo_repository.go` → `temporal_repository.go`
  - `test_turbo.go` → `test_high_performance.go`

### Environment Variables
- `ENTITYDB_DISABLE_TURBO` → `ENTITYDB_DISABLE_HIGH_PERFORMANCE`
- `ENTITYDB_TEMPORAL_TURBO` → `ENTITYDB_TEMPORAL`
- `ENTITYDB_ENABLE_TURBO` → `ENTITYDB_ENABLE_HIGH_PERFORMANCE`
- `ENTITYDB_TURBO_WORKERS` → `ENTITYDB_HIGH_PERFORMANCE_WORKERS`

### Documentation Updates
- Updated all references in documentation files
- Renamed documentation files:
  - `TURBO_MODE_REPORT.md` → `HIGH_PERFORMANCE_MODE_REPORT.md`
  - `TEMPORAL_TURBO_IMPLEMENTATION.md` → `TEMPORAL_IMPLEMENTATION.md`
- Updated CHANGELOG.md and CLAUDE.md

### Test Files
- Renamed all test files containing "turbo":
  - `turbo_benchmark.py` → `high_performance_benchmark.py`
  - `test_turbo_*.sh` → `test_high_performance_*.sh`
  - `temporal_turbo_*.py` → `temporal_*.py`

## Summary

This update ensures EntityDB uses professional, industry-standard terminology throughout its codebase while maintaining all existing functionality. The changes are purely cosmetic and do not affect the underlying implementation or performance.

## Version Information
- Version: v2.10.0
- Date: 2025-05-19
- Type: Major nomenclature update
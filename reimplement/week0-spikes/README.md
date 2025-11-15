# Week 0 Dependency Validation Spikes

This directory contains the infrastructure and tools for executing the three critical validation spikes that de-risk the v2 Go rewrite project.

## Overview

**Goal**: Validate critical technical assumptions before Phase 0 begins, reducing project risk from Medium-High to Low-Medium.

**Duration**: 3-5 days (can run in parallel)
**Team size**: 1-2 people
**Success criterion**: All three spikes completed with documented results

---

## The Three Spikes

### Spike 1: Amion HTML Scraping Feasibility (2 days)
**Owner**: Backend Engineer or Test Engineer

**Question**: Can goquery efficiently parse Amion schedules, or must we use Chromedp?

**Deliverable**: Document with goquery capability assessment
- CSS selector paths and stability
- Parsing accuracy (>95% target)
- Performance for 6-month dataset (target: 2-3 seconds)
- Recommendation: goquery OR fallback to Chromedp (+2 weeks cost)

**Location**: `spike1-amion-scraping/`

---

### Spike 2: Job Library Evaluation (2 days)
**Owner**: Test/DevOps Engineer

**Question**: Which job queue library should we use? (Asynq vs Machinery vs custom)

**Deliverable**: Comparison document
- Redis availability in hospital infrastructure
- Asynq integration test results
- Machinery (PostgreSQL broker) test results
- Recommendation: Asynq OR Machinery OR custom (+3 weeks cost)

**Location**: `spike2-job-library/`

---

### Spike 3: ODS Library Validation (1-2 days)
**Owner**: Backend Engineer

**Question**: Does the chosen ODS library support error collection pattern?

**Deliverable**: Validation document
- Library capability assessment
- Error collection pattern implementation
- Performance with production file sizes
- File size limits and constraints
- Recommendation: Proceed OR wrapper layer needed (+1 week cost)

**Location**: `spike3-ods-library/`

---

## Quick Start

### Prerequisites

```bash
# Python 3.9+
python3 --version

# Go 1.20+ (for job library testing)
go version

# PostgreSQL client (for connectivity tests)
psql --version

# Redis CLI (for Redis tests)
redis-cli --version
```

### Running All Spikes

```bash
# Clone the repo / navigate to week0-spikes directory
cd week0-spikes

# Run all spikes with defaults (uses mocks/test data)
./run_all_spikes.sh

# Or run individually:
./spike1-amion-scraping/run_spike.sh
./spike2-job-library/run_spike.sh
./spike3-ods-library/run_spike.sh
```

### Running Against Real Infrastructure

```bash
# Spike 1: Against real Amion (requires credentials)
./spike1-amion-scraping/run_spike.sh --real-amion --username <user> --password <pass>

# Spike 2: Against real Redis/PostgreSQL
./spike2-job-library/run_spike.sh --redis-host localhost --postgres-host localhost

# Spike 3: Against production ODS files
./spike3-ods-library/run_spike.sh --ods-dir /path/to/production/files
```

---

## Directory Structure

```
week0-spikes/
├── README.md                          # This file
├── run_all_spikes.sh                  # Master script to run all spikes
├── RESULTS_TEMPLATE.md                # Template for documenting results
│
├── spike1-amion-scraping/
│   ├── README.md                      # Spike 1 details
│   ├── run_spike.sh                   # Execute Spike 1
│   ├── requirements.txt                # Python dependencies
│   ├── main.py                        # Main spike logic
│   ├── parser.py                      # goquery-equivalent parsing
│   ├── performance_test.py            # Benchmarking
│   ├── mock_data/
│   │   ├── sample_amion_page.html     # Sample HTML for testing
│   │   └── amion_schedule_samples/    # 6-month sample data
│   ├── tests/
│   │   ├── test_parser.py             # Parser tests
│   │   ├── test_performance.py        # Performance tests
│   │   └── conftest.py                # Pytest fixtures
│   └── results/
│       └── spike1_results.md          # Populated after run
│
├── spike2-job-library/
│   ├── README.md                      # Spike 2 details
│   ├── run_spike.sh                   # Execute Spike 2
│   ├── go.mod                         # Go module
│   ├── main.go                        # Main spike logic
│   ├── asynq_integration_test.go      # Asynq testing
│   ├── machinery_integration_test.go  # Machinery testing
│   ├── infrastructure_check.go        # Redis/PostgreSQL checks
│   ├── docker-compose.yml             # Local Redis/PostgreSQL
│   ├── tests/
│   │   └── integration_test.go        # Integration tests
│   └── results/
│       └── spike2_results.md          # Populated after run
│
├── spike3-ods-library/
│   ├── README.md                      # Spike 3 details
│   ├── run_spike.sh                   # Execute Spike 3
│   ├── requirements.txt                # Python dependencies
│   ├── main.py                        # Main spike logic
│   ├── ods_validator.py               # ODS library testing
│   ├── performance_test.py            # Benchmarking
│   ├── mock_data/
│   │   ├── sample_small.ods           # Small test file
│   │   ├── sample_medium.ods          # Medium test file
│   │   └── sample_large.ods           # Large test file
│   ├── tests/
│   │   ├── test_ods_parsing.py        # Parsing tests
│   │   ├── test_error_collection.py   # Error pattern tests
│   │   └── conftest.py                # Pytest fixtures
│   └── results/
│       └── spike3_results.md          # Populated after run
│
├── shared/
│   ├── constants.py                   # Shared constants
│   ├── utils.py                       # Shared utilities
│   ├── result_generator.py            # Result document generator
│   └── infrastructure_utils.go        # Shared Go utilities
│
└── docs/
    ├── SPIKE_EXECUTION_GUIDE.md       # Step-by-step execution guide
    ├── SUCCESS_CRITERIA.md             # Success criteria for each spike
    ├── FALLBACK_TIMELINES.md          # Cost of each fallback
    └── TEAM_SKILLS_ASSESSMENT.md      # Skills needed per spike
```

---

## Running the Spikes

### Option 1: Quick Test (5 minutes)
Uses mock data and test infrastructure (no external dependencies required)

```bash
./run_all_spikes.sh --mock
```

This validates that the spike infrastructure itself works correctly.

### Option 2: Against Your Infrastructure (1-2 hours)
Tests against real hospital infrastructure (Redis, PostgreSQL, etc.)

```bash
# First, verify infrastructure connectivity
./shared/check_infrastructure.sh

# Then run spikes
./run_all_spikes.sh --infrastructure <redis-host> <postgres-host>
```

### Option 3: Manual Execution
Execute spikes individually with fine-grained control

```bash
cd spike1-amion-scraping
./run_spike.sh --help
./run_spike.sh --verbose --output-format json
```

---

## Success Criteria

Each spike has specific success criteria. See `RESULTS_TEMPLATE.md` for the format.

### Spike 1: Amion Parsing
- ✓ goquery CSS selectors work (>95% accuracy) OR
- ✓ Clear documentation of fallback to Chromedp (cost: +2 weeks)

### Spike 2: Job Library
- ✓ Asynq works with hospital Redis OR
- ✓ Machinery works with hospital PostgreSQL OR
- ✓ Clear documentation of custom queue approach (cost: +3 weeks)

### Spike 3: ODS Library
- ✓ ODS library handles error collection pattern OR
- ✓ Wrapper layer designed with implementation plan (cost: +1 week)

---

## Output & Documentation

After each spike, results are documented in `results/spike<N>_results.md`:

```markdown
# Spike N Results

**Date**: 2025-11-15
**Engineer**: Backend Engineer
**Status**: ✓ SUCCESS / ⚠ WARNING / ✗ FAILURE

## Findings
- ...

## Recommendation
- ...

## Timeline Impact
- None / +X weeks

## Next Steps
- ...
```

---

## Integration with MASTER_PLAN_v2.md

After all spikes are completed:

1. Document results in `results/` directory
2. Update MASTER_PLAN_v2.md Week 0 section with findings
3. Adjust Phase 0-4 timelines based on fallback recommendations
4. Proceed to Phase 0 with confidence ✓

---

## Questions?

See individual spike READMEs for detailed instructions:
- `spike1-amion-scraping/README.md`
- `spike2-job-library/README.md`
- `spike3-ods-library/README.md`

Or check the execution guide:
- `docs/SPIKE_EXECUTION_GUIDE.md`

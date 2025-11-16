# Minimum Spanning Set Cover Calculator

Find the **minimum number of spanning sets** needed to cover ALL study types using a greedy set cover algorithm.

## Purpose

Given all possible spanning sets (by modality, specialty, hospital, cross-dimensional), what's the **smallest collection** that covers every single study type?

This solves the classic **SET COVER problem** (NP-complete) using a greedy approximation algorithm.

## The Problem

**Given:**
- Universe U = all 23 study types (e.g., "CPMC CT Neuro", "Allen MR Body", etc.)
- Collection S = 22 spanning sets (All CT, All MRI, CPMC - All Services, etc.)
- Each set covers some subset of U

**Find:** Minimum number of sets from S that cover all elements of U

## Algorithm

**Greedy Set Cover:**
1. Start with all study types marked as "uncovered"
2. Pick the spanning set that covers the **most uncovered** studies
3. Mark those studies as covered
4. Repeat until all studies are covered

This is an approximation algorithm that guarantees a solution within `ln(n)` of optimal (where n = number of elements).

## Usage

### Build and Run

```bash
cd /home/user/schedCU/reimplement/cmd/minimum-spanning-cover
go build -o minimum-spanning-cover main.go
./minimum-spanning-cover
```

Or specify a custom ODS file:

```bash
./minimum-spanning-cover /path/to/your/schedule.ods
```

### Output

The tool automatically:
1. Runs greedy set cover algorithm
2. Shows step-by-step which sets are chosen
3. Displays final minimum cover
4. Analyzes alternative covers by dimension
5. Exports results to JSON

No interactive menu - just run and get results.

## Example Output

```
======================================================================
         MINIMUM SPANNING SET COVER CALCULATOR
======================================================================

✓ Found 23 unique study types to cover
✓ Built 22 spanning sets

======================================================================
GREEDY SET COVER ALGORITHM
======================================================================

Step 1: CPMC - All Services (hospital)
        Covers 8 new studies (total: 8/23 covered)

Step 2: Allen - All Services (hospital)
        Covers 7 new studies (total: 15/23 covered)

Step 3: NYPLH - All Services (hospital)
        Covers 7 new studies (total: 22/23 covered)

Step 4: All Unknown (modality)
        Covers 1 new studies (total: 23/23 covered)
        New: CHONY Neuro

======================================================================
MINIMUM SPANNING SET COVER RESULT
======================================================================

✓ Minimum sets needed: 4
✓ Coverage: 23 studies (100.0%)

CHOSEN SPANNING SETS:
──────────────────────────────────────────────────────────────────────
1. CPMC - All Services (hospital)
   Members: 8
2. Allen - All Services (hospital)
   Members: 7
3. NYPLH - All Services (hospital)
   Members: 7
4. All Unknown (modality)
   Members: 1

INTERPRETATION:
This is the MINIMUM number of spanning sets needed to describe all
study types. Use these sets for:
  • Simplest possible summary of coverage
  • Prioritizing which aggregations to show first
  • Understanding natural groupings in your schedule

======================================================================
ALTERNATIVE COVERS BY DIMENSION
======================================================================

Using only MODALITY dimension:
  Sets needed: 5
  Coverage: 100.0%
  Sets: All CT, All MRI, All X-Ray, All US, All Unknown

Using only SPECIALTY dimension:
  Sets needed: 5
  Coverage: 100.0%
  Sets: Neuro - All Modalities, Body - All Modalities,
        Chest - All Modalities, General - All Modalities,
        Bone - All Modalities

Using only HOSPITAL dimension:
  Sets needed: 4
  Coverage: 100.0%
  Sets: CPMC - All Services, Allen - All Services,
        NYPLH - All Services, CHONY - All Services

Using only CROSS dimension:
  Sets needed: 8
  Coverage: 87.0%
  Sets: MRI Neuro - All Hospitals, MRI Body - All Hospitals,
        X-Ray Chest - All Hospitals, ...

INSIGHT: Compare which dimension provides the most efficient coverage.

✓ Results exported to minimum_spanning_cover.json
  Efficiency: Using 4/22 sets (18.2% of total)
```

## Results Interpretation

### For cuSchedNormalized.ods

**Minimum cover: 4 sets (18.2% efficiency)**

The greedy algorithm chose:
1. **CPMC - All Services** (hospital) - covers 8 studies
2. **Allen - All Services** (hospital) - covers 7 studies
3. **NYPLH - All Services** (hospital) - covers 7 studies
4. **All Unknown** (modality) - covers 1 study (CHONY Neuro)

**Key insight**: Hospital-based grouping is most efficient!

### Dimension Comparison

| Dimension | Sets Needed | Coverage | Efficiency |
|-----------|-------------|----------|------------|
| **Hospital** | **4** | **100%** | **Best** |
| Modality | 5 | 100% | Good |
| Specialty | 5 | 100% | Good |
| Cross-dimensional | 8 | 87% | Incomplete |

**Conclusion**: If you can only show a few spanning sets in your UI, show hospital-based sets first.

## Use Cases

### 1. Simplify Dashboards

**Problem**: You have 22 spanning sets, but can only display 3-5 on a dashboard.

**Solution**: Use minimum cover to prioritize which sets to show:
- Show: CPMC, Allen, NYPLH, CHONY (covers everything)
- Hide: Less efficient aggregations

### 2. Executive Summaries

**Problem**: Executives want the "simplest possible" summary of coverage.

**Solution**:
```
"We provide coverage across 4 hospital locations:
  • CPMC (8 services)
  • Allen (7 services)
  • NYPLH (7 services)
  • CHONY (1 service)"
```

This is mathematically the **most concise** accurate summary.

### 3. UI Prioritization

**Problem**: Which spanning sets should appear first in dropdown menus?

**Solution**: Order by appearance in minimum cover:
1. CPMC - All Services
2. Allen - All Services
3. NYPLH - All Services
4. All Unknown
5. (Everything else)

### 4. Natural Grouping Discovery

**Problem**: What's the "natural" way to group our services?

**Solution**: The minimum cover reveals that hospital location is the primary organizing principle (4 hospital sets vs 5 modality/specialty sets).

### 5. Coverage Optimization

**Problem**: We want to add new services - where should we add them?

**Solution**: If minimum cover requires 5+ sets, there may be opportunities to consolidate. If it's already 4, services are well-distributed across hospitals.

### 6. Comparative Analysis

**Problem**: How does our schedule compare to another hospital's?

**Solution**: Compare minimum cover sizes:
- Hospital A: 4 sets (efficient, hospital-focused)
- Hospital B: 8 sets (fragmented, needs consolidation)

## Advanced Insights

### Why Hospital Dimension Wins

The hospital dimension is most efficient (4 sets) because:
1. **Non-overlapping**: Each study type belongs to exactly one hospital
2. **Balanced**: Hospitals have similar numbers of services (8, 7, 7, 1)
3. **Complete**: Every study type is assigned to a hospital

### Why Cross-Dimensional Loses

Cross-dimensional sets only achieve 87% coverage because:
1. **Incomplete**: Not all modality+specialty combinations exist
2. **Fragmented**: Many small sets instead of few large ones
3. **Overlapping**: Some studies could fit multiple cross-dimensional sets

### Greedy vs Optimal

The greedy algorithm is NOT guaranteed to find the absolute optimal solution, but:
- It's fast: O(n * m) where n=studies, m=sets
- It's close: Within ln(n) factor of optimal
- It's practical: Exact solution is NP-complete

For our problem (23 studies, 22 sets), greedy is sufficient.

## Export Format

Results saved to `minimum_spanning_cover.json`:

```json
{
  "minimum_cover": {
    "sets_needed": 4,
    "coverage_percent": 100.0,
    "sets": [
      "CPMC - All Services",
      "Allen - All Services",
      "NYPLH - All Services",
      "All Unknown"
    ]
  },
  "all_study_types": [
    "Allen CT Body",
    "Allen CT Neuro",
    ...
  ],
  "total_sets_available": 22,
  "optimal_efficiency_percent": 18.2
}
```

## Mathematical Background

### Set Cover Problem

**Definition**: Given a universe U and a collection S of subsets of U, find the smallest sub-collection of S whose union equals U.

**Complexity**: NP-complete (no known polynomial-time exact solution)

**Approximation**: Greedy algorithm achieves `H(n)` approximation where `H(n) = ln(n) + O(1)` is the harmonic number

For our problem:
- Universe size: 23 studies
- Greedy approximation factor: ln(23) ≈ 3.14
- Greedy solution: 4 sets
- Theoretical optimal: Could be as low as ⌈4/3.14⌉ = 2 sets

(In practice, greedy often matches optimal for structured problems like this)

### Proof of Correctness

The greedy algorithm is correct because:
1. It always terminates (finite set of sets)
2. It always makes progress (covers at least 1 new study per step)
3. It achieves 100% coverage (or reports if impossible)

## Tips

1. **Run periodically**: As schedules change, minimum cover may change

2. **Compare over time**: Track if minimum cover size is increasing (fragmentation) or decreasing (consolidation)

3. **Use with rank-spanning-sets**: Combine minimum cover with user preferences:
   - Start with minimum cover for efficiency
   - Swap in preferred sets that have similar coverage

4. **Export for analysis**: Use JSON output for programmatic analysis

5. **Validate completeness**: If coverage < 100%, investigate why some studies aren't covered by any spanning set

## Limitations

1. **Greedy not optimal**: May not find absolute minimum (but close)

2. **No preferences**: Doesn't consider which sets are clinically meaningful

3. **No constraints**: Doesn't enforce "must include" or "must exclude" sets

4. **Single objective**: Only minimizes count, not other factors (e.g., clinical coherence)

## Future Enhancements

Potential improvements:
- **Exact solver**: Branch-and-bound or ILP for guaranteed optimal
- **Weighted cover**: Prioritize clinically important sets
- **Constrained cover**: Force inclusion/exclusion of specific sets
- **Multi-objective**: Minimize count while maximizing clinical coherence
- **Historical analysis**: Track how minimum cover changes over time
- **Integration with rankings**: Use preference data from rank-spanning-sets
- **Visual graph**: Show overlaps and coverage graphically

## Related Tools

- **spanning-sets**: Generate all possible spanning sets
- **rank-spanning-sets**: Collect user preferences on which sets are most useful
- **cutout-sets**: Define sets with exclusions
- **minimum-spanning-cover**: Find minimum collection to cover everything (THIS TOOL)

**Workflow**: Use spanning-sets to generate sets, minimum-spanning-cover to find the essential ones, then rank-spanning-sets to refine based on clinical usefulness.

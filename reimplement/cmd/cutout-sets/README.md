# Cutout Sets Analyzer

Define spanning sets with exclusion logic - e.g., "All MSK except MRI" to create custom aggregations.

## Purpose

Create spanning sets by **subtracting** specific dimensions from a base set. This allows you to define coverage patterns that may not be explicitly enumerated in the schedule file.

**Key insight**: Some cutouts may be **EMPTY** (no matching studies exist). This is intentional and useful for finding gaps or testing hypothetical scenarios.

## What are Cutout Sets?

A cutout set is defined by:
1. **Base set**: What to include (e.g., "All MSK specialty")
2. **Exclusions**: What to remove (e.g., "except MRI modality")

Example: "All MSK except MRI"
- Base: specialty = MSK
- Exclusions: modality = MRI
- Result: CT MSK, X-Ray MSK, US MSK (but NOT MRI MSK)

## Usage

### Build and Run

```bash
cd /home/user/schedCU/reimplement/cmd/cutout-sets
go build -o cutout-sets main.go
./cutout-sets
```

Or specify a custom ODS file:

```bash
./cutout-sets /path/to/your/schedule.ods
```

### Interactive Menu

```
1. Create new cutout set
   - Define name and description
   - Choose base set (modality/specialty/hospital/all)
   - Add exclusion rules
   - Preview results before saving

2. View all cutout sets
   - List all defined cutouts with member counts
   - Shows coverage status (24/7, weekday, weekend)
   - Indicates if cutout is empty

3. View cutout set details
   - Full definition (base + exclusions)
   - All matching study types
   - Coverage information

4. Delete cutout set
   - Remove unwanted cutouts

5. Show examples
   - See example cutout definitions

6. Export cutout definitions
   - Save to cutout_sets_export.json

7. Quit
```

## Examples

### Example 1: All MSK except MRI

```
Name: All MSK except MRI
Base Set: specialty = MSK
Exclusions: modality = MRI

Result:
  • CPMC CT MSK (if exists)
  • Allen X-Ray MSK (if exists)
  • NYPLH US MSK (if exists)

(MRI MSK is excluded)
```

**Use case**: Define musculoskeletal imaging that doesn't require MRI equipment

### Example 2: All CT except Neuro

```
Name: All CT except Neuro
Base Set: modality = CT
Exclusions: specialty = Neuro

Result:
  • CPMC CT Body
  • Allen CT Chest
  • NYPLH CT MSK

(CT Neuro is excluded)
```

**Use case**: CT coverage for non-neuro cases

### Example 3: All CPMC except Ultrasound

```
Name: All CPMC except Ultrasound
Base Set: hospital = CPMC
Exclusions: modality = US

Result:
  • CPMC CT Neuro
  • CPMC MRI Body
  • CPMC X-Ray Chest

(CPMC US is excluded)
```

**Use case**: CPMC services that require advanced equipment (not portable US)

### Example 4: All Imaging except Neuro and Body

```
Name: All except Neuro and Body
Base Set: all (*)
Exclusions:
  - specialty = Neuro
  - specialty = Body

Result:
  • CPMC CT Chest
  • Allen X-Ray Bone
  • NYPLH CT MSK

(All Neuro and Body studies excluded)
```

**Use case**: Specialty-specific coverage excluding general body and neuro

### Example 5: All MRI except CPMC and Allen

```
Name: All MRI except CPMC and Allen
Base Set: modality = MRI
Exclusions:
  - hospital = CPMC
  - hospital = Allen

Result:
  • NYPLH MRI Neuro
  • CHONY MRI Body

(CPMC MRI and Allen MRI excluded)
```

**Use case**: MRI coverage at satellite hospitals only

## Empty Cutouts (Gap Detection)

Some cutouts may have **zero members** - this is intentional and useful:

### Example: All Pediatric except CT

```
Name: All Pediatric except CT
Base Set: specialty = Pediatric
Exclusions: modality = CT

Result: EMPTY (no matching studies)

Interpretation:
  - Either: No pediatric imaging exists except CT
  - Or: Pediatric imaging doesn't exist at all in schedule
```

**Use case**: Identify coverage gaps - "We should have Pediatric MRI but don't"

### Why Empty Cutouts Matter

1. **Find gaps**: "We defined 'All MSK except MRI' but it's empty - maybe we only have MRI MSK?"
2. **Test hypotheses**: "What if we offered all modalities for MSK?"
3. **Plan expansion**: "We want 'All Neuro except CPMC' to distribute load"

## Cutout Definition Format

Cutouts are saved to `cutout_sets.json`:

```json
{
  "sets": [
    {
      "name": "All MSK except MRI",
      "description": "Musculoskeletal imaging without MRI",
      "base_set": {
        "dimension": "specialty",
        "value": "MSK"
      },
      "exclusions": [
        {
          "dimension": "modality",
          "value": "MRI"
        }
      ],
      "members": [
        "CPMC CT MSK",
        "Allen X-Ray MSK"
      ],
      "member_count": 2,
      "has_weekday": true,
      "has_weekend": false,
      "is_empty": false
    }
  ]
}
```

## Dimensions

### Base Set Dimensions

- **modality**: CT, MRI, X-Ray, US, NM, PET
- **specialty**: Neuro, Body, Chest, Bone, MSK, General
- **hospital**: CPMC, Allen, NYPLH, CHONY
- **all**: Use "*" to include everything

### Exclusion Dimensions

- **modality**: Exclude specific imaging modality
- **specialty**: Exclude specific body area/specialty
- **hospital**: Exclude specific hospital location

## Use Cases

### 1. Coverage Gap Analysis

Define hypothetical cutouts to find what's missing:
- "All Body imaging except CT" → Empty? We only have CT Body!
- "All hospitals except CPMC" → Low coverage? CPMC dominates!

### 2. Equipment Planning

Identify services that require specific equipment:
- "All imaging except US" → Needs fixed equipment
- "All except X-Ray and US" → Needs advanced imaging (CT/MRI)

### 3. Workflow Optimization

Group studies by workflow patterns:
- "All Neuro except Allen" → Centralize neuro at main hospitals
- "All CT except CHONY" → Adult hospitals only for CT

### 4. Load Distribution

Test different coverage models:
- "All CPMC except emergency modalities (CT, X-Ray)" → Elective MRI/US only
- "All Allen except Neuro" → Neuro centralized elsewhere

### 5. Conceptual Aggregations

Define sets that don't exist but could:
- "All Cardiac except MRI" → Empty now, but could add CT Cardiac, US Cardiac
- "All Spine except CPMC" → Test load distribution scenarios

## Tips

1. **Start Simple**: Begin with single exclusions before adding multiple

2. **Check for Empty**: Review empty cutouts - they reveal gaps or opportunities

3. **Use Descriptive Names**: "All MSK except MRI" is clearer than "MSK-MRI"

4. **Combine with Other Tools**:
   - Use `spanning-sets` to see what exists
   - Use `cutout-sets` to define what could exist with exclusions
   - Use `rank-spanning-sets` to see which cutouts users prefer

5. **Export for Analysis**: Export to JSON and analyze patterns programmatically

## Technical Details

- Written in Go (no Python dependencies)
- Reuses ODS parsing from spanning-sets tool
- Evaluates cutouts against actual coverage data on-the-fly
- Saves definitions to JSON for portability
- Supports multiple exclusions per cutout

## Future Enhancements

Potential improvements:
- Union/intersection logic (not just exclusion)
- Regular expressions for flexible matching
- Time-based cutouts ("All except overnight shifts")
- Percentage-based exclusions ("Top 80% by volume")
- Visual graph of cutout relationships

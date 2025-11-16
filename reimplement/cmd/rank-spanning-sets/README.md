# Spanning Set Preference Collector

Interactive tool for collecting user preferences about which spanning sets are most clinically useful.

## Purpose

This tool helps you build a ranking of spanning sets by comparing pairs and recording which ones you find more useful. Over time, this builds a preference model showing which aggregations are most valuable for clinical understanding.

## Features

- **Pairwise Comparison**: Compare two spanning sets side-by-side
- **Preference Recording**: Choose which set is more useful and optionally explain why
- **Ranking System**: View current rankings based on wins/losses
- **Statistics**: See which sets are most/least preferred
- **Export**: Save preferences to JSON for further analysis

## Usage

### Build and Run

```bash
cd /home/user/schedCU/reimplement/cmd/rank-spanning-sets
go build -o rank-spanning-sets main.go
./rank-spanning-sets
```

Or specify a custom ODS file:

```bash
./rank-spanning-sets /path/to/your/schedule.ods
```

### Interactive Menu

```
1. Compare two spanning sets (record preference)
   - Shows two random sets with descriptions
   - Choose A, B, or Skip
   - Optionally provide reason

2. View current rankings
   - Shows all sets ranked by preference score (wins - losses)
   - Displays win-loss record for each set

3. View all spanning sets
   - Lists all available sets grouped by dimension
   - Shows member counts

4. Show preference statistics
   - Total comparisons recorded
   - Top 5 most preferred sets
   - Recent comparisons with reasons

5. Export preferences to file
   - Saves to spanning_set_preferences_export.json

6. Quit
```

## How Preferences Work

### Comparison Process

1. Tool presents two spanning sets (e.g., "All CT" vs "CT Neuro - All Hospitals")
2. You choose which is **more useful for clinical understanding**
3. Optionally provide a reason (e.g., "More comprehensive" or "Better granularity")
4. Preference is saved immediately

### Scoring System

- **Wins**: Number of times a set was preferred over another
- **Losses**: Number of times a set was NOT preferred
- **Score**: Wins - Losses (higher is better)
- **Ranking**: Sorted by score, then by total comparisons (more data = higher rank)

### Example Comparison

```
A) All CT
   All CT scans across all hospitals and body parts
   Dimension: modality | Members: 7 | Coverage: 24/7
     • Allen CT Body
     • Allen CT Neuro
     • CPMC CT Body
     • CPMC CT Chest
     • CPMC CT Neuro
     • NYPLH Body CT
     • NYPLH Neuro CT

B) CT Neuro - All Hospitals
   All Neuro CT scans across all hospital locations
   Dimension: cross | Members: 3 | Coverage: 24/7
     • Allen CT Neuro
     • CPMC CT Neuro
     • NYPLH Neuro CT

Which spanning set is MORE USEFUL for clinical understanding?
(A) All CT
(B) CT Neuro - All Hospitals
(S) Skip this comparison

Your choice (A/B/S): B

Optional - Why is this more useful? (press Enter to skip): More specific to neuro workflow

✓ Preference saved!
```

## Output Files

### spanning_set_preferences.json

Auto-saved after each comparison. Contains all preference records:

```json
{
  "comparisons": [
    {
      "winner_id": "cross_CT Neuro",
      "loser_id": "modality_CT",
      "winner_name": "CT Neuro - All Hospitals",
      "loser_name": "All CT",
      "reason": "More specific to neuro workflow"
    }
  ]
}
```

### spanning_set_preferences_export.json

Manual export via menu option 5. Same format as above.

## Use Cases

### 1. Build Clinical Understanding
Discover which aggregations are most meaningful:
- "Do radiologists think more by modality or by hospital?"
- "Is 'All CT' more useful than 'CT Neuro - All Hospitals'?"

### 2. Train Summarization Models
Use preference data to:
- Weight which spanning sets appear first in summaries
- Learn patterns in what makes aggregations useful
- Build rules for automatic schedule summarization

### 3. Prioritize UI Display
Determine which spanning sets should be:
- Shown first in dashboards
- Used as default filters
- Highlighted in reports

### 4. Collect Qualitative Feedback
Understand WHY certain aggregations matter:
- "Cross-dimensional sets better align with clinical workflows"
- "Hospital-based sets more useful for staffing decisions"
- "Modality sets better for equipment planning"

## Tips

1. **Compare Diverse Sets**: The tool randomly selects pairs, but try to compare sets from different dimensions (modality vs. cross, hospital vs. specialty)

2. **Provide Reasons**: Adding reasons helps build qualitative understanding of preferences

3. **Regular Sessions**: Do 10-15 comparisons per session to build up preference data

4. **Review Rankings**: Check rankings periodically to see emerging patterns

5. **Export for Analysis**: Export preferences to analyze patterns (e.g., do cross-dimensional sets always win?)

## Technical Details

- Written in Go (no Python dependencies)
- Reuses ODS parsing from spanning-sets tool
- Preferences stored in JSON for portability
- Win/loss scores calculated on-the-fly from comparison history
- Random pairing (currently using simple modulo-based randomness)

## Future Enhancements

Potential improvements:
- Better randomization (crypto/rand)
- Weighted comparisons (compare sets with similar scores more often)
- ELO-style ranking system
- Export to CSV for statistical analysis
- Import preferences from other users to build consensus rankings
- Filter comparisons by dimension (e.g., only compare modality sets)

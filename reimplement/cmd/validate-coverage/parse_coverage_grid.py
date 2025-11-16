#!/usr/bin/env python3
"""
Coverage Grid Validation Tool for schedCU ODS Files

Parses ODS files organized as coverage grids where:
- Rows = Study types/modalities
- Columns = Shift positions
- Sheets = Time periods and day types
- 'x' = Coverage exists
"""

import zipfile
import xml.etree.ElementTree as ET
from collections import defaultdict

# ODS namespace mappings
NS = {
    'office': 'urn:oasis:names:tc:opendocument:xmlns:office:1.0',
    'table': 'urn:oasis:names:tc:opendocument:xmlns:table:1.0',
    'text': 'urn:oasis:names:tc:opendocument:xmlns:text:1.0',
}


def extract_text(cell):
    """Extract text content from ODS table cell"""
    texts = []
    for elem in cell.iter():
        if elem.tag == f"{{{NS['text']}}}p":
            if elem.text:
                texts.append(elem.text)
    return ' '.join(texts).strip()


def parse_coverage_grid(filepath):
    """Parse ODS coverage grid"""
    print("="*70)
    print("         SchedCU Coverage Grid Validation Tool")
    print("="*70)
    print()
    print(f"Opening file: {filepath}")

    try:
        with zipfile.ZipFile(filepath, 'r') as zf:
            content = zf.read('content.xml')
    except Exception as e:
        print(f"❌ Failed to read ODS file: {e}")
        return None

    print("✓ File opened successfully")

    # Parse XML
    try:
        root = ET.fromstring(content)
    except Exception as e:
        print(f"❌ Failed to parse XML: {e}")
        return None

    print("✓ XML parsed successfully\n")

    # Find all sheets
    body = root.find(f'.//{{{NS["office"]}}}body')
    spreadsheet = body.find(f'.//{{{NS["office"]}}}spreadsheet')
    tables = spreadsheet.findall(f'.//{{{NS["table"]}}}table')

    print(f"✓ Found {len(tables)} sheet(s)\n")

    # Data structure: coverage_data[sheet_name][study_type][shift_position] = has_coverage
    coverage_data = {}

    for table in tables:
        sheet_name = table.get(f'{{{NS["table"]}}}name', 'Unknown')
        print(f"Processing sheet: {sheet_name}")

        # Determine day type from sheet name
        is_weekend = 'weekend' in sheet_name.lower()
        day_type = 'Weekend' if is_weekend else 'Weekday'

        # Find all rows
        rows = table.findall(f'.//{{{NS["table"]}}}table-row')

        if not rows:
            continue

        # Get shift position headers from first row
        header_row = rows[0]
        header_cells = header_row.findall(f'.//{{{NS["table"]}}}table-cell')
        shift_positions = [extract_text(cell) for cell in header_cells]

        print(f"  Day type: {day_type}")
        print(f"  Shift positions: {[p for p in shift_positions if p]}")
        print(f"  Data rows: {len(rows) - 1}")

        # Parse data rows (study types)
        study_count = 0
        coverage_count = 0

        for row in rows[1:]:
            cells = row.findall(f'.//{{{NS["table"]}}}table-cell')

            if not cells:
                continue

            # First cell is the study type
            study_type = extract_text(cells[0])

            if not study_type:
                continue

            study_count += 1

            # Check each shift position column for coverage ('x' marker)
            for i, cell in enumerate(cells[1:], 1):
                cell_value = extract_text(cell).lower()

                if cell_value in ['x', 'yes', '1']:
                    coverage_count += 1
                    shift_position = shift_positions[i] if i < len(shift_positions) else f"Column{i}"

                    # Store coverage data
                    key = (sheet_name, day_type, study_type, shift_position)

                    if sheet_name not in coverage_data:
                        coverage_data[sheet_name] = {}
                    if study_type not in coverage_data[sheet_name]:
                        coverage_data[sheet_name][study_type] = {}

                    coverage_data[sheet_name][study_type][shift_position] = True

        print(f"  ✓ {study_count} study types, {coverage_count} coverage markers\n")

    return coverage_data


def extract_study_category(study_type):
    """Extract high-level category from study type"""
    study_lower = study_type.lower()

    if 'ct' in study_lower:
        if 'neuro' in study_lower:
            return 'CT Neuro'
        elif 'body' in study_lower:
            return 'CT Body'
        return 'CT'
    elif 'mr' in study_lower or 'mri' in study_lower:
        if 'neuro' in study_lower:
            return 'MRI Neuro'
        elif 'body' in study_lower:
            return 'MRI Body'
        return 'MRI'
    elif 'us' in study_lower or 'ultrasound' in study_lower:
        return 'Ultrasound'
    elif 'dx' in study_lower or 'x-ray' in study_lower or 'xray' in study_lower:
        return 'X-Ray'
    elif 'nm' in study_lower or 'nuclear' in study_lower:
        return 'Nuclear Medicine'
    elif 'fluoro' in study_lower:
        return 'Fluoroscopy'
    elif 'pet' in study_lower:
        return 'PET'
    else:
        return 'Other'


def analyze_coverage(coverage_data):
    """Analyze coverage by study category and day type"""
    print("="*70)
    print("                   COVERAGE ANALYSIS")
    print("="*70)
    print()

    # Aggregate by study category and day type
    category_coverage = defaultdict(lambda: {'Weekday': set(), 'Weekend': set()})
    study_type_coverage = defaultdict(lambda: {'Weekday': set(), 'Weekend': set()})

    for sheet_name, studies in coverage_data.items():
        # Determine day type
        is_weekend = 'weekend' in sheet_name.lower()
        day_type = 'Weekend' if is_weekend else 'Weekday'

        for study_type, positions in studies.items():
            # Track actual study types
            study_type_coverage[study_type][day_type].add(sheet_name)

            # Track categories
            category = extract_study_category(study_type)
            category_coverage[category][day_type].add(sheet_name)

    # Print study type coverage
    print("-"*70)
    print("Study Type Coverage (Detailed)")
    print("-"*70)
    print()

    for study_type in sorted(study_type_coverage.keys()):
        weekday_sheets = study_type_coverage[study_type]['Weekday']
        weekend_sheets = study_type_coverage[study_type]['Weekend']

        print(f"📋 {study_type}")
        print(f"   Weekday coverage: {len(weekday_sheets)} time period(s)")
        print(f"   Weekend coverage: {len(weekend_sheets)} time period(s)")

        if not weekday_sheets:
            print(f"   ❌ WARNING: NO WEEKDAY COVERAGE")
        if not weekend_sheets:
            print(f"   ❌ WARNING: NO WEEKEND COVERAGE")
        if weekday_sheets and weekend_sheets:
            print(f"   ✅ Has both weekday and weekend coverage")
        print()

    # Print category summary
    print("-"*70)
    print("Study Category Coverage (Summary)")
    print("-"*70)
    print()

    for category in sorted(category_coverage.keys()):
        weekday_sheets = category_coverage[category]['Weekday']
        weekend_sheets = category_coverage[category]['Weekend']

        print(f"📊 {category}")
        print(f"   Weekday coverage: {len(weekday_sheets)} time period(s)")
        print(f"   Weekend coverage: {len(weekend_sheets)} time period(s)")

        if not weekday_sheets:
            print(f"   ❌ WARNING: NO WEEKDAY COVERAGE")
        if not weekend_sheets:
            print(f"   ❌ WARNING: NO WEEKEND COVERAGE")
        if weekday_sheets and weekend_sheets:
            print(f"   ✅ Has both weekday and weekend coverage")
        print()

    # Find gaps
    print("-"*70)
    print("Coverage Gaps")
    print("-"*70)
    print()

    gaps = []

    # Check study types
    for study_type, coverage in study_type_coverage.items():
        if not coverage['Weekday']:
            gaps.append(f"{study_type} - Missing WEEKDAY coverage")
        if not coverage['Weekend']:
            gaps.append(f"{study_type} - Missing WEEKEND coverage")

    if gaps:
        print(f"❌ Found {len(gaps)} gap(s):\n")
        for i, gap in enumerate(gaps, 1):
            print(f"{i}. {gap}")
        print()
    else:
        print("✅ NO GAPS FOUND - All study types have both weekday and weekend coverage\n")

    # Summary
    print("="*70)
    print("                   VALIDATION SUMMARY")
    print("="*70)
    print()

    total_study_types = len(study_type_coverage)
    total_categories = len(category_coverage)

    print(f"📊 Statistics:")
    print(f"   Total study types: {total_study_types}")
    print(f"   Total categories: {total_categories}")
    print(f"   Coverage gaps: {len(gaps)}")
    print()

    if len(gaps) == 0:
        print("✅ VALIDATION PASSED")
        print("   All study types have coverage for both weekdays and weekends")
    else:
        print("❌ VALIDATION FAILED")
        print(f"   {len(gaps)} coverage gap(s) detected")
        print("   Review gaps listed above")
    print()


def main():
    import sys

    if len(sys.argv) > 1:
        filepath = sys.argv[1]
    else:
        filepath = '/home/user/schedCU/cuSchedNormalized.ods'

    coverage_data = parse_coverage_grid(filepath)

    if coverage_data:
        analyze_coverage(coverage_data)
    else:
        print("Failed to parse coverage data")
        sys.exit(1)


if __name__ == '__main__':
    main()

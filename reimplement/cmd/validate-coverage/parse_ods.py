#!/usr/bin/env python3
"""
Coverage Validation Tool for schedCU ODS Files

Parses ODS files and validates that every study type has coverage
for both weekdays and weekends.
"""

import sys
import zipfile
import xml.etree.ElementTree as ET
from collections import defaultdict
from datetime import datetime

# ODS namespace mappings
NS = {
    'office': 'urn:oasis:names:tc:opendocument:xmlns:office:1.0',
    'table': 'urn:oasis:names:tc:opendocument:xmlns:table:1.0',
    'text': 'urn:oasis:names:tc:opendocument:xmlns:text:1.0',
}

class Shift:
    def __init__(self, sheet_name, row_data):
        self.sheet_name = sheet_name
        self.date = row_data.get('date', '')
        self.shift = row_data.get('shift', '')
        self.position = row_data.get('position', '')
        self.location = row_data.get('location', '')
        self.staff_member = row_data.get('staff_member', '')
        self.specialty_constraint = row_data.get('specialty_constraint', '')
        self.study_type = row_data.get('study_type', '')
        self.required_qualification = row_data.get('required_qualification', '')

        # Determine if weekend
        self.is_weekend = self._is_weekend()

        # Extract study type from sheet name if not in data
        if not self.study_type:
            self.study_type = self._extract_study_type_from_sheet()

    def _is_weekend(self):
        """Determine if shift is on weekend based on sheet name or date"""
        sheet_lower = self.sheet_name.lower()
        if 'weekend' in sheet_lower:
            return True
        if 'weekday' in sheet_lower:
            return False

        # Try to parse date if available
        if self.date:
            try:
                # Try common date formats
                for fmt in ['%Y-%m-%d', '%m/%d/%Y', '%d/%m/%Y']:
                    try:
                        dt = datetime.strptime(self.date, fmt)
                        return dt.weekday() >= 5  # Saturday=5, Sunday=6
                    except:
                        continue
            except:
                pass

        return None  # Unknown

    def _extract_study_type_from_sheet(self):
        """Extract study type from sheet name"""
        sheet_lower = self.sheet_name.lower()
        if 'body' in sheet_lower:
            return 'Body'
        elif 'neuro' in sheet_lower:
            return 'Neuro'
        elif 'on' in sheet_lower and 'neuro' in sheet_lower:
            return 'Neuro'
        elif 'on' in sheet_lower and 'body' in sheet_lower:
            return 'Body'
        return 'General'

    def get_day_type(self):
        """Return 'Weekday', 'Weekend', or 'Unknown'"""
        if self.is_weekend is None:
            return 'Unknown'
        return 'Weekend' if self.is_weekend else 'Weekday'

    def __repr__(self):
        return f"Shift({self.date}, {self.shift}, {self.study_type}, {self.get_day_type()})"


def extract_text(cell):
    """Extract text content from ODS table cell"""
    texts = []
    for elem in cell.iter():
        if elem.tag == f"{{{NS['text']}}}p":
            if elem.text:
                texts.append(elem.text)
    return ' '.join(texts).strip()


def parse_ods(filepath):
    """Parse ODS file and extract all shifts"""
    print("="*70)
    print("           SchedCU Coverage Validation Tool")
    print("="*70)
    print()
    print(f"Opening file: {filepath}")

    try:
        with zipfile.ZipFile(filepath, 'r') as zf:
            content = zf.read('content.xml')
    except Exception as e:
        print(f"❌ Failed to read ODS file: {e}")
        sys.exit(1)

    print("✓ File opened successfully")

    # Parse XML
    try:
        root = ET.fromstring(content)
    except Exception as e:
        print(f"❌ Failed to parse XML: {e}")
        sys.exit(1)

    print("✓ XML parsed successfully\n")

    # Find all sheets
    body = root.find(f'.//{{{NS["office"]}}}body')
    spreadsheet = body.find(f'.//{{{NS["office"]}}}spreadsheet')
    tables = spreadsheet.findall(f'.//{{{NS["table"]}}}table')

    print(f"✓ Found {len(tables)} sheet(s)\n")

    all_shifts = []

    for table in tables:
        sheet_name = table.get(f'{{{NS["table"]}}}name', 'Unknown')
        print(f"Processing sheet: {sheet_name}")

        # Find all rows
        rows = table.findall(f'.//{{{NS["table"]}}}table-row')

        if not rows:
            continue

        # Get headers from first row
        header_row = rows[0]
        header_cells = header_row.findall(f'.//{{{NS["table"]}}}table-cell')
        headers = [extract_text(cell).lower().strip() for cell in header_cells]

        print(f"  Headers: {headers}")
        print(f"  Rows: {len(rows)}")

        # Parse data rows
        shifts_in_sheet = 0
        for row in rows[1:]:
            cells = row.findall(f'.//{{{NS["table"]}}}table-cell')

            # Extract row data
            row_data = {}
            for i, cell in enumerate(cells):
                if i < len(headers) and headers[i]:
                    value = extract_text(cell)
                    row_data[headers[i]] = value

            # Skip empty rows
            if not any(row_data.values()):
                continue

            # Create shift object
            shift = Shift(sheet_name, row_data)
            if shift.date or shift.staff_member:  # Has some valid data
                all_shifts.append(shift)
                shifts_in_sheet += 1

        print(f"  ✓ Extracted {shifts_in_sheet} shifts\n")

    print(f"✓ Total shifts parsed: {len(all_shifts)}\n")
    return all_shifts


def analyze_coverage(shifts):
    """Analyze coverage by study type and day type"""
    print("="*70)
    print("                   COVERAGE ANALYSIS")
    print("="*70)
    print()

    # Count by day type
    weekday_count = sum(1 for s in shifts if s.is_weekend == False)
    weekend_count = sum(1 for s in shifts if s.is_weekend == True)
    unknown_count = sum(1 for s in shifts if s.is_weekend is None)

    print(f"📊 Total Shifts: {len(shifts)}")
    print(f"   Weekday: {weekday_count}")
    print(f"   Weekend: {weekend_count}")
    if unknown_count > 0:
        print(f"   Unknown: {unknown_count}")
    print()

    # Group by study type and day type
    coverage_matrix = defaultdict(lambda: {'Weekday': 0, 'Weekend': 0, 'Unknown': 0})

    for shift in shifts:
        study_type = shift.study_type or 'Unspecified'
        day_type = shift.get_day_type()
        coverage_matrix[study_type][day_type] += 1

    # Print study type coverage
    print("-"*70)
    print("Study Type Coverage (Weekday vs Weekend)")
    print("-"*70)
    print()

    for study_type in sorted(coverage_matrix.keys()):
        counts = coverage_matrix[study_type]
        print(f"📋 {study_type}")
        print(f"   Weekday shifts: {counts['Weekday']}")
        print(f"   Weekend shifts: {counts['Weekend']}")
        if counts['Unknown'] > 0:
            print(f"   Unknown shifts: {counts['Unknown']}")

        # Check for gaps
        if counts['Weekday'] == 0:
            print(f"   ❌ WARNING: NO WEEKDAY COVERAGE")
        if counts['Weekend'] == 0:
            print(f"   ❌ WARNING: NO WEEKEND COVERAGE")
        if counts['Weekday'] > 0 and counts['Weekend'] > 0:
            print(f"   ✅ Has both weekday and weekend coverage")
        print()

    # Find gaps
    print("-"*70)
    print("Coverage Gaps")
    print("-"*70)
    print()

    gaps = []
    for study_type, counts in coverage_matrix.items():
        if counts['Weekday'] == 0:
            gaps.append(f"{study_type} - Missing WEEKDAY coverage")
        if counts['Weekend'] == 0:
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

    if len(gaps) == 0:
        print("✅ VALIDATION PASSED")
        print("   All study types have coverage for both weekdays and weekends")
    else:
        print("❌ VALIDATION FAILED")
        print(f"   {len(gaps)} coverage gap(s) detected")
        print("   Review gaps listed above")
    print()


def main():
    if len(sys.argv) > 1:
        filepath = sys.argv[1]
    else:
        filepath = '/home/user/schedCU/cuSchedNormalized.ods'

    shifts = parse_ods(filepath)
    analyze_coverage(shifts)


if __name__ == '__main__':
    main()

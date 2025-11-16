#!/usr/bin/env python3
"""
Hour-by-Hour Coverage Viewer

Shows exactly what's covered during a specific time period,
listing all study types and shift positions that are active.
"""

import sys
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


def parse_time_from_sheet_name(sheet_name):
    """Extract time period information from sheet name"""
    name_lower = sheet_name.lower()

    # Determine day type
    is_weekend = 'weekend' in name_lower
    day_type = 'Weekend' if is_weekend else 'Weekday'

    # Extract time range
    if '5' in name_lower and '6' in name_lower and 'pm' in name_lower:
        time_range = "5-6 PM"
        start_hour = 17
        end_hour = 18
    elif '6' in name_lower and '12' in name_lower and 'am' in name_lower:
        time_range = "6 PM to Midnight"
        start_hour = 18
        end_hour = 24
    elif '5' in name_lower and '12' in name_lower and 'am' in name_lower:
        time_range = "5 PM to Midnight"
        start_hour = 17
        end_hour = 24
    elif '10' in name_lower and 'midnight' in name_lower:
        time_range = "10 PM to Midnight"
        start_hour = 22
        end_hour = 24
    elif '12am' in name_lower and '1am' in name_lower:
        time_range = "Midnight to 1 AM"
        start_hour = 0
        end_hour = 1
    elif '1' in name_lower and '8' in name_lower and 'am' in name_lower:
        time_range = "1 AM to 8 AM"
        start_hour = 1
        end_hour = 8
    else:
        time_range = "Unknown"
        start_hour = None
        end_hour = None

    # Determine specialty
    specialty = None
    if 'body' in name_lower:
        specialty = 'Body'
    elif 'neuro' in name_lower:
        specialty = 'Neuro'

    return {
        'sheet_name': sheet_name,
        'day_type': day_type,
        'time_range': time_range,
        'start_hour': start_hour,
        'end_hour': end_hour,
        'specialty': specialty,
        'is_weekend': is_weekend,
    }


def parse_ods_coverage(filepath):
    """Parse ODS file and extract all coverage data"""
    try:
        with zipfile.ZipFile(filepath, 'r') as zf:
            content = zf.read('content.xml')
    except Exception as e:
        print(f"❌ Failed to read ODS file: {e}")
        return None

    try:
        root = ET.fromstring(content)
    except Exception as e:
        print(f"❌ Failed to parse XML: {e}")
        return None

    body = root.find(f'.//{{{NS["office"]}}}body')
    spreadsheet = body.find(f'.//{{{NS["office"]}}}spreadsheet')
    tables = spreadsheet.findall(f'.//{{{NS["table"]}}}table')

    all_coverage = []

    for table in tables:
        sheet_name = table.get(f'{{{NS["table"]}}}name', 'Unknown')
        time_info = parse_time_from_sheet_name(sheet_name)

        rows = table.findall(f'.//{{{NS["table"]}}}table-row')
        if not rows:
            continue

        # Get shift position headers from first row
        header_row = rows[0]
        header_cells = header_row.findall(f'.//{{{NS["table"]}}}table-cell')
        shift_positions = [extract_text(cell) for cell in header_cells]

        # Parse data rows (study types)
        for row in rows[1:]:
            cells = row.findall(f'.//{{{NS["table"]}}}table-cell')
            if not cells:
                continue

            # First cell is the study type
            study_type = extract_text(cells[0])
            if not study_type:
                continue

            # Check each shift position column for coverage
            for i, cell in enumerate(cells[1:], 1):
                cell_value = extract_text(cell).lower()
                if cell_value in ['x', 'yes', '1']:
                    shift_position = shift_positions[i] if i < len(shift_positions) else f"Column{i}"

                    all_coverage.append({
                        'study_type': study_type,
                        'shift_position': shift_position,
                        'day_type': time_info['day_type'],
                        'time_range': time_info['time_range'],
                        'start_hour': time_info['start_hour'],
                        'end_hour': time_info['end_hour'],
                        'specialty': time_info['specialty'],
                        'is_weekend': time_info['is_weekend'],
                        'sheet_name': sheet_name,
                    })

    return all_coverage


def show_coverage_at_hour(coverage_data, hour, is_weekend=False):
    """Show what's covered at a specific hour"""
    day_type = "Weekend" if is_weekend else "Weekday"

    # Filter coverage for this hour
    relevant_coverage = [
        c for c in coverage_data
        if c['start_hour'] is not None
        and c['end_hour'] is not None
        and c['start_hour'] <= hour < c['end_hour']
        and c['is_weekend'] == is_weekend
    ]

    if not relevant_coverage:
        print(f"\n❌ No coverage data found for {day_type} at hour {hour}")
        return

    print("\n" + "=" * 70)
    print(f"         COVERAGE AT {format_hour(hour)} ({day_type})")
    print("=" * 70)
    print()

    # Group by time period/sheet
    by_sheet = defaultdict(list)
    for c in relevant_coverage:
        by_sheet[c['sheet_name']].append(c)

    # Show each time period
    for sheet_name in sorted(by_sheet.keys()):
        items = by_sheet[sheet_name]
        time_range = items[0]['time_range']
        specialty = items[0]['specialty']

        specialty_label = f" ({specialty})" if specialty else ""
        print(f"📅 Time Period: {time_range}{specialty_label}")
        print("─" * 70)
        print()

        # Group by shift position
        by_position = defaultdict(list)
        for item in items:
            by_position[item['shift_position']].append(item['study_type'])

        for position in sorted(by_position.keys()):
            studies = sorted(by_position[position])
            print(f"  👤 {position} Position:")
            for study in studies:
                print(f"     • {study}")
            print()

        print()

    # Summary statistics
    print("─" * 70)
    print("📊 SUMMARY:")
    print()

    total_studies = len(set(c['study_type'] for c in relevant_coverage))
    total_positions = len(set(c['shift_position'] for c in relevant_coverage))
    total_assignments = len(relevant_coverage)

    print(f"  • Total unique study types covered: {total_studies}")
    print(f"  • Total shift positions staffed: {total_positions}")
    print(f"  • Total coverage assignments: {total_assignments}")
    print()

    # List all unique study types
    print("  📋 All study types covered during this hour:")
    all_studies = sorted(set(c['study_type'] for c in relevant_coverage))
    for study in all_studies:
        print(f"     • {study}")
    print()


def format_hour(hour):
    """Format hour as readable time"""
    if hour == 0:
        return "Midnight (12:00 AM)"
    elif hour < 12:
        return f"{hour}:00 AM"
    elif hour == 12:
        return "Noon (12:00 PM)"
    else:
        return f"{hour - 12}:00 PM"


def list_available_hours(coverage_data):
    """Show what hours have coverage data"""
    print("\n" + "=" * 70)
    print("         AVAILABLE COVERAGE HOURS")
    print("=" * 70)
    print()

    # Collect all hours
    weekday_hours = set()
    weekend_hours = set()

    for c in coverage_data:
        if c['start_hour'] is not None and c['end_hour'] is not None:
            for h in range(c['start_hour'], c['end_hour']):
                if c['is_weekend']:
                    weekend_hours.add(h)
                else:
                    weekday_hours.add(h)

    print("📅 WEEKDAY HOURS:")
    for h in sorted(weekday_hours):
        print(f"  • {format_hour(h)}")
    print()

    print("📅 WEEKEND HOURS:")
    for h in sorted(weekend_hours):
        print(f"  • {format_hour(h)}")
    print()


def interactive_viewer(coverage_data):
    """Interactive mode to browse coverage by hour"""
    print("\n" + "=" * 70)
    print("         INTERACTIVE COVERAGE VIEWER")
    print("=" * 70)
    print()
    print("Commands:")
    print("  <hour> weekday   - Show weekday coverage at hour (0-23)")
    print("  <hour> weekend   - Show weekend coverage at hour (0-23)")
    print("  list             - List available hours")
    print("  examples         - Show example queries")
    print("  quit             - Exit")
    print()

    while True:
        try:
            cmd = input("📍 Enter command: ").strip().lower()

            if cmd == 'quit' or cmd == 'exit' or cmd == 'q':
                break

            elif cmd == 'list':
                list_available_hours(coverage_data)

            elif cmd == 'examples':
                print("\nExample queries:")
                print("  6 weekday    - Show weekday coverage at 6 PM")
                print("  18 weekday   - Show weekday coverage at 6 PM (24-hour)")
                print("  2 weekend    - Show weekend coverage at 2 AM")
                print("  22 weekday   - Show weekday coverage at 10 PM")
                print()

            else:
                parts = cmd.split()
                if len(parts) == 2:
                    try:
                        hour = int(parts[0])
                        day_type = parts[1]

                        if hour < 0 or hour > 23:
                            print("❌ Hour must be between 0-23")
                            continue

                        if day_type not in ['weekday', 'weekend']:
                            print("❌ Day type must be 'weekday' or 'weekend'")
                            continue

                        is_weekend = (day_type == 'weekend')
                        show_coverage_at_hour(coverage_data, hour, is_weekend)

                        # Prompt for user's summary
                        print("=" * 70)
                        print("💭 HOW WOULD YOU SUMMARIZE THIS COVERAGE?")
                        print("=" * 70)
                        print()
                        summary = input("Your summary: ").strip()
                        if summary:
                            print(f"\n✓ Recorded: \"{summary}\"")
                            print()

                            # Save to examples file
                            with open('coverage_examples.txt', 'a') as f:
                                f.write(f"Hour: {format_hour(hour)} ({day_type})\n")
                                f.write(f"Summary: {summary}\n")
                                f.write("-" * 70 + "\n")
                            print("✓ Saved to coverage_examples.txt")
                        print()

                    except ValueError:
                        print("❌ Invalid hour (must be a number 0-23)")
                else:
                    print("❌ Invalid command. Try 'examples' for help.")

        except KeyboardInterrupt:
            print("\n\nExiting...")
            break
        except EOFError:
            break


def main():
    if len(sys.argv) > 1:
        filepath = sys.argv[1]
    else:
        filepath = '/home/user/schedCU/cuSchedNormalized.ods'

    print("\n📊 Loading coverage data from:", filepath)
    coverage_data = parse_ods_coverage(filepath)

    if not coverage_data:
        print("Failed to load coverage data")
        return

    print(f"✓ Loaded {len(coverage_data)} coverage assignments")

    # Start interactive mode
    interactive_viewer(coverage_data)


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""
Study Type Description Generator

Translates technical study type codes into plain English descriptions
that explain what each study is and where it's performed.

Examples:
- "CPMC CT Neuro" → "Brain and spine CT scans at California Pacific Medical Center"
- "Allen MR Body" → "Body MRI scans at Allen Hospital"
- "NYPLH DX Chest/Abd" → "Chest and abdominal X-rays at NewYork-Presbyterian Lower Manhattan"
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

# Hospital full names
HOSPITALS = {
    'CPMC': 'California Pacific Medical Center',
    'Allen': 'Allen Hospital',
    'NYPLH': 'NewYork-Presbyterian Lower Manhattan Hospital',
    'CHONY': "Children's Hospital of New York",
}

# Modality descriptions
MODALITIES = {
    'CT': {
        'name': 'CT scan',
        'long_name': 'Computed Tomography',
        'description': 'detailed cross-sectional X-ray imaging',
        'verb': 'CT scans',
    },
    'MR': {
        'name': 'MRI scan',
        'long_name': 'Magnetic Resonance Imaging',
        'description': 'soft tissue visualization using magnetic fields',
        'verb': 'MRI scans',
    },
    'MRI': {
        'name': 'MRI scan',
        'long_name': 'Magnetic Resonance Imaging',
        'description': 'soft tissue visualization using magnetic fields',
        'verb': 'MRI scans',
    },
    'DX': {
        'name': 'X-ray',
        'long_name': 'Radiography',
        'description': 'standard X-ray imaging',
        'verb': 'X-rays',
    },
    'US': {
        'name': 'Ultrasound',
        'long_name': 'Ultrasonography',
        'description': 'real-time imaging using sound waves',
        'verb': 'ultrasound scans',
    },
    'NM': {
        'name': 'Nuclear Medicine scan',
        'long_name': 'Nuclear Medicine',
        'description': 'functional imaging using radioactive tracers',
        'verb': 'nuclear medicine scans',
    },
    'PET': {
        'name': 'PET scan',
        'long_name': 'Positron Emission Tomography',
        'description': 'metabolic imaging using radioactive tracers',
        'verb': 'PET scans',
    },
}

# Study area descriptions
STUDY_AREAS = {
    'Neuro': {
        'name': 'Brain and spine',
        'long_name': 'Neurological',
        'description': 'head, brain, spine, and nervous system',
        'examples': ['brain', 'spine', 'head', 'neck'],
    },
    'Body': {
        'name': 'Body',
        'long_name': 'Body imaging',
        'description': 'chest, abdomen, pelvis, and internal organs',
        'examples': ['chest', 'abdomen', 'pelvis', 'organs'],
    },
    'Chest/Abd': {
        'name': 'Chest and abdomen',
        'long_name': 'Thoracic and abdominal',
        'description': 'chest, lungs, heart, and abdominal organs',
        'examples': ['chest', 'lungs', 'heart', 'abdomen'],
    },
    'Chest': {
        'name': 'Chest',
        'long_name': 'Thoracic',
        'description': 'chest, lungs, and heart',
        'examples': ['chest', 'lungs', 'heart'],
    },
    'Bone': {
        'name': 'Bone and skeletal',
        'long_name': 'Musculoskeletal',
        'description': 'bones, joints, and skeletal structure',
        'examples': ['bones', 'fractures', 'joints', 'spine'],
    },
}


def extract_text(cell):
    """Extract text content from ODS table cell"""
    texts = []
    for elem in cell.iter():
        if elem.tag == f"{{{NS['text']}}}p":
            if elem.text:
                texts.append(elem.text)
    return ' '.join(texts).strip()


def parse_study_type(study_type_code):
    """Parse a study type code and return structured information"""

    # Extract hospital
    hospital_code = None
    hospital_name = None
    for code, name in HOSPITALS.items():
        if code in study_type_code:
            hospital_code = code
            hospital_name = name
            break

    # Extract modality
    modality_code = None
    modality_info = None
    for code, info in MODALITIES.items():
        if code in study_type_code.upper():
            modality_code = code
            modality_info = info
            break

    # Extract study area
    study_area_code = None
    study_area_info = None
    for area, info in STUDY_AREAS.items():
        if area in study_type_code:
            study_area_code = area
            study_area_info = info
            break

    return {
        'original': study_type_code,
        'hospital_code': hospital_code,
        'hospital_name': hospital_name,
        'modality_code': modality_code,
        'modality_info': modality_info,
        'study_area_code': study_area_code,
        'study_area_info': study_area_info,
    }


def describe_study_type(study_type_code, format='short'):
    """Generate human-readable description of a study type"""

    parsed = parse_study_type(study_type_code)

    hospital = parsed['hospital_name'] or 'Unknown Hospital'
    modality = parsed['modality_info'] or {'verb': 'imaging studies'}
    area = parsed['study_area_info'] or {'name': 'general'}

    if format == 'short':
        # "Brain CT scans at CPMC"
        return f"{area['name']} {modality['verb']} at {parsed['hospital_code'] or hospital}"

    elif format == 'medium':
        # "Brain and spine CT scans at California Pacific Medical Center"
        return f"{area['name']} {modality['verb']} at {hospital}"

    elif format == 'long':
        # "Brain and spine CT scans (Computed Tomography) at California Pacific Medical Center - detailed cross-sectional X-ray imaging of head, brain, spine, and nervous system"
        modality_desc = f"{modality['verb']} ({modality.get('long_name', '')})"
        area_desc = area.get('description', area['name'])
        return f"{area['name']} {modality_desc} at {hospital} - {modality.get('description', '')} of {area_desc}"

    elif format == 'ultra_short':
        # "Brain CT - CPMC"
        return f"{area['name']} {modality.get('name', 'scan')} - {parsed['hospital_code']}"

    elif format == 'patient_friendly':
        # "CT scans of your brain and spine at California Pacific Medical Center"
        return f"{modality['name']}s of your {area['name'].lower()} at {hospital}"

    else:
        return study_type_code


def extract_all_study_types(filepath):
    """Extract all unique study types from ODS file"""
    try:
        with zipfile.ZipFile(filepath, 'r') as zf:
            content = zf.read('content.xml')
    except Exception as e:
        print(f"❌ Failed to read ODS file: {e}")
        return set()

    try:
        root = ET.fromstring(content)
    except Exception as e:
        print(f"❌ Failed to parse XML: {e}")
        return set()

    body = root.find(f'.//{{{NS["office"]}}}body')
    spreadsheet = body.find(f'.//{{{NS["office"]}}}spreadsheet')
    tables = spreadsheet.findall(f'.//{{{NS["table"]}}}table')

    study_types = set()

    for table in tables:
        rows = table.findall(f'.//{{{NS["table"]}}}table-row')
        if not rows:
            continue

        # Skip header row
        for row in rows[1:]:
            cells = row.findall(f'.//{{{NS["table"]}}}table-cell')
            if not cells:
                continue

            # First cell is the study type
            study_type = extract_text(cells[0])
            if study_type:
                study_types.add(study_type)

    return study_types


def generate_study_glossary(filepath, format='medium'):
    """Generate a complete glossary of all study types"""

    print("=" * 70)
    print("              STUDY TYPE GLOSSARY")
    print("=" * 70)
    print()
    print(f"Reading from: {filepath}")
    print()

    study_types = extract_all_study_types(filepath)

    if not study_types:
        print("No study types found")
        return

    print(f"Found {len(study_types)} unique study types\n")
    print("=" * 70)
    print()

    # Group by hospital
    by_hospital = defaultdict(list)
    for study_type in sorted(study_types):
        parsed = parse_study_type(study_type)
        hospital = parsed['hospital_name'] or 'Other'
        by_hospital[hospital].append(study_type)

    # Print grouped by hospital
    for hospital in sorted(by_hospital.keys()):
        print(f"🏥 {hospital}")
        print("─" * 70)
        print()

        for study_type in sorted(by_hospital[hospital]):
            description = describe_study_type(study_type, format=format)
            print(f"  • {study_type}")
            print(f"    → {description}")
            print()

        print()


def generate_quick_reference(filepath):
    """Generate a quick reference table"""

    study_types = extract_all_study_types(filepath)

    print("=" * 70)
    print("              QUICK REFERENCE GUIDE")
    print("=" * 70)
    print()

    # Create table
    print(f"{'Code':<25} {'What It Means':<45}")
    print("─" * 70)

    for study_type in sorted(study_types):
        description = describe_study_type(study_type, format='short')
        print(f"{study_type:<25} {description:<45}")

    print()


def generate_patient_guide(filepath):
    """Generate patient-friendly descriptions"""

    print("=" * 70)
    print("        PATIENT GUIDE: Understanding Your Imaging Study")
    print("=" * 70)
    print()

    study_types = extract_all_study_types(filepath)

    # Group by modality
    by_modality = defaultdict(list)
    for study_type in sorted(study_types):
        parsed = parse_study_type(study_type)
        modality = parsed['modality_info']['name'] if parsed['modality_info'] else 'Other'
        by_modality[modality].append(study_type)

    for modality in sorted(by_modality.keys()):
        print(f"📋 {modality}s")
        print("─" * 70)
        print()

        for study_type in sorted(by_modality[modality]):
            description = describe_study_type(study_type, format='patient_friendly')
            parsed = parse_study_type(study_type)

            print(f"  {description}")

            # Add what to expect
            if parsed['modality_code'] == 'CT':
                print(f"     What to expect: Lie on a table that slides through a donut-shaped machine")
            elif parsed['modality_code'] in ['MR', 'MRI']:
                print(f"     What to expect: Lie still in a tube-shaped scanner (can be noisy)")
            elif parsed['modality_code'] == 'DX':
                print(f"     What to expect: Stand or lie down while images are taken")
            elif parsed['modality_code'] == 'US':
                print(f"     What to expect: Gel applied to skin, handheld device moved over area")

            print()

        print()


def main():
    if len(sys.argv) > 1:
        filepath = sys.argv[1]
    else:
        filepath = '/home/user/schedCU/cuSchedNormalized.ods'

    print("\n")

    # Generate all formats
    generate_study_glossary(filepath, format='medium')
    print("\n")
    generate_quick_reference(filepath)
    print("\n")
    generate_patient_guide(filepath)


if __name__ == '__main__':
    main()

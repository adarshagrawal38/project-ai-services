#!/usr/bin/env python3
"""
Updates image tags in values.yaml files when a Makefile is modified.

This script reads the IMAGE and TAG values from a specified Makefile and updates
all corresponding values.yaml files based on the COMPONENTS mapping from the
check_image_names.py script.

Usage:
    python hack/update_image_tags.py --makefile services/chatbot/Makefile --registry icr.io/ai-services-cicd
    python hack/update_image_tags.py -m services/digitize/Makefile -r icr.io/ai-services-cicd
"""

import argparse
import re
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Optional

# Import shared component mapping and helpers from .github/scripts/common.py
sys.path.insert(0, str(Path(__file__).parent.parent / ".github" / "scripts"))
from common import COMPONENTS, get_makefile_info  # noqa: E402


def get_modified_makefiles(base_branch: str = "origin/main") -> List[str]:
    """
    Get list of modified Makefiles compared to base branch using git diff.
    
    Args:
        base_branch: The base branch to compare against (default: origin/main)
    
    Returns:
        List of modified Makefile paths that are in COMPONENTS mapping
    """
    try:
        # Run git diff to get modified files
        result = subprocess.run(
            ["git", "diff", "--name-only", base_branch],
            capture_output=True,
            text=True,
            check=True
        )
        
        # Get all modified files
        modified_files = result.stdout.strip().split('\n')
        
        # Filter for Makefiles that are in our COMPONENTS mapping
        modified_makefiles = []
        for file_path in modified_files:
            if file_path in COMPONENTS:
                modified_makefiles.append(file_path)
        
        return modified_makefiles
    
    except subprocess.CalledProcessError as e:
        print(f"❌ Error running git diff: {e}")
        return []
    except Exception as e:
        print(f"❌ Error getting modified Makefiles: {e}")
        return []

def update_image_in_values_yaml(values_path: Path, key: str, registry: str, image_name: str, tag: str) -> bool:
    """
    Update the image reference in a values.yaml file for a specific key.
    
    Returns True if the file was modified, False otherwise.
    """
    content = values_path.read_text()
    new_image = f"{registry}/{image_name}:{tag}"
    
    # Find the section for the key and extract the image line within it
    pattern = re.compile(
        rf"^({key}:\s*\n(?:.*?\n)*?)(  image:\s*)(\S+)",
        re.MULTILINE,
    )
    
    match = pattern.search(content)
    if not match:
        print(f"  ⚠️  Could not find 'image:' in '{key}' section of {values_path}")
        return False
    
    old_image = match.group(3)
    
    # Check if update is needed
    if old_image == new_image:
        print(f"  ✅ Already up-to-date: {values_path} [{key}]")
        return False
    
    # Replace the image
    new_content = pattern.sub(rf"\1\g<2>{new_image}", content)
    
    # Write back to file
    values_path.write_text(new_content)
    print(f"  ✅ Updated: {values_path} [{key}]")
    print(f"     Old: {old_image}")
    print(f"     New: {new_image}")
    
    return True


def process_makefile(makefile_rel: str, registry: str, repo_root: Path) -> Tuple[int, int, int]:
    """
    Process a single Makefile and update its corresponding values.yaml files.
    
    Returns:
        Tuple of (updated_count, skipped_count, error_count)
    """
    makefile_path = repo_root / makefile_rel
    
    if not makefile_path.exists():
        print(f"  ❌ Makefile not found: {makefile_rel}")
        return 0, 0, 1
    
    # Extract IMAGE and TAG from Makefile
    try:
        image_name, tag = get_makefile_info(makefile_path)
    except ValueError as e:
        print(f"  ❌ Error: {e}")
        return 0, 0, 1
    
    print(f"\n📦 Processing: {makefile_rel}")
    print(f"   Image: {image_name}")
    print(f"   Tag:   {tag}")
    print(f"   Full:  {registry}/{image_name}:{tag}")
    print()
    
    # Get list of values.yaml files to update
    values_entries = COMPONENTS[makefile_rel]
    
    updated_count = 0
    skipped_count = 0
    error_count = 0
    
    for values_rel, values_key in values_entries:
        values_path = repo_root / values_rel
        
        if not values_path.exists():
            print(f"  ❌ File not found: {values_rel}")
            error_count += 1
            continue
        
        try:
            if update_image_in_values_yaml(values_path, values_key, registry, image_name, tag):
                updated_count += 1
            else:
                skipped_count += 1
        except Exception as e:
            print(f"  ❌ Error updating {values_rel}: {e}")
            error_count += 1
    
    return updated_count, skipped_count, error_count


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Update image tags in values.yaml files based on Makefile changes",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Update all values.yaml files for a specific Makefile
  %(prog)s --makefile services/chatbot/Makefile --registry icr.io/ai-services-cicd
  
  # Auto-detect modified Makefiles using git diff
  %(prog)s --registry icr.io/ai-services-cicd
  
  # Sync ALL Makefiles (update all values.yaml to match current Makefile TAGs)
  %(prog)s --sync --registry icr.io/ai-services-cicd
  
  # Auto-detect with custom base branch
  %(prog)s --registry icr.io/ai-services-cicd --base-branch origin/develop
        """
    )
    
    parser.add_argument(
        "-m", "--makefile",
        required=False,
        help="Path to the Makefile (relative to repository root). If not provided, auto-detects modified Makefiles using git diff (unless --sync is used)"
    )
    
    parser.add_argument(
        "--sync",
        action="store_true",
        help="Sync ALL Makefiles - iterate over all COMPONENTS and update values.yaml files to match current Makefile TAGs. Ignores --makefile and --base-branch flags"
    )
    
    parser.add_argument(
        "--base-branch",
        default="origin/main",
        help="Base branch for git diff comparison (default: origin/main). Only used when --makefile and --sync are not provided"
    )
    
    parser.add_argument(
        "-r", "--registry",
        required=True,
        help="Container registry URL (e.g., icr.io/ai-services-cicd)"
    )
    
    args = parser.parse_args()
    
    # Determine repository root (script is in hack/ directory)
    script_path = Path(__file__).resolve()
    repo_root = script_path.parent.parent
    
    # Determine which Makefiles to process
    makefiles_to_process = []
    
    if args.sync:
        # Sync mode: process ALL Makefiles in COMPONENTS
        print("🔄 SYNC MODE: Processing all Makefiles in COMPONENTS mapping...")
        makefiles_to_process = sorted(COMPONENTS.keys())
        mode = "sync"
    elif args.makefile:
        # Single Makefile specified
        makefile_path = repo_root / args.makefile
        
        if not makefile_path.exists():
            print(f"❌ Error: Makefile not found: {args.makefile}")
            return 1
        
        # Check if this Makefile is in our COMPONENTS mapping
        if args.makefile not in COMPONENTS:
            print(f"❌ Error: Makefile not found in COMPONENTS mapping: {args.makefile}")
            print(f"\nAvailable Makefiles:")
            for makefile in sorted(COMPONENTS.keys()):
                print(f"  - {makefile}")
            return 1
        
        makefiles_to_process = [args.makefile]
        mode = "single"
    else:
        # Auto-detect modified Makefiles
        print(f"🔍 Auto-detecting modified Makefiles using git diff against {args.base_branch}...")
        makefiles_to_process = get_modified_makefiles(args.base_branch)
        
        if not makefiles_to_process:
            print(f"\n✅ No modified Makefiles found in COMPONENTS mapping.")
            print(f"   Compared against: {args.base_branch}")
            print(f"\n   Tracked Makefiles:")
            for makefile in sorted(COMPONENTS.keys()):
                print(f"     - {makefile}")
            return 0
        
        mode = "auto"
    
    # Print header
    print("=" * 70)
    if mode == "sync":
        print(f"SYNC MODE: Syncing all {len(makefiles_to_process)} Makefiles with values.yaml")
        print("All values.yaml files will be synced to match current Makefile TAGs")
    elif mode == "single":
        print(f"Updating image tags for: {args.makefile}")
    else:
        print(f"Updating image tags for {len(makefiles_to_process)} modified Makefile(s)")
        print(f"Base branch: {args.base_branch}")
    print("=" * 70)
    print(f"Registry: {args.registry}")
    print("=" * 70)
    
    # Process each Makefile
    total_updated = 0
    total_skipped = 0
    total_errors = 0
    
    for makefile_rel in makefiles_to_process:
        updated, skipped, errors = process_makefile(
            makefile_rel, args.registry, repo_root
        )
        total_updated += updated
        total_skipped += skipped
        total_errors += errors
    
    print()
    print("=" * 70)
    print("Summary:")
    print("=" * 70)
    if mode in ["auto", "sync"]:
        print(f"Makefiles processed: {len(makefiles_to_process)}")
    print(f"Updated:             {total_updated}")
    print(f"Already current:     {total_skipped}")
    print(f"Errors:              {total_errors}")
    print("=" * 70)
    
    if total_errors > 0:
        return 1
    
    if total_updated > 0:
        if mode == "sync":
            print("\n✅ All Makefiles synced successfully!")
        else:
            print("\n✅ Image tags updated successfully!")
    else:
        print("\n✅ All image tags are already up-to-date!")
    
    return 0


if __name__ == "__main__":
    sys.exit(main())

# Made with Bob
#!/usr/bin/env python3
"""
Bumps the TAG value in a Makefile when files under its component root path are modified.

Detects which component root paths have changed files (via git diff), reads the current
TAG from the Makefile, increments the patch version, and writes it back.

Usage:
    python hack/bump_makefile_tag.py
    python hack/bump_makefile_tag.py --base-branch origin/main
    python hack/bump_makefile_tag.py --component services/chatbot
"""

import argparse
import re
import subprocess
import sys
from pathlib import Path
from typing import List, Optional, Tuple

# Import shared mappings from .github/scripts/common.py
sys.path.insert(0, str(Path(__file__).parent.parent / ".github" / "scripts"))
from common import EXCLUDED_PATHS, SOURCE_COMPONENTS  # noqa: E402


def get_changed_files(base_branch: str) -> List[str]:
    """Return list of files changed compared to base_branch."""
    try:
        result = subprocess.run(
            ["git", "diff", "--name-only", base_branch],
            capture_output=True,
            text=True,
            check=True,
        )
        return [l.strip() for l in result.stdout.strip().splitlines() if l.strip()]
    except subprocess.CalledProcessError as e:
        print(f"❌ Error running git diff: {e}")
        sys.exit(1)


def get_tag_from_makefile(makefile_path: Path, ref: Optional[str] = None, component_path: Optional[str] = None) -> Optional[str]:
    """
    Extract and resolve the TAG value from a Makefile.

    Args:
        makefile_path:   Path to the Makefile on disk (used when ref is None).
        ref:             Git ref to read from (e.g. 'origin/main'). When set,
                         the file is read from git rather than the working tree.
        component_path:  Component root path required for git show (e.g. 'services/chatbot').
    """
    try:
        if ref and component_path:
            result = subprocess.run(
                ["git", "show", f"{ref}:./{component_path}/Makefile"],
                capture_output=True,
                text=True,
                check=True,
            )
            content = result.stdout
        else:
            content = makefile_path.read_text()
    except subprocess.CalledProcessError as e:
        stderr = e.stderr.strip() if e.stderr else "unknown error"
        print(f"   ⚠️  Warning: Could not read {makefile_path} from {ref}: {stderr}")
        return None

    variables = {}
    for line in content.splitlines():
        m = re.match(r'^(\w+)\s*\??\s*=\s*(.+?)(?:\s*#.*)?$', line.strip())
        if m:
            variables[m.group(1)] = m.group(2).strip()

    tag_value = variables.get('TAG')
    if not tag_value:
        return None

    # Resolve variable references: $(VAR)
    pattern = r'\$\((\w+)\)'
    while re.search(pattern, tag_value):
        m = re.search(pattern, tag_value)
        if m:
            tag_value = tag_value.replace(m.group(0), variables.get(m.group(1), m.group(0)))

    return tag_value


def bump_patch(tag: str) -> str:
    """
    Increment the last numeric segment of a version string.

    Examples:
        "v0.0.32"  -> "v0.0.33"
        "0.11"     -> "0.12"
        "1.2.3-4"  -> "1.2.3-5"
    """
    # Find the last sequence of digits in the tag and increment it
    match = re.search(r'(\d+)(?=\D*$)', tag)
    if not match:
        raise ValueError(f"Cannot determine patch version to bump in tag: '{tag}'")
    old_num = match.group(1)
    new_num = str(int(old_num) + 1)
    # Replace only the last occurrence
    idx = tag.rfind(old_num)
    return tag[:idx] + new_num + tag[idx + len(old_num):]


def write_tag_to_makefile(makefile_path: Path, new_tag: str) -> None:
    """Replace the TAG assignment in a Makefile with new_tag."""
    content = makefile_path.read_text()
    # Match TAG?=... or TAG=... (optional ?), preserve any inline comment
    new_content = re.sub(
        r'^(TAG\s*\??\s*=\s*)(\S+)',
        lambda m: m.group(1) + new_tag,
        content,
        flags=re.MULTILINE,
    )
    if new_content == content:
        raise ValueError(f"Could not find TAG= line to update in {makefile_path}")
    makefile_path.write_text(new_content)


def components_changed_for(component_path: str, changed_files: List[str]) -> bool:
    """Return True if any changed file belongs to component_path (respecting EXCLUDED_PATHS)."""
    return any(
        f.startswith(f"{component_path}/")
        and not any(f.startswith(ex) for ex in EXCLUDED_PATHS)
        for f in changed_files
    )


def process_component(
    component_path: str,
    name: str,
    repo_root: Path,
    base_branch: str,
) -> Tuple[bool, Optional[str]]:
    """
    Bump TAG in component_path/Makefile.

    The new TAG is derived from the base branch version, not the working tree,
    so re-running the script on the same branch never double-bumps.

    Returns (bumped, error_message).
    """
    makefile_path = repo_root / component_path / "Makefile"

    if not makefile_path.exists():
        return False, f"❌ Makefile not found: {component_path}/Makefile"

    # Read TAG from the base branch (origin) to avoid double-bumping
    base_tag = get_tag_from_makefile(makefile_path, ref=base_branch, component_path=component_path)
    if base_tag is None:
        return False, f"❌ Could not read TAG from {base_branch} for {component_path}/Makefile"

    try:
        new_tag = bump_patch(base_tag)
    except ValueError as e:
        return False, f"❌ {e}"

    # Check if the working tree already has the bumped tag (idempotent)
    current_tag = get_tag_from_makefile(makefile_path)
    if current_tag == new_tag:
        print(f"  ⏭  {name}: TAG already bumped to {new_tag}  ({component_path}/Makefile)")
        return True, None

    write_tag_to_makefile(makefile_path, new_tag)
    print(f"  ✅ {name}: TAG {base_tag} → {new_tag}  ({component_path}/Makefile)")
    return True, None


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Bump Makefile TAG when component source files are modified",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Auto-detect changed components via git diff against origin/main
  %(prog)s

  # Use a different base branch
  %(prog)s --base-branch origin/develop

  # Bump a specific component regardless of git diff
  %(prog)s --component services/chatbot
        """,
    )
    parser.add_argument(
        "--base-branch",
        default="origin/main",
        help="Base branch for git diff (default: origin/main). Ignored when --component is set.",
    )
    parser.add_argument(
        "--component",
        default=None,
        help="Component root path to bump (e.g. services/chatbot). Skips git diff detection.",
    )
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parent.parent

    # --- Determine which components to bump ---
    if args.component:
        # Find matching entry in SOURCE_COMPONENTS
        entry = next(
            ((cp, name) for cp, name in SOURCE_COMPONENTS if cp == args.component),
            None,
        )
        if entry is None:
            print(f"❌ Component '{args.component}' not found in SOURCE_COMPONENTS.")
            print("\nAvailable components:")
            for cp, name in SOURCE_COMPONENTS:
                print(f"  {cp}  ({name})")
            return 1
        components_to_bump = [entry]
    else:
        print(f"🔍 Detecting changed components via git diff against {args.base_branch}...")
        changed_files = get_changed_files(args.base_branch)
        if not changed_files:
            print("ℹ️  No changed files detected.")
            return 0
        components_to_bump = [
            (cp, name)
            for cp, name in SOURCE_COMPONENTS
            if components_changed_for(cp, changed_files)
        ]
        if not components_to_bump:
            print("ℹ️  No tracked components have changed files.")
            return 0

    # --- Process ---
    print("=" * 70)
    print(f"Bumping TAG for {len(components_to_bump)} component(s)...")
    print("=" * 70)
    print()

    errors = []
    bumped = []

    for component_path, name in components_to_bump:
        ok, err = process_component(component_path, name, repo_root, args.base_branch)
        if err:
            errors.append(err)
        elif ok:
            bumped.append(name)

    print()
    print("=" * 70)
    print("Summary:")
    print("=" * 70)
    print(f"Bumped:  {len(bumped)}")
    print(f"Errors:  {len(errors)}")
    print("=" * 70)

    if errors:
        print()
        for err in errors:
            print(err)
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())

# Made with Bob

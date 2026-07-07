import re
from pathlib import Path
from typing import Tuple

# Registry used in values.yaml files
EXPECTED_REGISTRY = "icr.io/ai-services-cicd"

# ---------------------------------------------------------------------------
# Path helpers for COMPONENTS
# ---------------------------------------------------------------------------
_rag_app_base_path = "ai-services/assets/applications/rag"
_podman = "podman"
_openshift = "openshift"
_values_yaml = "values.yaml"


def _p(variant: str, deploy: str) -> str:
    """Build an application values.yaml path: rag[-variant]/deploy/values.yaml"""
    suffix = f"-{variant}" if variant else ""
    return f"{_rag_app_base_path}{suffix}/{deploy}/{_values_yaml}"


# ---------------------------------------------------------------------------
# Map of: makefile_path -> list of (values_yaml_path, values_key)
# values_key is the top-level key in values.yaml that contains the image reference
# ---------------------------------------------------------------------------
COMPONENTS = {
    "services/chatbot/Makefile": [
        (_p("", _podman), "backend"),
        (_p("", _openshift), "backend"),
        (_p("dev", _podman), "backend"),
        (_p("dev", _openshift), "backend"),
        (_p("cpu", _podman), "backend"),
        (f"ai-services/assets/services/chat/{_podman}/{_values_yaml}", "backend"),
    ],
    "services/digitize/Makefile": [
        (_p("", _podman), "digitize"),
        (_p("", _openshift), "digitize"),
        (_p("dev", _podman), "digitize"),
        (_p("dev", _openshift), "digitize"),
        (_p("cpu", _podman), "digitize"),
        (f"ai-services/assets/services/digitize/{_podman}/{_values_yaml}", "digitize"),
    ],
    "services/summarize/Makefile": [
        (_p("", _podman), "summarize"),
        (_p("", _openshift), "summarize"),
        (_p("dev", _podman), "summarize"),
        (_p("dev", _openshift), "summarize"),
        (_p("cpu", _podman), "summarize"),
        (f"ai-services/assets/services/summarize/{_podman}/{_values_yaml}", "summarize"),
    ],
    "services/similarity/Makefile": [
        (_p("", _podman), "similarity"),
        (_p("", _openshift), "similarity"),
        (_p("dev", _podman), "similarity"),
        (_p("dev", _openshift), "similarity"),
        (_p("cpu", _podman), "similarity"),
        (f"ai-services/assets/services/similarity/{_podman}/{_values_yaml}", "similarity"),
    ],
    "ui/chatbot/Makefile": [
        (_p("", _podman), "ui"),
        (_p("dev", _podman), "ui"),
        (_p("", _openshift), "ui"),
        (_p("dev", _openshift), "ui"),
        (_p("cpu", _podman), "ui"),
        (f"ai-services/assets/services/chat/{_podman}/{_values_yaml}", "ui"),
    ],
    "ui/digitize/Makefile": [
        (_p("", _openshift), "digitizeUi"),
        (_p("", _podman), "digitizeUi"),
        (_p("dev", _openshift), "digitizeUi"),
        (_p("dev", _podman), "digitizeUi"),
        (_p("cpu", _podman), "digitizeUi"),
        (f"ai-services/assets/services/digitize/{_podman}/{_values_yaml}", "digitizeUi"),
    ],
    "ui/catalog/Makefile": [
        (f"ai-services/assets/catalog/{_podman}/{_values_yaml}", "ui"),
    ],
    "ai-services/Makefile": [
        (f"ai-services/assets/catalog/{_podman}/{_values_yaml}", "backend"),
    ],
    "images/postgres/Makefile": [
        (f"ai-services/assets/catalog/{_podman}/{_values_yaml}", "db"),
        (_p("", _podman), "postgres"),
        (_p("", _openshift), "postgres"),
        (_p("dev", _podman), "postgres"),
        (_p("dev", _openshift), "postgres"),
        (_p("cpu", _podman), "postgres"),
    ],
    "images/litellm/Makefile": [
        (_p("cpu", _podman), "litellm"),
        (f"ai-services/assets/components/llm/watsonx/{_podman}/{_values_yaml}", ""),
    ],
    "images/caddy/Makefile": [
        (f"ai-services/assets/catalog/{_podman}/{_values_yaml}", "caddy"),
    ],
    # images/tools/Makefile uses a different registry (icr.io/ai-services-private)
    # and stores its reference in a Go file — validation is skipped for now.
}

# ---------------------------------------------------------------------------
# Map of: component root path -> display name
# Used by check_makefile_version_bump.py and bump_makefile_tag.py
# ---------------------------------------------------------------------------
SOURCE_COMPONENTS = [
    # Services Images
    ("services/chatbot", "chatbot-service"),
    ("services/digitize", "digitize-service"),
    ("services/similarity", "similarity-service"),
    ("services/summarize", "summarize-service"),
    # Images
    ("images/service-base", "service-base"),
    ("images/postgres", "postgres"),
    ("images/caddy", "caddy"),
    ("images/litellm", "litellm"),
    ("images/tools", "tools"),
    # UI Images
    ("ui/chatbot", "chatbot-ui"),
    ("ui/digitize", "digitize-ui"),
    ("ui/catalog", "catalog-ui"),
    # Ai Services
    ("ai-services", "ai-services"),
]

# Paths that don't require version bumps when modified
EXCLUDED_PATHS = [
    "ai-services/assets/catalog",
    "ai-services/assets/bootstrap",
]


# ---------------------------------------------------------------------------
# Shared helper
# ---------------------------------------------------------------------------
def get_makefile_info(makefile_path: Path) -> Tuple[str, str]:
    """Extract IMAGE= and TAG= values from a Makefile, resolving any variable references."""
    content = makefile_path.read_text()

    variables = {}
    for line in content.split('\n'):
        var_match = re.match(r'^(\w+)\s*\??\s*=\s*(.+?)(?:\s*#.*)?$', line.strip())
        if var_match:
            variables[var_match.group(1)] = var_match.group(2).strip()

    image = variables.get('IMAGE')
    if not image:
        raise ValueError(f"Could not find IMAGE= in {makefile_path}")

    tag_value = variables.get('TAG')
    if not tag_value:
        raise ValueError(f"Could not find TAG= in {makefile_path}")

    def resolve_variables(value: str) -> str:
        pattern = r'\$\((\w+)\)'
        while re.search(pattern, value):
            match = re.search(pattern, value)
            if match:
                var_replacement = variables.get(match.group(1), match.group(0))
                value = value.replace(match.group(0), var_replacement)
        return value

    return image, resolve_variables(tag_value)

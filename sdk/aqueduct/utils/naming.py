import re

from aqueduct.error import InvalidUserArgumentException


def sanitize_artifact_name(name: str) -> str:
    """Strip out whitespace before and after user-supplied artifact names."""
    if len(name) == 0:
        raise InvalidUserArgumentException("Artifact name cannot be empty.")
    return name.strip()


def default_artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


def bump_artifact_suffix(artifact_name: str) -> str:
    """Assumption: the artifact name has been sanitized already."""

    # No need to do any fancy regex parsing if the artifact name doesn't end with a ')'.
    if artifact_name[-1] != ")":
        return artifact_name + " (1)"

    # Check if the last few characters of artifact_name match the pattern "([0-9]+)"
    suffix_match = re.findall(r" \([0-9]+\)$", artifact_name)
    suffix_idx = 1
    if len(suffix_match) > 0:
        val_with_parens = suffix_match[-1][1:]  # cut off the leading space
        val = val_with_parens[1:-1]  # cut off the parens

        # Remove the suffix to replace it with the new one.
        artifact_name = artifact_name[: -len(suffix_match[-1])]
        suffix_idx = int(val) + 1

    return artifact_name + f" ({suffix_idx})"

# The TAG for 'previous table' when the user specifies a chained query.
import re
from typing import List

from aqueduct.resources.parameters import BUILT_IN_EXPANSIONS, BUILTIN_TAG_PATTERN, USER_TAG_PATTERN

PREV_TABLE_TAG = "$"


def _replace_builtin_tags(query: str) -> str:
    """Expands any builtin tags found in the raw query, eg. {{ today }}."""
    matches = re.findall(BUILTIN_TAG_PATTERN, query)
    for match in matches:
        tag_name = match.strip(" {}")
        if tag_name in BUILT_IN_EXPANSIONS:
            expansion_func = BUILT_IN_EXPANSIONS[tag_name]
            query = query.replace(match, expansion_func())
    return query


def _replace_param_sql_placeholders(query: str, parameter_vals: List[str]) -> str:
    """Replaces any user-defined placeholders in the query with the corresponding parameter value.

    Assumes that we've already validated that every parameter value has a corresponding placeholder in the query.
    """
    for i in range(len(parameter_vals)):
        query = query.replace("$" + str(i + 1), parameter_vals[i])
    return query


def _replace_parameterized_user_strings(user_defined_string: str, parameter_vals: List[str]) -> str:
    """Expands any parameters interpolated in the given string with '{  }' syntax."""
    matches = re.findall(USER_TAG_PATTERN, user_defined_string)

    if len(matches) != len(parameter_vals):
        raise Exception(
            "Mismatch between number of parameters (%s) and number of placeholders (%s) in the user-defined string. "
            % (len(parameter_vals), len(matches))
        )
    for idx, match in enumerate(matches):
        user_defined_string = user_defined_string.replace(match, parameter_vals[idx])
    return user_defined_string

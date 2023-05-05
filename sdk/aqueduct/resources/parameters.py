import re
from datetime import date
from typing import List

from aqueduct.artifacts.base_artifact import BaseArtifact

# Regular Expression that matches any substring appearance with
# "{{ }}" and a word inside with optional space in front or after
# Potential Matches: "{{today}}", "{{ today  }}""
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException

TAG_PATTERN = r"{{\s*[\w-]+\s*}}"


def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


# A dictionary of built-in tags to their replacement string functions.
BUILT_IN_EXPANSIONS = {
    "today": replace_today,
}


def _validate_builtin_expansions(queries: List[str]) -> None:
    """Check that if {{ }} syntax is used, the keyword is a valid, builtin expansion, such as {{today}}."""
    for query in queries:
        matches = re.findall(TAG_PATTERN, query)
        for match in matches:
            tag_name = match.strip(" {}")
            if tag_name not in BUILT_IN_EXPANSIONS:
                raise InvalidUserActionException(
                    "`%s` is not a valid Aqueduct placeholder. Valid placeholders are: [%s]. If you're trying to parameterize "
                    "this query please pass parameters in explicitly and use our $1, $2 etc. query syntax."
                    % (tag_name, ", ".join(BUILT_IN_EXPANSIONS.keys()))
                )


def _validate_parameters(queries: List[str], parameters: List[BaseArtifact]) -> None:
    """Validates that:
    1) All parameters are strings
    2) All supplied parameters have corresponding placeholders in the query/queries. For example, if there are
       two parameters, then $1 and $2 must exist in the query!
    """
    for param in parameters:
        if param.type() != ArtifactType.STRING:
            raise InvalidUserArgumentException(
                "The parameter `%s` must be defined as a string. Instead, got type %s"
                % (param.name(), param.type())
            )

    # String all the queries together, since all we're doing is some string matching.
    # Separate them with something to avoid accidental $(num) matches.
    concatenated_queries = ", ".join(queries)
    for i in range(len(parameters)):
        placeholder = "$%d" % (i + 1)
        if placeholder not in concatenated_queries:
            raise InvalidUserArgumentException(
                "Unused parameter `%s`. The query/queries `%s` must contain the placeholder %s."
                % (parameters[i].name(), concatenated_queries, placeholder)
            )

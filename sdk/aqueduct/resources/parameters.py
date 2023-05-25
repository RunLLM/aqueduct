import re
import uuid
from datetime import date
from typing import List, Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException

# Regular Expression that matches any substring appearance with
# "{{ }}" and a word inside with optional space in front or after
# Potential Matches: "{{today}}", "{{ today  }}""
#
# This is only expected to be used in SQL queries.
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG

BUILTIN_TAG_PATTERN = r"{{\s*[\w-]+\s*}}"


def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


# A dictionary of built-in tags to their replacement string functions.
BUILT_IN_EXPANSIONS = {
    "today": replace_today,
}


def _validate_builtin_expansions(queries: List[str]) -> None:
    """Check that if {{ }} syntax is used, the keyword is a valid, builtin expansion, such as {{today}}."""
    for query in queries:
        matches = re.findall(BUILTIN_TAG_PATTERN, query)
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
        _validate_artifact_is_string(param)

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


# Regular Expression that matches any substring appearance with "{ }" and a word inside
# with optional space in front or after. Example string match: "{directory_path}/{file_name}".
#
# This is *not* expected to be used in SQL queries, but instead in table names, file paths, etc.
USER_TAG_PATTERN = r"{\s*[\w-]+\s*}"


def _fetch_param_artifact_ids_embedded_in_string(
    dag: DAG, user_supplied_str: str
) -> List[uuid.UUID]:
    """Looks for any user-defined parameters in the string, looks up those parameters and returns
    them in the order they appear in the string.
    """
    param_artifact_ids: List[uuid.UUID] = []

    matches = re.findall(USER_TAG_PATTERN, user_supplied_str)
    for match in matches:
        param_name = match.strip(" {}")
        param_ops = dag.get_param_ops_by_name(param_name)

        found_artifact: Optional[ArtifactMetadata] = None
        if len(param_ops) > 0:
            # Use the first explicitly-named parameter artifact we find.
            param_artifacts = dag.must_get_artifacts(
                [param_op.outputs[0] for param_op in param_ops]
            )
            for param_artifact in param_artifacts:
                if param_artifact.explicitly_named:
                    found_artifact = param_artifact
                    break

        if not found_artifact:
            raise InvalidUserArgumentException(
                "The parameter `%s` is not defined but is used in `%s`."
                % (param_name, user_supplied_str)
            )
        else:
            _validate_artifact_metadata_is_string(found_artifact)
            param_artifact_ids.append(found_artifact.id)

    return param_artifact_ids


def _validate_artifact_metadata_is_string(artifact: ArtifactMetadata) -> None:
    if artifact.type != ArtifactType.STRING:
        raise InvalidUserArgumentException(
            "The parameter `%s` must be defined as a string. Instead, got type %s"
            % (artifact.name, artifact.type)
        )


def _validate_artifact_is_string(artifact: BaseArtifact) -> None:
    if artifact.type() != ArtifactType.STRING:
        raise InvalidUserArgumentException(
            "The parameter `%s` must be defined as a string. Instead, got type %s"
            % (artifact.name(), artifact.type())
        )

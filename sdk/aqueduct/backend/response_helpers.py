import json
from typing import Any, Dict, List

import multipart
import requests
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import AqueductError, InternalAqueductError
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator
from aqueduct.utils.utils import indent_multiline_string, is_string_valid_uuid
from requests_toolbelt.multipart import decoder

from .response_models import ArtifactResult, Logs, OperatorResult, PreviewResponse


def _parse_artifact_result_response(response: requests.Response) -> Dict[str, Any]:
    multipart_data = decoder.MultipartDecoder.from_response(response)
    parse = multipart.parse_options_header

    result = {}

    for part in multipart_data.parts:
        field_name = part.headers[b"Content-Disposition"].decode(multipart_data.encoding)
        field_name = parse(field_name)[1]["name"]

        if field_name == "metadata":
            result[field_name] = json.loads(part.content.decode(multipart_data.encoding))
        elif field_name == "data":
            result[field_name] = part.content
        else:
            raise AqueductError(
                "Unexpected form field %s for artifact result response" % field_name
            )

    return result


def _construct_preview_response(response: requests.Response) -> PreviewResponse:
    artifact_results = {}
    artifact_result_constructor = {}
    preview_response = {}
    is_metadata_received = False
    multipart_data = decoder.MultipartDecoder.from_response(response)
    parse = multipart.parse_options_header

    for part in multipart_data.parts:
        field_name = part.headers[b"Content-Disposition"].decode(multipart_data.encoding)
        field_name = parse(field_name)[1]["name"]

        if field_name == "metadata":
            is_metadata_received = True
            metadata = json.loads(part.content.decode(multipart_data.encoding))
        elif is_string_valid_uuid(field_name):
            if is_metadata_received:
                artifact_result_constructor = metadata["artifact_types_metadata"][field_name]
                artifact_result_constructor["content"] = part.content
                artifact_results[field_name] = ArtifactResult(**artifact_result_constructor)
            else:
                raise AqueductError("Unable to retrieve artifacts metadata")
        else:
            raise AqueductError("Unable to get correct preview response")

    preview_response["status"] = metadata["status"]
    preview_response["operator_results"] = metadata["operator_results"]
    preview_response["artifact_results"] = artifact_results

    return PreviewResponse(**preview_response)


GITHUB_ISSUE_LINK = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"


def _handle_preview_resp(preview_resp: PreviewResponse, dag: DAG) -> None:
    """
    Prints all the logs generated during preview, in BFS order.

    Raises:
        AqueductError:
            If the preview execution has failed. This error will have the context
            and error message of every failed operator in it.
        InternalAqueductError:
            If something unexpected happened in our system.
    """
    # There can be multiple operator failures, one for each entry.
    op_err_msgs: List[str] = []

    def _construct_failure_error_msg(op_name: str, op_result: OperatorResult) -> str:
        """This is the message is raised in the Exception message."""
        assert op_result.error is not None
        return (
            f"Operator `{op_name}` failed!\n"
            f"{op_result.error.context}\n"
            f"\n"
            f"{op_result.error.tip}\n"
            f"\n"
        )

    def _print_op_user_logs(op_name: str, logs: Logs) -> None:
        """Prints out the logs for a single operator. The format is:

        stdout:
            {logs}
            {logs}
        ----------------------------------
        stderr:
            {logs}
            {logs}

        If either stdout or stderr is empty, we do not print anything for
        the empty section, and do not draw the "--" delimiter line.
        """
        if logs.is_empty():
            return

        print(f"Operator {op_name} Logs:")
        if len(logs.stdout) > 0:
            print("stdout:")
            print(indent_multiline_string(logs.stdout).rstrip("\n"))

        if len(logs.stdout) > 0 and len(logs.stderr) > 0:
            print("----------------------------------")

        if len(logs.stderr) > 0:
            print("stderr:")
            print(indent_multiline_string(logs.stderr).rstrip("\n"))
        print("")

    q: List[Operator] = dag.list_root_operators()
    seen_op_ids = set(op.id for op in q)
    while len(q) > 0:
        curr_op = q.pop(0)

        if curr_op.id in preview_resp.operator_results:
            curr_op_result = preview_resp.operator_results[curr_op.id]

            if curr_op_result.user_logs is not None:
                _print_op_user_logs(curr_op.name, curr_op_result.user_logs)

            if curr_op_result.error is not None:
                op_err_msgs.append(_construct_failure_error_msg(curr_op.name, curr_op_result))
            else:
                # Continue traversing, marking operators added to the queue as "seen"
                for output_artifact_id in curr_op.outputs:
                    next_operators = [
                        op
                        for op in dag.list_operators(on_artifact_id=output_artifact_id)
                        if op.id not in seen_op_ids
                    ]
                    q.extend(next_operators)
                    seen_op_ids.union(set(op.id for op in next_operators))

    if preview_resp.status == ExecutionStatus.PENDING:
        raise InternalAqueductError("Preview route should not be returning PENDING status.")

    if preview_resp.status == ExecutionStatus.FAILED:
        # If non of the operators failed, this must be an issue with our
        if len(op_err_msgs) == 0:
            raise InternalAqueductError(
                f"Unexpected Server Error! If this issue persists, please file a bug report in github: "
                f"{GITHUB_ISSUE_LINK} . We will get back to you as soon as we can.",
            )

        failure_err_msg = "\n".join(op_err_msgs)
        raise AqueductError(f"Preview Execution Failed:\n\n{failure_err_msg}\n")

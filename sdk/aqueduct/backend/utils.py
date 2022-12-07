import json
from typing import Any, Dict

import multipart
import requests
from aqueduct.error import AqueductError
from requests_toolbelt.multipart import decoder


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

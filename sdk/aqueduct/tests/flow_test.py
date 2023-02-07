import uuid
from typing import Tuple

import pytest
from aqueduct.error import InvalidUserArgumentException
from aqueduct.utils.utils import find_flow_with_user_supplied_id_and_name


def test_find_flow_with_user_supplied_id_and_name():
    flow_1_id = "9740eb10-d77d-4393-a9bc-42862d7008e0"
    flow_2_id = "6d9d7b93-028f-48b1-b723-ebd9e66ac867"
    flow_3_id = "431841d6-aac5-450f-b395-66e110ba547b"

    flows = [
        (uuid.UUID(flow_1_id), "flow_1"),
        (uuid.UUID(flow_2_id), "flow_2"),
        (uuid.UUID(flow_3_id), "flow_3"),
    ]

    flow_id = find_flow_with_user_supplied_id_and_name(flows, flow_id=flow_1_id)
    assert flow_id == flow_1_id

    flow_id = find_flow_with_user_supplied_id_and_name(flows, flow_name="flow_1")
    assert flow_id == flow_1_id

    flow_id = find_flow_with_user_supplied_id_and_name(flows, flow_id=flow_1_id, flow_name="flow_1")
    assert flow_id == flow_1_id

    with pytest.raises(InvalidUserArgumentException):
        flow_id = find_flow_with_user_supplied_id_and_name(
            flows, flow_id=flow_1_id, flow_name="flow_2"
        )

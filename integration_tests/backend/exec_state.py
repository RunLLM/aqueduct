from typing import Any, Dict

from dateutil.parser import parse as parse_datetime


# TODO: ENG-1685 we should use pydantic to handle responses
# assert_exec_state runs a number of asserts to ensure
# `exec_state` is consistent with the given `status`.
def assert_exec_state(exec_state: Dict[str, Any], status: str) -> None:
    timestamps = exec_state["timestamps"]
    pending_at = parse_datetime(timestamps["pending_at"]) if timestamps["pending_at"] else None
    running_at = parse_datetime(timestamps["running_at"]) if timestamps["running_at"] else None
    finished_at = parse_datetime(timestamps["finished_at"]) if timestamps["finished_at"] else None
    assert exec_state["status"] == status

    assert pending_at is not None
    if status == "succeeded" or status == "failed":
        assert running_at is not None
        assert finished_at is not None
        assert pending_at < running_at
        assert running_at < finished_at
        return

    if status == "running":
        assert running_at is not None
        assert finished_at is None
        assert pending_at < running_at
        return

    if status == "canceled":
        assert finished_at is not None
        assert pending_at < finished_at

        # note that a canceled state may or may not have `running_at`
        if running_at:
            assert pending_at < running_at
            assert running_at < finished_at
        return

    if status == "pending":
        assert running_at is None
        assert finished_at is None

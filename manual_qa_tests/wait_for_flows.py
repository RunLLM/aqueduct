from aqueduct.constants.enums import ExecutionStatus
import time

def _polling(
    stop_condition_fn,
    timeout=60,
    poll_threshold=5,
    timeout_comment="Timed out waiting for workflow run to complete.",
):
    begin = time.time()

    while True:
        assert time.time() - begin < timeout, timeout_comment

        if stop_condition_fn():
            break
        else:
            time.sleep(poll_threshold)


def _stop_condition(client):
    return all(map(lambda x: not x['last_run_status'] == str(ExecutionStatus.PENDING), client.list_flows()))

def wait_for_all_flows_to_complete(client):
    _polling(
        lambda: _stop_condition(client),
        timeout=600,
        poll_threshold=10,
        timeout_comment="Timed out waiting for workflow run to complete.",
    )

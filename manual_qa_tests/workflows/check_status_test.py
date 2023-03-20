import aqueduct as aq

NAME = "check_status_test"
DESCRIPTION = """
    * Workflows Page: "Check Status Test" should fail.
    * There should be four checks:
        * warning_level_pass which shows success icon.
        * warning_level_fail which shows warning icon.
        * error_level_fail which shows error icon.
        * error_level_pass which shows success icon.
    * Workflow Details Page:
        * Error message should appear below workflow header:
            A workflow-level error occurred!
            Error executing workflow
            Operator execution failed due to user error.

            <stack trace>
        * Two DAGs should appear - one for success and one for failure cases.
            * Success Test Dag:
                - test_pass operator (succeeded) -> test_pass_artifact (created) -> warning_level_pass (passed)
                                                                                 -> error_level_pass (passed)
            * Fail Test Dag:
                -test_fail operator (succeeded) -> test_fail_artifact (created) -> warning_level_fail (warning)
                                                                                -> error_level_pass (failed)
"""


@aq.op
def test_fail():
    return 1


@aq.op
def test_pass():
    return 0


@aq.check(severity="warning")
def warning_level_pass(res):
    return res == 0


@aq.check(severity="warning")
def warning_level_fail(res):
    return res == 0


@aq.check(severity="error")
def error_level_pass(res):
    return res == 0


@aq.check(severity="error")
def error_level_fail(res):
    return res == 0


def deploy(client, integration_name):
    fail_artf = test_fail()
    success_artf = test_pass()

    pass_level_warning_artf = warning_level_pass(success_artf)
    failure_level_warning_arf = warning_level_fail(fail_artf)
    pass_level_error_artf = error_level_pass(success_artf)

    # TODO: Get this to work without the whole workflow erroring out.
    # failure_level_error_artifact = error_level_fail.lazy(fail_artf)

    client.publish_flow(
        "Checks Status Test",
        artifacts=[
            pass_level_warning_artf,
            failure_level_warning_arf,
            pass_level_error_artf,
            # failure_level_error_artifact
        ],
    )

import aqueduct as aq

NAME = "succeed_dag_layout_test"
DESCRIPTION = """
    * This test is mainly here so we can look at the layout and check for edge crossing or
      awkward positioning.
    * Check that nodes are spaced out evenly and that there are no edge overlaps.
    * Each node should be in the "Succeeded" state
    * Make sure that ther isn't any extra space between checks/metrics and operator nodes.
    * op7 should have a value of 1
    * op6 should have a value of 3.
"""


@aq.op(requirements=[])
def op1():
    return 1


@aq.check(requirements=[])
def simple_check(param):
    return param == 1


@aq.op(requirements=[])
def op2(param):
    return param


@aq.op(requirements=[])
def op3(param1, param2):
    return param1 + param2


@aq.op(requirements=[])
def op4(param):
    return param


@aq.op(requirements=[])
def op5(param):
    return param


@aq.metric(requirements=[])
def op6(param1, param2):
    return param1 + param2


@aq.metric(requirements=[])
def op7(param):
    return param


def deploy(client, integration):
    res1 = op1()
    check1 = simple_check(res1)
    res2 = op2(res1)
    res3 = op3(res1, res2)
    res4 = op4(res2)
    res5 = op5(res2)
    res6 = op6(res3, res5)
    res7 = op7(res2)

    client.publish_flow(
        name=NAME,
        description=DESCRIPTION,
        artifacts=[res1, check1, res2, res3, res4, res5, res6, res7]
    )

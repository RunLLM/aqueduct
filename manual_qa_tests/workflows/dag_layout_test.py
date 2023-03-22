import aqueduct as aq

NAME = "dag_layout_test4"
DESCRIPTION = """
    * Workflows Page: TODO: Describe layouts here.
"""


@aq.op
def op1():
    return 1


@aq.op
def op2(param):
    return param


@aq.op
def op3(param1, param2):
    return param1


@aq.op
def op4(param):
    return param


@aq.op
def op5(param):
    return param


@aq.op
def op6(param1, param2):
    return param1 + param2


@aq.op
def op7(param):
    return param


def deploy(client, integration):
    res1 = op1()
    res2 = op2(res1)
    res3 = op3(res1, res2)
    res4 = op4(res2)
    res5 = op5(res2)
    res6 = op6(res3, res5)
    res7 = op7(res2)

    client.publish_flow(
        name=NAME,
        description=DESCRIPTION,
        artifacts=[res1, res2, res3, res4, res5, res6, res7]
    )

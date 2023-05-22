# from aqueduct import Client

# client = Client()

# demo_db = client.integration("snowflake_integration")
# reviews_table = demo_db.sql("select * from hotel_reviews;")

# # demo_db.save(reviews_table, table_name="reviews_table_2", update_mode="replace")

# source_flow = client.publish_flow(
#     name="hotel_reviews", artifacts=[reviews_table]
# ).id()  # Or set your workflow ID here.

# # Wait for workflow to be created.
# import time

# time.sleep(10)

# import random

# # source_flow can be a Flow object, workflow name, or workflow ID
# import string

# letters = string.ascii_lowercase

# workflow_b = "".join([random.choice(letters) for i in range(10)])

# flow = client.publish_flow(name=workflow_b, artifacts=[reviews_table], source_flow=source_flow)

# workflow_b = "".join([random.choice(letters) for i in range(10)])

# flow = client.publish_flow(name=workflow_b, artifacts=[reviews_table], source_flow=source_flow)

# from aqueduct import Client

# client = Client()

# demo_db = client.integration("aqueduct_demo")
# reviews_table = demo_db.sql("select * from reviews_table_2;")

# # source_flow can be a Flow object, workflow name, or workflow ID
# flow = client.publish_flow(name='workflow_b',
#                            artifacts=[reviews_table],
#                            source_flow="24f65a7c-16f6-4cf4-a69c-cfca2be8d922")


# import pandas as pd

# from aqueduct import Client

# client = Client()
# db = client.integration("aqueduct_demo")
# my_tbl = db.sql("Select * from customers")
# number_col = "n_workflows"

# mean_val = my_tbl.mean(number_col)

# # The workflow will fail if the mean of `number_col` is 0.
# mean_val.bound(notequal=0, severity='error')

# # The workflow will raise an error but finish executing if the mean of
# # `number_col` is > 100.
# mean_val.bound(upper=100, severity='warning')

# import aqueduct as aq

# NAME = "2_dags"
# DESCRIPTION = """
#     * Workflows Page: "Check Status Test" should succeed.
#     * There should be four checks:
#         * warning_level_pass which shows success icon.
#         * warning_level_fail which shows warning icon.
#         * error_level_pass which shows success icon.
#     * Workflow Details Page:
#         * Two DAGs should appear - one for success and one for failure cases.
#             * Success Test Dag:
#                 - test_pass operator (succeeded) -> test_pass_artifact (created) -> warning_level_pass (passed)
#                                                                                  -> error_level_pass (passed)
#             * Fail Test Dag:
#                 -test_fail operator (succeeded) -> test_fail_artifact (created) -> warning_level_fail (warning)
# """


# @aq.op(requirements=[])
# def test_fail():
#     return 1


# @aq.op(requirements=[])
# def test_pass():
#     return 0


# @aq.check(severity="warning", requirements=[])
# def warning_level_pass(res):
#     return res == 0


# @aq.check(severity="warning", requirements=[])
# def warning_level_fail(res):
#     return res == 0


# @aq.check(severity="error", requirements=[])
# def error_level_pass(res):
#     return res == 0


# @aq.check(severity="error", requirements=[])
# def error_level_fail(res):
#     return res == 0


# def deploy(client, integration_name):
#     # fail_artf = test_fail()
#     success_artf = test_pass()

#     # pass_level_warning_artf = warning_level_pass(success_artf)
#     # failure_level_warning_arf = warning_level_fail(fail_artf)
#     # pass_level_error_artf = error_level_pass(success_artf)

#     client.publish_flow(
#         name=NAME,
#         description=DESCRIPTION,
#         artifacts=[
#             success_artf#, fail_artf
#         ],
#     )

# deploy(aq.Client(), 'aqueduct_demo')

# print(aq.globals.__GLOBAL_CONFIG__.engine)
# # import sys
# # import time

# # import pandas as pd

# # import aqueduct as aq

from aqueduct import Client, check, globals, metric, op

client = Client()

# demo_db = client.integration("aqueduct_demo")


# # @aq.op(requirements=[])
# # def test_fail():
# #     return 1


# # @aq.op(requirements=[])
# # def test_pass():
# #     return 0


# # @aq.check(severity="warning", requirements=[])
# # def warning_level_pass(res):
# #     return res == 0


# # @aq.check(severity="warning", requirements=[])
# # def warning_level_fail(res):
# #     return res == 0


# # @aq.check(severity="error", requirements=[])
# # def error_level_pass(res):
# #     return res == 0


# # @aq.check(severity="error", requirements=[])
# # def error_level_fail(res):
# #     return res == 0


# # def deploy(client, integration_name):
# #     fail_artf = test_fail()
# #     success_artf = test_pass()

# #     pass_level_warning_artf = warning_level_pass(success_artf)
# #     failure_level_warning_arf = warning_level_fail(fail_artf)
# #     pass_level_error_artf = error_level_pass(success_artf)

# #     client.publish_flow(
# #         name="test_check",
# #         description="...",
# #         artifacts=[
# #             pass_level_warning_artf,
# #             failure_level_warning_arf,
# #             pass_level_error_artf,
# #         ],
# #     )
# # deploy(client, "aqueduct_demo")

# df1 = demo_db.sql("select * from hotel_reviews;")
# num_param = client.create_param("num", default=10)


@op()
def op1(df):
    return df


@op()
def op2(df):
    return df


@check()
def check1(df):
    return True


@metric()
def metric1(df):
    return 10


import pandas as pd

df1 = pd.DataFrame([])
df2 = op1.lazy(df1)
df3 = op2.lazy(df2)
bool1 = check1.lazy(df2)
num1 = metric1.lazy(df2)

# demo_db.save(df1, table_name='df1', update_mode='replace')
# client.integration("s3_storage").save(df1, filepath='dfs/this_is_a_very_long_path_for_testing_purposes/df1.csv', format="CSV")


flow = client.publish_flow(
    "test",
    artifacts=[df2],
)

# print(flow)

# # # ###################################################

# # # # import aqueduct as aq

# # # # from aqueduct import Client, check, metric, op

# # # # client = Client()

# # # # @aq.metric(requirements=[])
# # # # def row_count(_):
# # # #     x = [1]
# # # #     return x[2] # bug


# # # # @aq.op(requirements=[])
# # # # def good_op(_):
# # # #     x = [1]
# # # #     return x[0]


# # # # def deploy(client, integration_name):
# # # #     integration = client.integration(integration_name)
# # # #     reviews = integration.sql("SELECT * FROM hotel_reviews")
# # # #     op_artf = good_op.lazy(reviews)
# # # #     row_count_artf = row_count.lazy(op_artf)
# # # #     client.publish_flow(
# # # #         artifacts=[row_count_artf],
# # # #         name="Test ENG-2578",
# # # #         description="ENG-2578",
# # # #         schedule="",
# # # #     )

# # # # deploy(client, "aqueduct_demo")

# # # #####################################

import sys
import time

import pandas as pd

from aqueduct import Client, check, metric, op

client = Client()

demo_db = client.integration("Demo")
df1 = demo_db.sql("select * from hotel_reviews;")
num_param = client.create_param("num", default=10)


@op()
def op1(df, num):
    return df


@op()
def op2(df):
    return df


@check()
def check1(df):
    return True


@metric()
def metric1(df):
    return 10


df2 = op1.lazy(df1, num_param)
df3 = op2.lazy(df2)
bool1 = check1.lazy(df2)
# num1 = metric1.lazy(df2)

flow = client.publish_flow(
    "no_run",
    run_now=False,
    artifacts=[df1, df2, bool1],
)
print(flow)

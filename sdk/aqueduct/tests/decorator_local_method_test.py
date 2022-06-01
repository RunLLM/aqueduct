from aqueduct.decorator import op,metric,check
from aqueduct.tests.utils import (
    default_table_artifact,
)
inp = default_table_artifact()
@op()
def op_fn_with_parentheses(df):
    print(1)
print(op_fn_with_parentheses.local(inp))
@check()
def check_fn_with_parentheses(df):
    print(2)
print(check_fn_with_parentheses.local(inp))
@metric()
def metric_fn_with_parentheses(df):
    print(3)
print(metric_fn_with_parentheses.local(inp))

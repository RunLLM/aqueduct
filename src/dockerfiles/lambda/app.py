import sys

import aqueduct_executor


def handler(event, context):
    return "Hello from AWS Lambda using Python" + sys.version + "!"

import argparse
import os
import subprocess
import sys
from contextlib import redirect_stdout
from io import StringIO

from aqueduct import Client, get_apikey

"""
This script is used in regression testing to compare workflows.

The checkpoint flag specifies the action to take
create - stores workflow info as a checkpoint
diff - compares the current state of the workflow with the last stored checkpoint

Current comparisons are comparing the description of the flow each time for equality
"""

parser = argparse.ArgumentParser()
parser.add_argument("--path", required=True, help="The relative path for checkpoints.")
parser.add_argument(
    "--flow_id",
    required=False,
    default=None,
    help="The flow id of the workflow we intend to store/diff. If not provided will take the first one fetched from the client",
)
parser.add_argument(
    "--checkpoint",
    required=False,
    default=None,
    help="Use create to create checkpoint for the flow. Use compare to compare with a previous checkpoint",
)
parser.add_argument(
    "--diff",
    required=False,
    default=None,
    help="Indicates that we wants to compare the flow with a previously stored snapshot",
)
parser.add_argument(
    "--server_address",
    required=False,
    default=None,
    help="The server address where flow is located",
)
args = parser.parse_args()

client = Client(get_apikey(), args.server_address.strip('"'))

if args.flow_id is None:
    flows = client.list_flows()
    if not flows:
        raise Exception("Could not find any flows on the server")
    flow_id = client.list_flows()[0]["flow_id"]
else:
    flow_id = str(args.flow_id)

flow = client.flow(flow_id)
checkpoint_path = os.path.join(args.path, flow_id)
if not os.path.isdir(args.path):
    os.mkdir(args.path)

if args.checkpoint == "create":
    with open(checkpoint_path, "w+") as sys.stdout:
        flow.describe()

if args.checkpoint == "diff":
    with open(checkpoint_path, "r") as file:
        previous_flow_info = file.read()

    stdout_log = StringIO()
    with redirect_stdout(stdout_log):
        flow.describe()
    stdout_log.seek(0)
    flow_info = stdout_log.read()

    if previous_flow_info == flow_info:
        print("Checkpoint passed. Flow diff for %s from checkpoint is same" % flow_id)
    else:
        raise Exception(
            "Flow diff for %s from checkpoint is not the same. \nCheckpoint info:\n%s\nCurrent info:\n%s"
            % (flow_id, previous_flow_info, flow_info)
        )

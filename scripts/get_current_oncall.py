"""
This script prints out the Slack Member ID of the current oncall, to be consumed by our GH actions so that we tag
the current oncall on any Slack notifications (to #eng-monitoring, for example).

```
Rotation:
    <oncall name>: <start date> # the date is in the format MM/DD
    <oncall name>: <start date>
    ...

Slack Member ID:
    <oncall name>: <slack id>
    <oncall name>: <slack id>
    ...
```

Prints out Slack Member ID of the current oncall.
"""
import argparse
from datetime import datetime, timedelta
from typing import Optional

import yaml

SLACK_MEMBER_ID_KEY = "Slack Member ID"
ROTATION_KEY = "Rotation"


def _most_recently_passed_monday():
    """Calculate the date of the most recently passed Monday. Return it in MM/DD string format."""
    today = datetime.today()
    last_monday = today - timedelta(days=today.weekday())  # Since Monday's weekday() value is 0.
    return last_monday.strftime("%m/%d")


def print_current_oncall_slack_member_id(config_filepath: str):
    with open(config_filepath, "r") as f:
        config_dict = yaml.safe_load(f)

        assert SLACK_MEMBER_ID_KEY in config_dict.keys()
        assert ROTATION_KEY in config_dict.keys()

        curr_oncall_start_date_str = _most_recently_passed_monday()
        if curr_oncall_start_date_str not in config_dict[ROTATION_KEY].values():
            raise Exception(
                "Expected an entry in %s for the week of %s, but found none."
                % (config_filepath, curr_oncall_start_date_str)
            )

        oncall_name: Optional[str] = None
        for name, start_date_str in config_dict[ROTATION_KEY].items():
            if start_date_str == curr_oncall_start_date_str:
                oncall_name = name
        assert oncall_name is not None

        # Look up the oncall's slack member id.
        assert (
            oncall_name in config_dict[SLACK_MEMBER_ID_KEY].keys()
        ), "%s does not have a slack entry in %s" % (oncall_name, config_filepath)

        print(config_dict[SLACK_MEMBER_ID_KEY][oncall_name])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--filepath",
        dest="filepath",
        required=True,
        action="store",
        help="The yml file containing the appropriate format described at the top of this script.",
    )
    args = parser.parse_args()

    print_current_oncall_slack_member_id(args.filepath)

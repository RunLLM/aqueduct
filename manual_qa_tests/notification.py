import aqueduct as aq
from aqueduct.constants.enums import NotificationLevel
from aqueduct.integrations.connect_config import ServiceType, SlackConfig


def connect_slack(
    client: aq.Client,
    token: str,
    channel: str,
    level: NotificationLevel,
) -> None:
    client.connect_integration(
        "test_slack_notification",
        "Slack",
        SlackConfig(token=token, channels=[channel], level=level, enabled=True),
    )

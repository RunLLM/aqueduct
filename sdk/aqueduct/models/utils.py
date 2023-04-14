from datetime import datetime

TIME_FORMAT = "%Y-%m-%d %H:%M:%S"


def human_readable_timestamp(ts: int) -> str:
    return datetime.utcfromtimestamp(ts).strftime(TIME_FORMAT)

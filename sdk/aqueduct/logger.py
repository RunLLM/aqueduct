import logging


def logger() -> logging.Logger:
    """
    This is the logger shared within the aqueduct module.
    ref: https://docs.python.org/3/howto/logging.html

    The log level is configured by the Aqueduct client. If not configured, it will default
    to level WARNING.
    """
    return logging.getLogger(__name__)

from enum import Enum
from typing import Union


def is_valid_minute(minute: int) -> bool:
    return minute >= 0 and minute < 60


def is_valid_hour(hour: int) -> bool:
    return hour >= 0 and hour < 24


def is_valid_day_of_week(day: int) -> bool:
    return day > 0 and day < 8


def is_valid_day_of_month(day: int) -> bool:
    return day > 0 and day < 32


class Minute:
    def __init__(self, minute: int):
        if not is_valid_minute(minute):
            raise Exception("Invalid minute value %s." % minute)
        self.val = minute


class Hour:
    def __init__(self, hour: int):
        if not is_valid_hour(hour):
            raise Exception("Invalid hour value %s." % hour)
        self.val = hour


class DayOfWeek(Enum):
    MONDAY = 1
    TUESDAY = 2
    WEDNESDAY = 3
    THURSDAY = 4
    FRIDAY = 5
    SATURDAY = 6
    SUNDAY = 7


class DayOfMonth:
    def __init__(self, day: int):
        if not is_valid_day_of_month(day):
            raise Exception("Invalid day value %s." % day)
        self.val = day


def hourly(minute: Union[int, Minute] = 0) -> str:
    if not (isinstance(minute, Minute)) and not (isinstance(minute, int)):
        raise Exception("Must be either an integer or an aqueduct.schedule.Minute object.")

    if isinstance(minute, int):
        minute = Minute(minute)

    assert isinstance(minute, Minute)
    return "%s * * * *" % (minute.val)


def daily(hour: Union[int, Hour] = 0, minute: Union[Minute, int] = 0) -> str:
    if not isinstance(hour, Hour) and not isinstance(hour, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Hour object.")
    if not isinstance(minute, Minute) and not isinstance(minute, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Minute object.")

    if isinstance(hour, int):
        hour = Hour(hour)
    if isinstance(minute, int):
        minute = Minute(minute)

    assert isinstance(hour, Hour) and isinstance(minute, Minute)
    return "%s %s * * *" % (minute.val, hour.val)


def weekly(day: DayOfWeek, hour: Hour = Hour(0), minute: Minute = Minute(0)) -> str:
    if not isinstance(day, DayOfWeek) and not isinstance(day, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.DayOfWeek object.")
    if not isinstance(hour, Hour) and not isinstance(hour, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Hour object.")
    if not isinstance(minute, Minute) and not isinstance(minute, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Minute object.")

    if isinstance(day, int):
        day = DayOfWeek(day)
    if isinstance(hour, int):
        hour = Hour(hour)
    if isinstance(minute, int):
        minute = Minute(minute)

    assert isinstance(day, DayOfWeek) and isinstance(hour, Hour) and isinstance(minute, Minute)
    return "%s %s * * %s" % (minute.val, hour.val, day.value)


def monthly(day: DayOfMonth, hour: Hour = Hour(0), minute: Minute = Minute(0)) -> str:
    if not isinstance(day, DayOfMonth) and not isinstance(day, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.DayOfMonth object.")
    if not isinstance(hour, Hour) and not isinstance(hour, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Hour object.")
    if not isinstance(minute, Minute) and not isinstance(minute, int):
        raise Exception("Must be either an integer or an aqueduct.schedule.Minute object.")

    if isinstance(day, int):
        day = DayOfMonth(day)
    if isinstance(hour, int):
        hour = Hour(hour)
    if isinstance(minute, int):
        minute = Minute(minute)

    assert isinstance(day, DayOfMonth) and isinstance(hour, Hour) and isinstance(minute, Minute)
    return "%s %s %s * *" % (minute.val, hour.val, day.val)

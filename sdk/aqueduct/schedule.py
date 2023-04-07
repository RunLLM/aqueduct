from enum import Enum


class Minute:
    def __init__(self, minute: int):
        if minute < 0 or minute >= 60:
            raise Exception("Invalid minute value %s." % minute)
        self.val = minute


class Hour:
    def __init__(self, hour: int):
        if hour < 0 or hour >= 24:
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
        if day < 0 or day >= 32:
            raise Exception("Invalid hour value %s." % day)
        self.val = day


def hourly(minute: Minute = Minute(0)) -> str:
    if not (isinstance(minute, Minute)):
        raise Exception("Invalid types provided in parameters: aqueduct.schedule.Minute required.")

    return "%s * * * *" % (minute.val)


def daily(hour: Hour = Hour(0), minute: Minute = Minute(0)) -> str:
    if not isinstance(hour, Hour):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Hour required for first argument."
        )
    if not isinstance(minute, Minute):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Minute required for second argument."
        )

    return "%s %s * * *" % (minute.val, hour.val)


def weekly(day: DayOfWeek, hour: Hour = Hour(0), minute: Minute = Minute(0)) -> str:
    if not isinstance(day, DayOfWeek):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.DayOfWeek required for first argument."
        )
    if not isinstance(hour, Hour):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Hour required for second argument."
        )
    if not isinstance(minute, Minute):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Minute required for third argument."
        )

    return "%s %s * * %s" % (minute.val, hour.val, day)


def monthly(day: DayOfMonth, hour: Hour = Hour(0), minute: Minute = Minute(0)) -> str:
    if not isinstance(day, DayOfWeek):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.DayOfMonth required for first argument."
        )
    if not isinstance(hour, Hour):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Hour required for second argument."
        )
    if not isinstance(minute, Minute):
        raise Exception(
            "Invalid types provided in parameters: aqueduct.schedule.Minute required for third argument."
        )

    return "%s %s %s * *" % (minute.val, hour.val, day.val)

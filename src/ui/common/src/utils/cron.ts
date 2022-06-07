import cronParser, { CronDate } from 'cron-parser';

export enum DayOfWeek {
  Sunday,
  Monday,
  Tuesday,
  Wednesday,
  Thursday,
  Friday,
  Saturday,
}

export enum PeriodUnit {
  Hourly = 'Hourly',
  Daily = 'Daily',
  Weekly = 'Weekly',
  Monthly = 'Monthly',
}

type Schedule = {
  periodUnit: PeriodUnit;
  minute?: number;
  time?: string;
  dayOfMonth?: number;
  dayOfWeek: DayOfWeek;
};

function createCronString(schedule: Schedule): string {
  if (schedule.periodUnit === PeriodUnit.Hourly) {
    return `${schedule.minute} * * * *`;
  }

  // `time` will always be defined if periodUnit is daily/weekly/monthly.
  const timeList = schedule.time.split(':');
  const timeHour = parseInt(timeList[0]);
  const timeMinute = parseInt(timeList[1]);

  switch (schedule.periodUnit) {
    case PeriodUnit.Monthly:
      return `${timeMinute} ${timeHour} ${schedule.dayOfMonth} * *`;
    case PeriodUnit.Weekly:
      return `${timeMinute} ${timeHour} * * ${schedule.dayOfWeek}`;
    case PeriodUnit.Daily:
      return `${timeMinute} ${timeHour} * * *`;
    default:
      return '';
  }
}

function deconstructCronString(cronString: string): Schedule {
  const parsed = cronParser.parseExpression(cronString);
  const hourString =
    parsed.fields.hour[0] < 10
      ? `0${parsed.fields.hour[0]}`
      : `${parsed.fields.hour[0]}`;
  const minuteString =
    parsed.fields.minute[0] < 10
      ? `0${parsed.fields.minute[0]}`
      : `${parsed.fields.minute[0]}`;

  if (parsed.fields.dayOfMonth.length === 1) {
    // We picked a particular day of the month.
    return {
      // These are the values set by the user.
      periodUnit: PeriodUnit.Monthly,
      time: `${hourString}:${minuteString}`,
      dayOfMonth: Number(parsed.fields.dayOfMonth[0]),
      // These are placeholder values.
      minute: 0,
      dayOfWeek: DayOfWeek.Sunday,
    };
  } else if (parsed.fields.dayOfWeek.length === 1) {
    // We picked a particular day of the week.
    return {
      // These are the values set by the user.
      periodUnit: PeriodUnit.Weekly,
      time: `${hourString}:${minuteString}`,
      dayOfWeek: parsed.fields.dayOfWeek[0],
      // These are placeholder values.
      minute: 0,
      dayOfMonth: 1,
    };
  } else if (parsed.fields.hour.length === 1) {
    // We picked a particular hour of the day.
    return {
      // These are the values set by the user.
      periodUnit: PeriodUnit.Daily,
      time: `${hourString}:${minuteString}`,
      // These are placeholder values.
      minute: 0,
      dayOfMonth: 1,
      dayOfWeek: DayOfWeek.Sunday,
    };
  } else {
    // We picked an hourly schedule.
    return {
      // These are the values set by the user.
      periodUnit: PeriodUnit.Hourly,
      minute: parsed.fields.minute[0],
      // These are placeholder values.
      time: '00:00',
      dayOfMonth: 1,
      dayOfWeek: DayOfWeek.Sunday,
    };
  }
}

function getNextUpdateTime(
  cronString: string | undefined
): CronDate | undefined {
  if (!cronString) {
    return undefined;
  }
  const parsedCron = cronParser.parseExpression(cronString);
  return parsedCron.next();
}

export { createCronString, deconstructCronString, getNextUpdateTime };

import { Box, FormControl, MenuItem, Select, TextField } from '@mui/material';
import React, { useEffect, useState } from 'react';

import {
  createCronString,
  DayOfWeek,
  deconstructCronString,
  PeriodUnit,
} from '../../utils/cron';

type Props = {
  cronString: string;
  setSchedule: (string) => void;
};

const PeriodicScheduleSelector: React.FC<Props> = ({
  cronString,
  setSchedule,
}) => {
  const schedule = deconstructCronString(cronString);

  const [timeUnit, setTimeUnit] = useState(schedule.periodUnit);
  const [minute, setMinute] = useState(schedule.minute);
  const [time, setTime] = useState(schedule.time);
  const [dayOfWeek, setDayOfWeek] = useState(schedule.dayOfWeek);
  const [dayOfMonth, setDayOfMonth] = useState(schedule.dayOfMonth);

  useEffect(() => {
    // Don't try to update the cron schedule if the user enters an invalid
    // input.
    if (
      (timeUnit === PeriodUnit.Hourly && (minute < 0 || minute > 59)) ||
      (timeUnit === PeriodUnit.Monthly && (dayOfMonth < 1 || dayOfMonth > 31))
    ) {
      return;
    }

    setSchedule(
      createCronString({
        periodUnit: timeUnit,
        minute,
        time,
        dayOfWeek,
        dayOfMonth,
      })
    );
  }, [timeUnit, minute, time, dayOfWeek, dayOfMonth, setSchedule]);

  return (
    <Box sx={{ display: 'flex' }}>
      <FormControl size="small" sx={{ mr: 1 }}>
        <Select
          value={timeUnit}
          onChange={(e) => setTimeUnit(e.target.value as PeriodUnit)}
        >
          {Object.values(PeriodUnit).map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {timeUnit === 'Monthly' && (
        <TextField
          size="small"
          label="Date"
          sx={{ width: '100px' }}
          type="number"
          value={dayOfMonth}
          onChange={(e) => setDayOfMonth(Number(e.target.value))}
          error={dayOfMonth < 1 || dayOfMonth > 31}
        />
      )}

      {timeUnit === 'Weekly' && (
        <FormControl size="small" sx={{ mx: 1 }}>
          <Select
            value={dayOfWeek}
            onChange={(e) => setDayOfWeek(e.target.value as DayOfWeek)}
          >
            {
              // This is an ugly bit of code. Typescript creates
              // reverse mappings (key->value, value=>key) for
              // numerical enums, so we have to filter out the
              // value->key mappings here before generating the
              // options.
              Object.keys(DayOfWeek)
                .filter((key) => isNaN(Number(key)))
                .map((day) => (
                  <MenuItem key={day} value={DayOfWeek[day]}>
                    {day}
                  </MenuItem>
                ))
            }
          </Select>
        </FormControl>
      )}

      {timeUnit !== 'Hourly' && (
        <TextField
          label="Time"
          sx={{ width: '150px', mx: 1 }}
          size="small"
          type="time"
          value={time}
          onChange={(e) => setTime(e.target.value)}
        />
      )}

      {timeUnit === 'Hourly' && (
        <TextField
          label="Minute"
          sx={{ width: '100px', mx: 1 }}
          size="small"
          type="number"
          value={minute}
          onChange={(e) => setMinute(Number(e.target.value))}
        />
      )}
    </Box>
  );
};

export default PeriodicScheduleSelector;

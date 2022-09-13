import { useField } from 'formik';
import DateTimePicker from 'react-datetime-picker';
import { Calendar, X } from 'react-feather';
import { useMemo } from 'react';

import { isoDate, parseIsoDate } from '@/portainer/filters/filters';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';

import { FormValues } from './types';

interface Props {
  disabled?: boolean;
}

export function ScheduledTimeField({ disabled }: Props) {
  const [{ name, value }, { error }, { setValue }] =
    useField<FormValues['time']>('time');

  const dateValue = useMemo(() => parseIsoDate(value), [value]);

  return (
    <FormControl label="Schedule date & time" errors={error}>
      {!disabled ? (
        <DateTimePicker
          format="y-MM-dd HH:mm:ss"
          minDate={new Date()}
          className="form-control [&>div]:border-0"
          onChange={(date) => setValue(isoDate(date.valueOf()))}
          name={name}
          value={dateValue}
          calendarIcon={<Calendar className="feather" />}
          clearIcon={<X className="feather" />}
          disableClock
        />
      ) : (
        <Input defaultValue={value} disabled />
      )}
    </FormControl>
  );
}

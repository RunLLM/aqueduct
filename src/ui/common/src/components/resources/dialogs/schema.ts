import * as Yup from 'yup';
import { ObjectShape } from 'yup/lib/object';

export function requiredAtCreate(
  schema: Yup.ObjectSchema<ObjectShape>,
  editMode: boolean,
  msg?: string | undefined
): Yup.ObjectSchema<ObjectShape> {
  return editMode ? schema : schema.required(msg);
}

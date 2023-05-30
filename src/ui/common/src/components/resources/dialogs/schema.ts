export function requiredAtCreate(
  schema,
  editMode: boolean,
  msg?: string | undefined
): any {
  return editMode ? schema : schema.required(msg);
}

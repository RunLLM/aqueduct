// Takes the number of workflows using the integration and returns a consistent string message.
export function getNumWorkflowsUsingMessage(numWorkflowsUsing: number): string {
  if (numWorkflowsUsing > 0) {
    return `Used by ${numWorkflowsUsing} ${
      numWorkflowsUsing === 1 ? 'workflow' : 'workflows'
    }`;
  } else {
    return 'Not currently in use';
  }
}

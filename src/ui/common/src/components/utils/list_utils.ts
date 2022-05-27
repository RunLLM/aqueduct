// Helper function for removing duplicates from WorkflowStatusItem arrays.
// see: https://stackoverflow.com/questions/2218999/how-to-remove-all-duplicates-from-an-array-of-objects
export default function getUniqueListBy<T>(arr: T[], key: string): T[] {
    const mapValues = new Map(arr.map((item) => [item[key], item])).values();
    return Array.from(mapValues);
}

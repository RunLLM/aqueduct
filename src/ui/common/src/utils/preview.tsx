import {Data} from "./data";

/**
 * Exports the tableState to CSV.
 * Returns true if successful, and false if there is an error exporting the CSV.
 * @param tableState the tablState to export to CSV
 * @param fileName the name of the exported .csv file. e.g. fileName.csv
 */
export function exportCsv(tableState: Data, fileName: string): boolean {
    try {
        let content = tableState.schema.fields.map((column) => column.name).join(',') + '\n';
        const mappedData = tableState.data.map((row) => {
            const rowValues = Object.keys(row).map((rowKey) => {
                return row[rowKey];
            });
            return rowValues.join(',');
        });
        content += mappedData.join('\n');
        const csvBlob = new Blob([content], { type: 'text/csv' });
        const url = window.URL.createObjectURL(csvBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = fileName + '.csv';
        a.click();

        return true;
    } catch (err) {
        return false;
    }
}

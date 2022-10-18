import { Box, Typography } from '@mui/material';
import { styled, Theme } from '@mui/material/styles';
import TableCell from '@mui/material/TableCell';
import clsx from 'clsx';
import React from 'react';
import {
  AutoSizer,
  Column,
  Table,
  TableCellRenderer,
  TableHeaderProps,
} from 'react-virtualized';

const classes = {
  flexContainer: 'ReactVirtualizedDemo-flexContainer',
  tableRow: 'ReactVirtualizedDemo-tableRow',
  tableRowHover: 'ReactVirtualizedDemo-tableRowHover',
  tableCell: 'ReactVirtualizedDemo-tableCell',
  noClick: 'ReactVirtualizedDemo-noClick',
  headerRow: 'ReactVirtualized__Table__headerRow',
  headerColumn: 'ReactVirtualized__Table__headerColumn',
  rowColumn: 'ReactVirtualized__Table__rowColumn',
  headerColumnFirstOfType:
    'ReactVirtualized__Table__headerColumn:first-of-type',
  rowColumnFirstOfType: 'ReactVirtualized__Table__rowColumn:first-of-type',
};

const columnWidthMultiplier = 13;

const styles = ({ theme }: { theme: Theme }) =>
  ({
    // temporary right-to-left patch, waiting for
    // https://github.com/bvaughn/react-virtualized/issues/454
    [`& .${classes.headerRow}`]: {
      ...(theme.direction === 'rtl' && {
        paddingLeft: '0 !important',
      }),
      ...(theme.direction !== 'rtl' && {
        paddingRight: undefined,
      }),
    },
    [`& .${classes.headerColumn}`]: {
      ...{
        marginRight: '0px',
        width: '100%',
      },
    },
    [`& .${classes.rowColumn}`]: {
      ...{
        marginRight: '0px',
      },
    },
    [`& .${classes.headerColumnFirstOfType}`]: {
      ...{
        marginLeft: '0px',
      },
    },
    [`& .${classes.rowColumnFirstOfType}`]: {
      ...{
        marginLeft: '0px',
      },
    },
    [`& .${classes.flexContainer}`]: {
      display: 'flex',
      alignItems: 'center',
      boxSizing: 'border-box',
      paddingRight: '0 !important',
    },
    [`& .${classes.tableRow}`]: {
      cursor: 'pointer',
    },
    [`& .${classes.tableRowHover}`]: {
      '&:hover': {
        backgroundColor: theme.palette.grey[200],
      },
    },
    [`& .${classes.tableCell}`]: {
      flex: 1,
    },
    [`& .${classes.noClick}`]: {
      cursor: 'initial',
    },
  } as const);

interface ColumnData {
  dataKey: string;
  label: string;
  type: string;
  numeric?: boolean;
  columnWidth?: number;
}

interface Row {
  index: number;
}

interface DataTableProps {
  columns: readonly ColumnData[];
  headerHeight?: number;
  minColumnWidth?: number;
  onRowClick?: () => void;
  rowCount: number;
  rowGetter: (row: Row) => any;
  rowHeight?: number;
}

const DEFAULT_ROW_HEIGHT = 48;
const DEFAULT_HEADER_HEIGHT = 72;

class PreStyledDataTable extends React.PureComponent<DataTableProps> {
  static defaultProps = {
    headerHeight: DEFAULT_HEADER_HEIGHT,
    rowHeight: DEFAULT_ROW_HEIGHT,
    minColumnWidth: 150,
  };

  getRowClassName = ({ index }: Row) => {
    const { onRowClick } = this.props;

    return clsx(classes.tableRow, classes.flexContainer, {
      [classes.tableRowHover]: index !== -1 && onRowClick != null,
    });
  };

  cellRenderer: TableCellRenderer = ({ columnData, cellData, columnIndex }) => {
    const { columns, rowHeight, onRowClick } = this.props;
    return (
      <TableCell
        component="div"
        className={clsx(classes.tableCell, classes.flexContainer, {
          [classes.noClick]: onRowClick == null,
        })}
        variant="body"
        style={{
          height: rowHeight,
        }}
        align={
          (columnIndex != null && columns[columnIndex].numeric) || false
            ? 'right'
            : 'left'
        }
      >
        <Typography
          variant="body1"
          noWrap
          sx={{
            textOverflow: 'ellipsis',
            overflow: 'hidden',
            width: columnData.columnWidth * 0.8,
          }}
        >
          {cellData}
        </Typography>
      </TableCell>
    );
  };

  headerRenderer = ({
    columnData,
    columnIndex,
  }: TableHeaderProps & { columnIndex: number }) => {
    const { headerHeight, columns } = this.props;

    return (
      <TableCell
        sx={{
          backgroundColor: 'blue.900',
          color: 'white',
        }}
        component="div"
        className={clsx(
          classes.tableCell,
          classes.flexContainer,
          classes.noClick
        )}
        variant="head"
        style={{ height: headerHeight }}
        align={columns[columnIndex].numeric || false ? 'right' : 'left'}
      >
        <Box style={{ display: 'flex', flexDirection: 'column' }}>
          <Typography
            variant="body1"
            sx={{
              textTransform: 'none',
              fontFamily: 'monospace',
              fontSize: '16px',
            }}
          >
            {columnData.label}
          </Typography>
          <Typography
            variant="caption"
            sx={{
              textTransform: 'none',
              fontFamily: 'monospace',
              fontSize: '12px',
            }}
          >
            {columnData.type}
          </Typography>
        </Box>
      </TableCell>
    );
  };

  render() {
    const {
      columns,
      rowHeight = DEFAULT_ROW_HEIGHT,
      headerHeight = DEFAULT_HEADER_HEIGHT,
      minColumnWidth,
      ...tableProps
    } = this.props;

    let MIN_TABLE_WIDTH = 0;
    columns.forEach((column) => {
      if (!column.columnWidth) {
        column.columnWidth = Math.max(
          column.label.length * columnWidthMultiplier,
          minColumnWidth
        );
      }
      MIN_TABLE_WIDTH += column.columnWidth;
    });

    // if there is only one column, set everything to 800, which is full width of MUI Drawer.
    if (MIN_TABLE_WIDTH < 800) {
      MIN_TABLE_WIDTH = 800;
      columns.forEach((column) => {
        column.columnWidth = 800;
      });
    }

    return (
      <AutoSizer style={{ width: '100%', minWidth: '800px' }}>
        {({ height }) => (
          <Table
            height={height}
            width={MIN_TABLE_WIDTH}
            rowHeight={rowHeight}
            gridStyle={{
              direction: 'inherit',
            }}
            headerHeight={headerHeight}
            {...tableProps}
            rowClassName={this.getRowClassName}
          >
            {columns.map(({ dataKey, columnWidth, ...other }, index) => {
              return (
                <Column
                  key={dataKey}
                  width={Math.max(MIN_TABLE_WIDTH, columnWidth)}
                  columnData={columns[index]}
                  headerRenderer={(headerProps) =>
                    this.headerRenderer({
                      ...headerProps,
                      columnIndex: index,
                    })
                  }
                  className={classes.flexContainer}
                  cellRenderer={this.cellRenderer}
                  dataKey={dataKey}
                  {...other}
                />
              );
            })}
          </Table>
        )}
      </AutoSizer>
    );
  }
}

const DataTable = styled(PreStyledDataTable)(styles);

export default React.memo(DataTable);

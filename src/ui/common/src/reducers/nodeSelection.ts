import { createSlice, PayloadAction } from '@reduxjs/toolkit';

import { ArtifactType } from '../utils/artifacts';
import { OperatorType } from '../utils/operators';

export enum NodeType {
  TableArtifact = 'tableArtifact',
  NumericArtifact = 'numericArtifact',
  BoolArtifact = 'boolArtifact',
  JsonArtifact = 'jsonArtifact',
  StringArtifact = 'stringArtifact',
  ImageArtifact = 'imageArtifact',
  DictArtifact = 'dictArtifact',
  GenericArtifact = 'genericArtifact',
  ExtractOp = 'extractOp',
  LoadOp = 'loadOp',
  FunctionOp = 'functionOp',
  MetricOp = 'metricOp',
  CheckOp = 'checkOp',
  ParamOp = 'paramOp', // These operators are hidden from the user.
  None = '', // No node is currently selected.
}

export const OperatorTypeToNodeTypeMap: { [key in OperatorType]: NodeType } = {
  [OperatorType.Extract]: NodeType.ExtractOp,
  [OperatorType.Load]: NodeType.LoadOp,
  [OperatorType.Metric]: NodeType.MetricOp,
  [OperatorType.Function]: NodeType.FunctionOp,
  [OperatorType.Check]: NodeType.CheckOp,
  [OperatorType.Param]: NodeType.ParamOp,
  [OperatorType.SystemMetric]: NodeType.MetricOp,
} as const;

export const ArtifactTypeToNodeTypeMap: { [key in ArtifactType]: NodeType } = {
  [ArtifactType.Table]: NodeType.TableArtifact,
  [ArtifactType.Numeric]: NodeType.NumericArtifact,
  [ArtifactType.Bool]: NodeType.BoolArtifact,
  [ArtifactType.Json]: NodeType.JsonArtifact,
  [ArtifactType.String]: NodeType.StringArtifact,
  [ArtifactType.Dict]: NodeType.DictArtifact,
  [ArtifactType.Tuple]: NodeType.DictArtifact,
  [ArtifactType.Image]: NodeType.ImageArtifact,
  [ArtifactType.Bytes]: NodeType.GenericArtifact,
  [ArtifactType.Picklable]: NodeType.GenericArtifact,
} as const;

export type SelectedNode = {
  id: string;
  type: NodeType;
};

export interface NodeSelectionState {
  selected: SelectedNode;
}

const initialState: NodeSelectionState = {
  selected: { id: '', type: NodeType.None },
};

export const propertiesSideSheetSlice = createSlice({
  name: 'propertiesSideSheet',
  initialState,
  reducers: {
    selectNode: (state, { payload }: PayloadAction<SelectedNode>) => {
      state.selected = payload;
    },
    resetSelectedNode: (state) => {
      state.selected = initialState.selected;
    },
  },
});

export const { selectNode, resetSelectedNode } =
  propertiesSideSheetSlice.actions;

export default propertiesSideSheetSlice.reducer;

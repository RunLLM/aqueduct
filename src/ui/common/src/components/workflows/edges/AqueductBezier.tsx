import React from 'react';
import { EdgeProps, getBezierPath } from 'reactflow';

import { theme } from '../../../styles/theme/theme';

export const AqueductBezier: React.FC<EdgeProps> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
}) => {
  // path, labelX, labelY, offsetX, offsetY
  const edgePathData = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });
  const edgePath = edgePathData[0];
  const color: string = style['color'] ?? (theme.palette.darkGray as string);

  return (
    <>
      <defs>
        <marker
          id="arrow-closed"
          // viewBox="-10 -10 20 20" TODO: investigate linter complaint: Invalid property 'viewBox' found on tag 'marker', but it is only allowed on: svg
          refX="0"
          refY="0"
          markerWidth="12.5"
          markerHeight="12.5"
          orient="auto"
        >
          {/* NOTE: This edge definition is copied from ReactFlow's but is redefined here so we can change the color. */}
          <polyline
            stroke={color}
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="1"
            fill={color}
            points="-5,-4 0,0 -5,4 -5,-4"
          ></polyline>
        </marker>
      </defs>
      <path
        id={id}
        style={{ stroke: color, strokeWidth: 2 }}
        className="react-flow__edge-path"
        d={edgePath}
        markerEnd="url(#arrow-closed)"
      />
    </>
  );
};

export default AqueductBezier;

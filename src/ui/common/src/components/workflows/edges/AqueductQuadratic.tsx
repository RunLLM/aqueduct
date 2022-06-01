import React from 'react';

import { theme } from '../../../styles/theme/theme';

type AqueductQuadraticProps = {
  id: string;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: string;
  targetPosition: string;
  arrowHeadType: string;
  markerEndId: string;
  // These are types that we're required to have here because ReactFlow
  // passes them in, but we don't use them (currently).
  style?: Record<string, string>;
  data: Record<string, string>;
};

const AqueductQuadratic: React.FC<AqueductQuadraticProps> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  data = {},
  style = {},
}) => {
  const curveMaxHeight = data.curveMaxHeight as unknown as number;

  const midX = (sourceX + targetX) / 2;
  const midY = (sourceY + targetY) / 2 - curveMaxHeight;
  const edgePath = `M${sourceX},${sourceY} Q${midX},${midY} ${targetX},${targetY}`;

  const color: string = style['color'] ?? (theme.palette.darkGray as string);

  return (
    <>
      <defs>
        <marker
          id="arrow-closed"
          viewBox="-10 -10 20 20"
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

export default AqueductQuadratic;

import React from 'react';
import { useNavigate } from 'react-router-dom';

import { OperatorResultResponse } from '../../handlers/responses/operator';

type Props = {
  operator?: OperatorResultResponse;
  children?: React.ReactElement | React.ReactElement[];
};

const RequireOperator: React.FC<Props> = ({ operator, children }) => {
  const navigate = useNavigate();
  if (!operator) {
    navigate('/404');
    return null;
  }

  return <>{children}</>;
};

export default RequireOperator;

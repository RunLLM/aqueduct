import * as React from 'react';

import { DetailsPageHeader } from '../../src/components/pages/components/DetailsPageHeader';
import ExecutionStatus from '../../src/utils/shared';
import { render } from '../testUtils';

describe('it', () => {
  it('renders DetailsPageHeader component', () => {
    const now = new Date().toString();
    render(
      <DetailsPageHeader
        name={'TestUser'}
        status={ExecutionStatus.Succeeded}
        createdAt={now}
        sourceLocation={'/src/test/e23'}
      />
    );
  });
});

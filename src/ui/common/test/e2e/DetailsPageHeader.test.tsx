import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { mockUser, render, screen } from '../testUtils';
import { DetailsPageHeader } from '../../src/components/pages/components/DetailsPageHeader';
import ExecutionStatus from '../../src/utils/shared';

describe('it', () => {
  it('renders DetailsPageHeader component', () => {
    const now = new Date().toString();
    render(<DetailsPageHeader name={'TestUser'} status={ExecutionStatus.Succeeded} createdAt={now} sourceLocation={'/src/test/e23'} />);
  });
});

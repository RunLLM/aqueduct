import '@testing-library/jest-dom'
import React from 'react';
import { render, screen, mockUser } from '../testUtils';
import GettingStartedTutorial from '../../src/components/Cards/GettingStartedTutorial';

describe('it', () => {
  it('renders GettingStartedTutorial component', () => {
    render(<GettingStartedTutorial user={mockUser} />);
    // To log contents of screen:
    // screen.debug();

    // Check that wave emoji is in the document
    expect(screen.getByText('👋')).toBeInTheDocument();

    // Check that welcome message is in the document
    expect(screen.getByText('Welcome ' + mockUser.given_name + '!')).toBeInTheDocument();
  });
});

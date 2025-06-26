import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';

describe('App', () => {
  it('renders the main UI elements', () => {
    render(<App />);
    expect(screen.getByText('AI UI Generator')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your prompt...')).toBeInTheDocument();
    expect(screen.getByText('Generate')).toBeInTheDocument();
    expect(screen.getByText('Stream')).toBeInTheDocument();
    expect(screen.getByText('Validate')).toBeInTheDocument();
  });
});

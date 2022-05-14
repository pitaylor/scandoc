import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

test('renders Scan button', () => {
  render(<App />);
  const buttonElement = screen.getByText('Scan');
  expect(buttonElement).toBeInTheDocument();
});

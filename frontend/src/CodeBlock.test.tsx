import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { vi } from 'vitest';
import { CodeBlock } from './CodeBlock';

Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined),
  },
});

describe('CodeBlock', () => {
  const code = '<div>Hello World</div>';

  it('renders code and copy button', () => {
    render(<CodeBlock code={code} language="html" />);
    expect(screen.getByText('Copy')).toBeInTheDocument();
    expect(screen.getByText('Hello World')).toBeInTheDocument();
  });

  it('copies code to clipboard and shows feedback', async () => {
    render(<CodeBlock code={code} language="html" />);
    const button = screen.getByRole('button', { name: /copy code to clipboard/i });
    fireEvent.click(button);
    await waitFor(() => expect(screen.getByText('Copied!')).toBeInTheDocument());
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(code);
  });
});

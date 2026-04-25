import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import HelpCenterPage from '../HelpCenterPage';

describe('HelpCenterPage', () => {
  it('shows strict fuel-pack purchase guidance', () => {
    render(
      <MemoryRouter>
        <HelpCenterPage />
      </MemoryRouter>
    );

    const question = screen.getByText('如何购买Token？');
    expect(question).toBeInTheDocument();
    fireEvent.click(question);
    expect(screen.getByText(/加油包（Token）不可单独购买/)).toBeInTheDocument();
  });
});

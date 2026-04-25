import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import UserAgreementPage from '../UserAgreementPage';

describe('UserAgreementPage', () => {
  it('contains strict and fuel-pack restriction terms', () => {
    render(
      <MemoryRouter>
        <UserAgreementPage />
      </MemoryRouter>
    );

    expect(screen.getByText('用户服务协议')).toBeInTheDocument();
    expect(screen.getByText(/在 strict 权益规则下，加油包不可单独购买/)).toBeInTheDocument();
  });
});

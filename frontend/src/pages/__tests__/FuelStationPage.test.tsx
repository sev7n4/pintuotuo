import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import FuelStationPage from '../FuelStationPage';

jest.mock('@/services/fuelStation', () => ({
  fuelStationService: {
    getPublicConfig: jest.fn().mockRejectedValue(new Error('mock no config')),
  },
}));

jest.mock('@/services/sku', () => ({
  skuService: {
    getPublicSKU: jest.fn(),
  },
}));

jest.mock('@/stores/cartStore', () => ({
  useCartStore: () => ({
    addItem: jest.fn(),
  }),
}));

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
}));

describe('FuelStationPage', () => {
  it('shows station title and strict purchase rule copy', async () => {
    render(
      <MemoryRouter>
        <FuelStationPage />
      </MemoryRouter>
    );

    expect(await screen.findByText('智燃加油站')).toBeInTheDocument();
    fireEvent.click(screen.getByText('购买规则'));
    expect(screen.getByText(/加油包不可单独购买/)).toBeInTheDocument();
  });
});

import { render, screen } from '@testing-library/react';
import { ProductCoverMedia } from './ProductCoverMedia';

jest.mock('./ProductCoverMedia.module.css', () => ({
  root: 'root',
  rootGrid: 'rootGrid',
  rootWide: 'rootWide',
  rootHome: 'rootHome',
  rootHero: 'rootHero',
  coverImg: 'coverImg',
  placeholder: 'placeholder',
  placeholderCompact: 'placeholderCompact',
  placeholderHero: 'placeholderHero',
  placeholderText: 'placeholderText',
  productBrandBadge: 'productBrandBadge',
  badgeGrid: 'badgeGrid',
  badgeWide: 'badgeWide',
  badgeHero: 'badgeHero',
  productBrandLogo: 'productBrandLogo',
  logoGrid: 'logoGrid',
  logoWide: 'logoWide',
  logoHero: 'logoHero',
}));

describe('ProductCoverMedia', () => {
  it('shows two-character fallback when no image or provider', () => {
    render(<ProductCoverMedia fallbackTitle="深度求索" resetKey={1} variant="grid" />);
    expect(screen.getByText('深度')).toBeInTheDocument();
  });
});

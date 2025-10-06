import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TimeWindowSelector } from './TimeWindowSelector';

describe('TimeWindowSelector', () => {
  it('should render with default selected value', () => {
    render(
      <TimeWindowSelector selectedHours={48} onHoursChange={vi.fn()} />
    );

    const select = screen.getByLabelText(/time window/i);
    expect(select).toHaveValue('48');
  });

  it('should display all time window options', () => {
    render(
      <TimeWindowSelector selectedHours={48} onHoursChange={vi.fn()} />
    );

    expect(screen.getByText('Last 12 Hours')).toBeInTheDocument();
    expect(screen.getByText('Last 24 Hours')).toBeInTheDocument();
    expect(screen.getByText('Last 48 Hours')).toBeInTheDocument();
    expect(screen.getByText('Last 3 Days (72 Hours)')).toBeInTheDocument();
    expect(screen.getByText('Last 7 Days (168 Hours)')).toBeInTheDocument();
    expect(screen.getByText('Last 30 Days (720 Hours)')).toBeInTheDocument();
  });

  it('should call onHoursChange when selection changes', async () => {
    const handleChange = vi.fn();
    const user = userEvent.setup();

    render(
      <TimeWindowSelector selectedHours={48} onHoursChange={handleChange} />
    );

    const select = screen.getByLabelText(/time window/i);
    await user.selectOptions(select, '24');

    expect(handleChange).toHaveBeenCalledWith(24);
  });

  it('should handle different selected values', () => {
    const { rerender } = render(
      <TimeWindowSelector selectedHours={12} onHoursChange={vi.fn()} />
    );

    expect(screen.getByLabelText(/time window/i)).toHaveValue('12');

    rerender(
      <TimeWindowSelector selectedHours={168} onHoursChange={vi.fn()} />
    );

    expect(screen.getByLabelText(/time window/i)).toHaveValue('168');
  });
});

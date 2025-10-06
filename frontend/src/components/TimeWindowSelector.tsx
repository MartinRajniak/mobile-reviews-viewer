interface TimeWindowSelectorProps {
  selectedHours: number;
  onHoursChange: (hours: number) => void;
}

const TIME_WINDOWS = [
  { hours: 12, label: 'Last 12 Hours' },
  { hours: 24, label: 'Last 24 Hours' },
  { hours: 48, label: 'Last 48 Hours' },
  { hours: 72, label: 'Last 3 Days (72 Hours)' },
  { hours: 168, label: 'Last 7 Days (168 Hours)' },
  { hours: 720, label: 'Last 30 Days (720 Hours)' },
];

export const TimeWindowSelector: React.FC<TimeWindowSelectorProps> = ({
  selectedHours,
  onHoursChange,
}) => {
  return (
    <div className="time-window-selector">
      <label htmlFor="time-window-select">Time Window:</label>
      <select
        id="time-window-select"
        value={selectedHours}
        onChange={(e) => onHoursChange(Number(e.target.value))}
      >
        {TIME_WINDOWS.map((window) => (
          <option key={window.hours} value={window.hours}>
            {window.label}
          </option>
        ))}
      </select>
    </div>
  );
};

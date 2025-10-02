interface AppSelectorProps {
  selectedAppId: string;
  onAppChange: (appId: string) => void;
}

const APPS = [
  { id: '389801252', name: 'Instagram' },
  { id: '447188370', name: 'Snapchat' },
  { id: '310633997', name: 'WhatsApp' },
];

export const AppSelector: React.FC<AppSelectorProps> = ({
  selectedAppId,
  onAppChange,
}) => {
  return (
    <div className="app-selector">
      <label htmlFor="app-select">Select App:</label>
      <select
        id="app-select"
        value={selectedAppId}
        onChange={(e) => onAppChange(e.target.value)}
      >
        {APPS.map((app) => (
          <option key={app.id} value={app.id}>
            {app.name}
          </option>
        ))}
      </select>
    </div>
  );
};
import { useState } from 'react';
import { AppSelector } from './components/AppSelector';
import { TimeWindowSelector } from './components/TimeWindowSelector';
import { ReviewList } from './components/ReviewList';
import './App.css';

function App() {
  const [selectedAppId, setSelectedAppId] = useState('389801252');
  const [selectedHours, setSelectedHours] = useState(48);

  return (
    <div className="app">
      <header className="app-header">
        <h1>App Store Reviews Viewer</h1>
        <AppSelector
          selectedAppId={selectedAppId}
          onAppChange={setSelectedAppId}
        />
        <TimeWindowSelector
          selectedHours={selectedHours}
          onHoursChange={setSelectedHours}
        />
      </header>
      <main className="app-main">
        <ReviewList appId={selectedAppId} hours={selectedHours} />
      </main>
    </div>
  );
}

export default App;
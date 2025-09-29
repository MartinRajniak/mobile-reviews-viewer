import { useState } from 'react';
import { AppSelector } from './components/AppSelector';
import { ReviewList } from './components/ReviewList';
import './App.css';

function App() {
  const [selectedAppId, setSelectedAppId] = useState('595068606');

  return (
    <div className="app">
      <header className="app-header">
        <h1>App Store Reviews Viewer</h1>
        <AppSelector
          selectedAppId={selectedAppId}
          onAppChange={setSelectedAppId}
        />
      </header>
      <main className="app-main">
        <ReviewList appId={selectedAppId} />
      </main>
    </div>
  );
}

export default App;
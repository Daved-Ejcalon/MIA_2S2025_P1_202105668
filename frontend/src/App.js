import React from 'react';
import './App.css';
import MIAConsole from './components/MIAConsole';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <div className="header-content">
          <h1>PROYECTO I - 202105668</h1>
          <span className="subtitle">File System Manager</span>
        </div>
      </header>
      <main className="App-main">
        <MIAConsole />
      </main>
    </div>
  );
}

export default App;

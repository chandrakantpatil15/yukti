import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Resources from './pages/Resources';
import CostAnalysis from './pages/CostAnalysis';
import Navbar from './components/Navbar';
import './App.css';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <Navbar />
        <main>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/resources" element={<Resources />} />
            <Route path="/cost-analysis" element={<CostAnalysis />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
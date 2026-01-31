import React from 'react';

const TestAdmin: React.FC = () => {
  return (
    <div style={{ padding: '40px' }}>
      <h1>Test Admin Page</h1>
      <p>If you see this, React is working!</p>
      <button onClick={() => {
        fetch((process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1').replace(/\/api\/v1\/?$/, '') + '/api/admin/customers')
          .then(r => r.json())
          .then(d => console.log('API Response:', d))
          .catch(e => console.error('API Error:', e));
      }}>
        Test API
      </button>
    </div>
  );
};

export default TestAdmin;

import { useState } from 'react';
import adminApi from '../../services/adminApi';

export default function ImpersonationBanner() {
  const [ending, setEnding] = useState(false);

  const handleEndImpersonation = async () => {
    if (!confirm('End impersonation and return to admin portal?')) return;
    
    setEnding(true);
    
    try {
      await adminApi.post('/api/admin/end-impersonation');
      localStorage.removeItem('impersonation_token');
      window.location.href = '/admin/dashboard';
    } catch (err) {
      alert('Failed to end impersonation');
      setEnding(false);
    }
  };

  return (
    <div className="bg-yellow-500 text-black px-6 py-3 flex items-center justify-between sticky top-0 z-50">
      <div className="flex items-center gap-2">
        <span className="font-bold">⚠️ IMPERSONATION MODE</span>
        <span>You are viewing this account as an administrator</span>
      </div>
      <button
        onClick={handleEndImpersonation}
        disabled={ending}
        className="px-4 py-1 bg-black text-white rounded hover:bg-gray-800 disabled:opacity-50"
      >
        {ending ? 'Ending...' : 'End Impersonation'}
      </button>
    </div>
  );
}

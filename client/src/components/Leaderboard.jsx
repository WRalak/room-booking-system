import React from 'react';
import gamificationService from '../services/gamificationService';

export default function Leaderboard() {
  const [leaders] = React.useState(() => gamificationService.getLeaderboard());

  return (
    <div className="p-4 bg-white rounded shadow">
      <h3 className="text-lg font-semibold mb-2">Leaderboard</h3>
      <ol className="list-decimal list-inside">
        {leaders.length === 0 && <li>No activity yet</li>}
        {leaders.map((l) => (
          <li key={l.userId} className="py-1">
            <span className="font-medium">{l.userId}</span>: {l.points} pts
          </li>
        ))}
      </ol>
    </div>
  );
}

import React from 'react';

const Leaderboard = () => {
  return (
    <div className="leaderboard">
      <h2>Leaderboard</h2>
      <div className="leaderboard-list">
        <div className="leaderboard-item">
          <span className="rank">1</span>
          <span className="name">Player 1</span>
          <span className="score">10000</span>
        </div>
        <div className="leaderboard-item">
          <span className="rank">2</span>
          <span className="name">Player 2</span>
          <span className="score">9000</span>
        </div>
        <div className="leaderboard-item">
          <span className="rank">3</span>
          <span className="name">Player 3</span>
          <span className="score">8000</span>
        </div>
      </div>
    </div>
  );
};

export default Leaderboard;

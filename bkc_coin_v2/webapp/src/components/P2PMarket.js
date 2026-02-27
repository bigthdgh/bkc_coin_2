import React from 'react';

const P2PMarket = () => {
  return (
    <div className="p2p-market">
      <h2>P2P Market</h2>
      <div className="market-list">
        <div className="market-item">
          <span className="item-name">BKC</span>
          <span className="price">1.00 TON</span>
          <button className="buy-btn">Buy</button>
        </div>
        <div className="market-item">
          <span className="item-name">NFT</span>
          <span className="price">100 BKC</span>
          <button className="buy-btn">Buy</button>
        </div>
      </div>
    </div>
  );
};

export default P2PMarket;

import React from 'react';

const Bank = () => {
  return (
    <div className="bank">
      <h2>Bank</h2>
      <div className="bank-options">
        <div className="bank-option">
          <h3>Deposit</h3>
          <p>Deposit your BKC coins</p>
          <button>Deposit</button>
        </div>
        <div className="bank-option">
          <h3>Withdraw</h3>
          <p>Withdraw your BKC coins</p>
          <button>Withdraw</button>
        </div>
        <div className="bank-option">
          <h3>Loan</h3>
          <p>Get a loan with collateral</p>
          <button>Get Loan</button>
        </div>
      </div>
    </div>
  );
};

export default Bank;

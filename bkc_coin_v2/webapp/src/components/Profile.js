import React from 'react';

const Profile = () => {
  return (
    <div className="profile">
      <h2>Profile</h2>
      <div className="profile-info">
        <div className="avatar">👤</div>
        <div className="user-details">
          <h3>User Name</h3>
          <p>Level: 10</p>
          <p>Balance: 1000 BKC</p>
        </div>
      </div>
    </div>
  );
};

export default Profile;

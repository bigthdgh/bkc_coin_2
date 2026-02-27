import { useState, useEffect } from 'react';

export const useTelegram = () => {
  const [tg, setTg] = useState(null);
  const [user, setUser] = useState(null);

  useEffect(() => {
    if (window.Telegram && window.Telegram.WebApp) {
      const telegram = window.Telegram.WebApp;
      setTg(telegram);
      telegram.ready();
      
      if (telegram.initDataUnsafe && telegram.initDataUnsafe.user) {
        setUser(telegram.initDataUnsafe.user);
      }
    }
  }, []);

  return { tg, user };
};

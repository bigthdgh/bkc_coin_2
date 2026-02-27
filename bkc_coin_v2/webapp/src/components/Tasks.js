import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  CheckCircle, 
  Clock, 
  XCircle, 
  ExternalLink, 
  Gift, 
  Users, 
  Video, 
  MessageCircle, 
  Twitter,
  Shield,
  Timer
} from 'lucide-react';
import toast from 'react-hot-toast';

const Tasks = ({ user, onNavigate, apiService }) => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [verifying, setVerifying] = useState({});
  const [secretCode, setSecretCode] = useState('');
  const [twitterUsername, setTwitterUsername] = useState('');
  const [countdown, setCountdown] = useState({});

  useEffect(() => {
    loadTasks();
  }, [user.id]);

  useEffect(() => {
    // Update countdowns every second
    const interval = setInterval(() => {
      updateCountdowns();
    }, 1000);

    return () => clearInterval(interval);
  }, [tasks]);

  const loadTasks = async () => {
    try {
      // Получаем user_id
      const getUserId = () => {
        if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initDataUnsafe) {
          const telegramUser = window.Telegram.WebApp.initDataUnsafe.user;
          if (telegramUser && telegramUser.id) {
            return telegramUser.id;
          }
        }
        
        const savedUserId = localStorage.getItem('bkc_user_id');
        if (savedUserId) {
          return parseInt(savedUserId);
        }
        
        const tempId = Math.floor(Math.random() * 1000000) + 1000000;
        localStorage.setItem('bkc_user_id', tempId);
        return tempId;
      };

      const userId = getUserId();
      
      // Запрашиваем реальные задания с API
      const response = await fetch(`/api/v1/tasks/list?user_id=${userId}`);
      const result = await response.json();
      
      let tasks = [];
      
      if (result.success && result.tasks) {
        // Используем реальные задания с API
        tasks = result.tasks;
      } else {
        // Fallback на базовые задания если API недоступен
        tasks = [
          {
            id: 'telegram',
            title: 'Подпишись на Telegram канал',
            description: 'Подпишись на наш основной канал для получения новостей',
            reward: 2000,
            icon: MessageCircle,
            color: '#0088cc',
            status: 'pending',
            action: 'verify_telegram',
            link: 'https://t.me/bkc_coin_official',
            required: true
          }
        ];
      }
      
      setTasks(tasks);
    } catch (error) {
      console.error('Failed to load tasks:', error);
      toast.error('Не удалось загрузить задания');
      
      // Fallback на базовые задания
      setTasks([
        {
          id: 'telegram',
          title: 'Подпишись на Telegram канал',
          description: 'Подпишись на наш основной канал для получения новостей',
          reward: 2000,
          icon: MessageCircle,
          color: '#0088cc',
          status: 'pending',
          action: 'verify_telegram',
          link: 'https://t.me/bkc_coin_official',
          required: true
        }
      ]);
    } finally {
      setLoading(false);
    }
  };

  const updateCountdowns = () => {
    const newCountdown = {};
    
    tasks.forEach(task => {
      if (task.status === 'pending' && task.timerEnd) {
        const remaining = Math.max(0, task.timerEnd - Date.now());
        if (remaining > 0) {
          newCountdown[task.id] = Math.ceil(remaining / 1000);
        }
      }
    });
    
    setCountdown(newCountdown);
  };

  const handleTaskAction = async (task) => {
    if (task.status === 'completed') {
      toast.success('Задание уже выполнено!');
      return;
    }

    if (task.status === 'expired') {
      toast.error('Срок действия задания истёк');
      return;
    }

    // Handle different task types
    switch (task.action) {
      case 'verify_telegram':
        await handleTelegramVerification(task);
        break;
      case 'verify_tiktok':
        await handleTikTokVerification(task);
        break;
      case 'verify_twitter':
        await handleTwitterVerification(task);
        break;
      case 'verify_discord':
        await handleDiscordVerification(task);
        break;
      default:
        toast.error('Неизвестный тип задания');
    }
  };

  const handleTelegramVerification = async (task) => {
    setVerifying(prev => ({ ...prev, [task.id]: true }));
    
    try {
      // Получаем user_id
      const getUserId = () => {
        if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initDataUnsafe) {
          const telegramUser = window.Telegram.WebApp.initDataUnsafe.user;
          if (telegramUser && telegramUser.id) {
            return telegramUser.id;
          }
        }
        
        const savedUserId = localStorage.getItem('bkc_user_id');
        if (savedUserId) {
          return parseInt(savedUserId);
        }
        
        const tempId = Math.floor(Math.random() * 1000000) + 1000000;
        localStorage.setItem('bkc_user_id', tempId);
        return tempId;
      };

      const userId = getUserId();
      
      // Open Telegram channel
      window.open(task.link, '_blank');
      
      // Проверяем подписку через API
      setTimeout(async () => {
        try {
          const response = await fetch('/api/v1/tasks/verify-telegram', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              user_id: userId,
              task_id: task.id,
              channel_username: 'bkc_coin_official' // Твой канал
            })
          });
          
          const result = await response.json();
          
          if (result.success) {
            toast.success(`✅ Получено ${task.reward} BKC!`);
            updateTaskStatus(task.id, 'completed');
          } else {
            toast.error(result.message || 'Вы не подписаны на канал');
          }
        } catch (error) {
          toast.error('Ошибка проверки подписки');
        } finally {
          setVerifying(prev => ({ ...prev, [task.id]: false }));
        }
      }, 2000);
      
    } catch (error) {
      toast.error('Ошибка при проверке');
      setVerifying(prev => ({ ...prev, [task.id]: false }));
    }
  };

  const handleTikTokVerification = async (task) => {
    if (!secretCode.trim()) {
      toast.error('Введите секретный код');
      return;
    }

    setVerifying(prev => ({ ...prev, [task.id]: true }));
    
    try {
      // Получаем user_id
      const getUserId = () => {
        if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initDataUnsafe) {
          const telegramUser = window.Telegram.WebApp.initDataUnsafe.user;
          if (telegramUser && telegramUser.id) {
            return telegramUser.id;
          }
        }
        
        const savedUserId = localStorage.getItem('bkc_user_id');
        if (savedUserId) {
          return parseInt(savedUserId);
        }
        
        const tempId = Math.floor(Math.random() * 1000000) + 1000000;
        localStorage.setItem('bkc_user_id', tempId);
        return tempId;
      };

      const userId = getUserId();
      
      // Отправляем запрос на проверку кода
      const response = await fetch('/api/v1/tasks/verify-tiktok-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          task_id: task.id,
          secret_code: secretCode.toUpperCase()
        })
      });
      
      const result = await response.json();
      
      if (result.success) {
        toast.success(`✅ Получено ${task.reward} BKC!`);
        updateTaskStatus(task.id, 'completed');
        setSecretCode('');
      } else {
        toast.error(result.message || 'Неверный секретный код');
      }
    } catch (error) {
      toast.error('Ошибка проверки секретного кода');
    } finally {
      setVerifying(prev => ({ ...prev, [task.id]: false }));
    }
  };

  const handleTwitterVerification = async (task) => {
    if (!twitterUsername.trim()) {
      toast.error('Введите имя пользователя Twitter');
      return;
    }

    setVerifying(prev => ({ ...prev, [task.id]: true }));
    
    try {
      // Получаем user_id
      const getUserId = () => {
        if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initDataUnsafe) {
          const telegramUser = window.Telegram.WebApp.initDataUnsafe.user;
          if (telegramUser && telegramUser.id) {
            return telegramUser.id;
          }
        }
        
        const savedUserId = localStorage.getItem('bkc_user_id');
        if (savedUserId) {
          return parseInt(savedUserId);
        }
        
        const tempId = Math.floor(Math.random() * 1000000) + 1000000;
        localStorage.setItem('bkc_user_id', tempId);
        return tempId;
      };

      const userId = getUserId();
      
      // Шаг 1: Отправляем сигнал о старте задания
      const startResponse = await fetch('/api/v1/tasks/start-twitter-task', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          task_id: task.id,
          username: twitterUsername
        })
      });
      
      const startResult = await startResponse.json();
      
      if (!startResult.success) {
        toast.error(startResult.message || 'Ошибка старта задания');
        setVerifying(prev => ({ ...prev, [task.id]: false }));
        return;
      }
      
      // Шаг 2: Открываем Twitter
      window.open(`https://twitter.com/${twitterUsername}`, '_blank');
      
      // Шаг 3: Показываем таймер на 15 секунд
      toast.info('Подпишитесь и подождите 15 секунд...');
      
      // Запускаем таймер
      let countdown = 15;
      const timerInterval = setInterval(() => {
        countdown--;
        setCountdown(prev => ({ ...prev, [task.id]: countdown }));
        
        if (countdown <= 0) {
          clearInterval(timerInterval);
        }
      }, 1000);
      
      // Шаг 4: Через 15 секунд разрешаем забрать награду
      setTimeout(async () => {
        try {
          // Шаг 5: Пытаемся забрать награду с проверкой времени
          const claimResponse = await fetch('/api/v1/tasks/claim-twitter-reward', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              user_id: userId,
              task_id: task.id,
              username: twitterUsername
            })
          });
          
          const claimResult = await claimResponse.json();
          
          if (claimResult.success) {
            toast.success(`✅ Получено ${task.reward} BKC!`);
            updateTaskStatus(task.id, 'completed');
            setTwitterUsername('');
          } else {
            toast.error(claimResult.message || 'Слишком быстро! Подождите еще немного.');
          }
        } catch (error) {
          toast.error('Ошибка получения награды');
        } finally {
          setVerifying(prev => ({ ...prev, [task.id]: false }));
          setCountdown(prev => ({ ...prev, [task.id]: null }));
        }
      }, 15000);
      
    } catch (error) {
      toast.error('Ошибка при проверке');
      setVerifying(prev => ({ ...prev, [task.id]: false }));
    }
  };

  const handleDiscordVerification = async (task) => {
    setVerifying(prev => ({ ...prev, [task.id]: true }));
    
    try {
      // Open Discord OAuth
      window.open(task.link, '_blank');
      
      // Simulate OAuth verification (in real app, this would handle OAuth flow)
      setTimeout(async () => {
        try {
          // Mock API call
          const result = { success: true, reward: task.reward };
          
          if (result.success) {
            toast.success(`✅ Получено ${task.reward} BKC!`);
            updateTaskStatus(task.id, 'completed');
          } else {
            toast.error('Вы не состоите в Discord сервере');
          }
        } catch (error) {
          toast.error('Ошибка проверки Discord');
        } finally {
          setVerifying(prev => ({ ...prev, [task.id]: false }));
        }
      }, 3000);
      
    } catch (error) {
      toast.error('Ошибка при проверке');
      setVerifying(prev => ({ ...prev, [task.id]: false }));
    }
  };

  const claimTwitterReward = async (task) => {
    setVerifying(prev => ({ ...prev, [task.id]: true }));
    
    try {
      // Mock API call to claim reward
      const result = { success: true, reward: task.reward };
      
      if (result.success) {
        toast.success(`✅ Получено ${task.reward} BKC!`);
        updateTaskStatus(task.id, 'completed');
        setTwitterUsername('');
      } else {
        toast.error('Слишком рано! Подождите еще');
      }
    } catch (error) {
      toast.error('Ошибка получения награды');
    } finally {
      setVerifying(prev => ({ ...prev, [task.id]: false }));
    }
  };

  const updateTaskStatus = (taskId, status, additionalData = {}) => {
    setTasks(prev => prev.map(task => 
      task.id === taskId 
        ? { ...task, status, ...additionalData }
        : task
    ));
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'completed':
        return <CheckCircle size={20} className="text-green-500" />;
      case 'expired':
        return <XCircle size={20} className="text-red-500" />;
      case 'pending':
        return <Clock size={20} className="text-yellow-500" />;
      default:
        return <Clock size={20} className="text-gray-500" />;
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'completed':
        return 'Выполнено';
      case 'expired':
        return 'Истёкло';
      case 'pending':
        return 'Ожидание';
      default:
        return 'Доступно';
    }
  };

  const formatCountdown = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner" />
        <p>Загрузка заданий...</p>
      </div>
    );
  }

  return (
    <div className="tasks-container">
      <div className="tasks-header">
        <h1>Задания</h1>
        <p>Выполняйте задания и получайте BKC монеты</p>
      </div>

      <div className="tasks-grid">
        <AnimatePresence>
          {tasks.map((task) => {
            const Icon = task.icon;
            const isVerifying = verifying[task.id];
            const countdownSeconds = countdown[task.id];
            
            return (
              <motion.div
                key={task.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className={`task-card ${task.status}`}
                whileHover={{ scale: 1.02 }}
                transition={{ duration: 0.2 }}
              >
                <div className="task-header">
                  <div className="task-icon" style={{ backgroundColor: task.color }}>
                    <Icon size={24} className="text-white" />
                  </div>
                  <div className="task-info">
                    <h3>{task.title}</h3>
                    <p>{task.description}</p>
                  </div>
                  <div className="task-status">
                    {getStatusIcon(task.status)}
                  </div>
                </div>

                <div className="task-reward">
                  <Gift size={16} />
                  <span className="reward-amount">{task.reward.toLocaleString()} BKC</span>
                </div>

                {task.status === 'completed' && (
                  <div className="task-completed">
                    <CheckCircle size={16} />
                    <span>Выполнено</span>
                  </div>
                )}

                {task.status === 'expired' && (
                  <div className="task-expired">
                    <XCircle size={16} />
                    <span>Срок истёк</span>
                  </div>
                )}

                {task.status === 'pending' && (
                  <div className="task-actions">
                    {task.requiresCode && (
                      <div className="code-input">
                        <input
                          type="text"
                          placeholder="Секретный код"
                          value={secretCode}
                          onChange={(e) => setSecretCode(e.target.value)}
                          className="code-field"
                        />
                      </div>
                    )}

                    {task.requiresUsername && (
                      <div className="username-input">
                        <input
                          type="text"
                          placeholder="@username"
                          value={twitterUsername}
                          onChange={(e) => setTwitterUsername(e.target.value)}
                          className="username-field"
                        />
                      </div>
                    )}

                    {task.hasTimer && countdownSeconds > 0 ? (
                      <div className="timer-container">
                        <Timer size={16} />
                        <span>{formatCountdown(countdownSeconds)}</span>
                        <button
                          className="claim-button"
                          onClick={() => claimTwitterReward(task)}
                          disabled={isVerifying || countdownSeconds > 0}
                        >
                          {isVerifying ? 'Проверка...' : 'Получить награду'}
                        </button>
                      </div>
                    ) : (
                      <div className="action-buttons">
                        <button
                          className="external-link-button"
                          onClick={() => window.open(task.link, '_blank')}
                        >
                          <ExternalLink size={16} />
                          Перейти
                        </button>
                        
                        <button
                          className="verify-button"
                          onClick={() => handleTaskAction(task)}
                          disabled={isVerifying}
                        >
                          {isVerifying ? 'Проверка...' : 'Проверить'}
                        </button>
                      </div>
                    )}
                  </div>
                )}

                {task.status === 'available' && (
                  <div className="task-actions">
                    <button
                      className="start-button"
                      onClick={() => handleTaskAction(task)}
                      disabled={isVerifying}
                    >
                      {isVerifying ? 'Загрузка...' : 'Начать'}
                    </button>
                  </div>
                )}
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>

      <div className="tasks-info">
        <div className="info-card">
          <Shield size={20} />
          <div>
            <h4>Безопасная верификация</h4>
            <p>Все задания проверяются автоматически и безопасно</p>
          </div>
        </div>
        
        <div className="info-card">
          <Gift size={20} />
          <div>
            <h4>Мгновенные награды</h4>
            <p>Получайте BKC монеты сразу после выполнения заданий</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Tasks;

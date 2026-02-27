import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Crown, Star, Zap, TrendingUp, Calendar, CheckCircle, X, Gift, Award, Users, DollarSign } from 'lucide-react';
import toast from 'react-hot-toast';

const SubscriptionSystem = ({ user, onNavigate, apiService }) => {
  const [activeTab, setActiveTab] = useState('plans');
  const [availablePlans, setAvailablePlans] = useState([]);
  const [userSubscription, setUserSubscription] = useState(null);
  const [subscriptionHistory, setSubscriptionHistory] = useState([]);
  const [userStats, setUserStats] = useState(null);
  const [isProcessing, setIsProcessing] = useState(false);
  
  // Загрузка данных
  useEffect(() => {
    loadInitialData();
    const interval = setInterval(loadInitialData, 30000);
    return () => clearInterval(interval);
  }, [user.id]);

  const loadInitialData = async () => {
    try {
      const [plansData, userSubData, historyData, statsData] = await Promise.all([
        apiService.getSubscriptionPlans(),
        apiService.getUserSubscription(user.id),
        apiService.getSubscriptionHistory(user.id),
        apiService.getSubscriptionStats()
      ]);

      setAvailablePlans(plansData.plans || []);
      setUserSubscription(userSubData.subscription);
      setSubscriptionHistory(historyData.history || []);
      setUserStats(statsData.stats);
    } catch (error) {
      console.error('Failed to load subscription data:', error);
      toast.error('Ошибка загрузки данных');
    }
  };

  // Купить подписку
  const purchaseSubscription = async (planId) => {
    if (isProcessing) return;
    
    const plan = availablePlans.find(p => p.id === planId);
    if (!plan) return;

    // Проверка текущей подписки
    if (userSubscription && userSubscription.status === 'active') {
      toast.error('У вас уже есть активная подписка');
      return;
    }

    setIsProcessing(true);
    
    try {
      const response = await apiService.purchaseSubscription({
        user_id: user.id,
        plan: plan.id
      });

      if (response.success) {
        toast.success(`🎉 Подписка ${plan.name} успешно куплена!`);
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка покупки подписки');
      }
    } catch (error) {
      toast.error('Ошибка при покупке подписки');
    } finally {
      setIsProcessing(false);
    }
  };

  // Отменить подписку
  const cancelSubscription = async () => {
    if (!userSubscription || userSubscription.status !== 'active') {
      toast.error('Нет активной подписки для отмены');
      return;
    }

    try {
      const response = await apiService.cancelSubscription(user.id);
      if (response.success) {
        toast.success('Подписка отменена');
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка отмены подписки');
      }
    } catch (error) {
      toast.error('Ошибка при отмене подписки');
    }
  };

  const getPlanIcon = (planId) => {
    switch (planId) {
      case 'basic':
        return <Star className="w-6 h-6 text-gray-400" />;
      case 'silver':
        return <Award className="w-6 h-6 text-gray-400" />;
      case 'gold':
        return <Crown className="w-6 h-6 text-yellow-400" />;
      default:
        return <Gift className="w-6 h-6 text-gray-400" />;
    }
  };

  const getPlanGradient = (planId) => {
    switch (planId) {
      case 'basic':
        return 'from-gray-600 to-gray-700';
      case 'silver':
        return 'from-purple-600 to-purple-700';
      case 'gold':
        return 'from-yellow-500 to-orange-600';
      default:
        return 'from-gray-600 to-gray-700';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-900 via-purple-900 to-pink-900 p-4">
      {/* Header */}
      <div className="bg-black/20 backdrop-blur-lg rounded-2xl p-6 mb-6 border border-purple-500/20">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2 flex items-center gap-3">
              <Crown className="w-8 h-8 text-yellow-400" />
              Премиум подписки
            </h1>
            <p className="text-gray-400">Basic, Silver, Gold - выберите свой уровень</p>
          </div>
          <div className="text-right">
            {userSubscription && (
              <div className="text-center">
                <div className="text-gray-400 text-sm mb-1">Ваша подписка</div>
                <div className="text-2xl font-bold flex items-center gap-2">
                  {getPlanIcon(userSubscription.plan)}
                  <span className="text-yellow-400 capitalize">{userSubscription.plan}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Основной контент */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Левая панель - План и подписка */}
        <div className="lg:col-span-2 space-y-6">
          {/* Вкладки */}
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-purple-500/20">
            <div className="flex space-x-2 mb-6">
              <button
                onClick={() => setActiveTab('plans')}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  activeTab === 'plans' 
                    ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Gift className="w-5 h-5 mr-2" />
                План подписки
              </button>
              <button
                onClick={() => setActiveTab('history')}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  activeTab === 'history' 
                    ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Calendar className="w-5 h-5 mr-2" />
                История
              </button>
            </div>

            {/* Доступные планы */}
            {activeTab === 'plans' && (
              <div className="space-y-6">
                <div className="text-center mb-6">
                  <h3 className="text-2xl font-bold text-white mb-2">Выберите план подписки</h3>
                  <p className="text-gray-400">Улучшите свои возможности и получите эксклюзивные преимущества</p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {availablePlans.map((plan) => (
                    <motion.div
                      key={plan.id}
                      initial={{ opacity: 0, y: 30 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.5 }}
                      whileHover={{ scale: 1.02 }}
                      className={`relative bg-gradient-to-br ${getPlanGradient(plan.id)} rounded-2xl p-6 border-2 cursor-pointer transform hover:scale-105 transition-all duration-300 ${
                        userSubscription?.plan === plan.id ? 'ring-4 ring-purple-400 ring-opacity-50' : ''
                      }`}
                      onClick={() => purchaseSubscription(plan.id)}
                    >
                      {/* Популярный тег */}
                      {plan.popular && (
                        <div className="absolute -top-2 -right-2 bg-gradient-to-r from-red-500 to-pink-500 text-white text-xs px-3 py-1 rounded-full font-bold">
                          🔥 ПОПУЛЯРНО
                        </div>
                      )}

                      {/* Заголовок плана */}
                      <div className="text-center mb-4">
                        <div className="flex justify-center mb-3">
                          {getPlanIcon(plan.id)}
                        </div>
                        <h4 className="text-2xl font-bold text-white mb-2">{plan.name}</h4>
                      </div>

                      {/* Цена */}
                      <div className="text-center mb-4">
                        <div className="text-4xl font-bold text-white">
                          {plan.price_bkc > 0 ? (
                            <>
                              <span className="text-yellow-400">{plan.price_bkc.toLocaleString()}</span>
                              <span className="text-gray-400 ml-2">BKC</span>
                            </>
                          ) : (
                            <>
                              <span className="text-blue-400">{plan.price_ton}</span>
                              <span className="text-gray-400 ml-2">TON</span>
                            </>
                          )}
                        </div>
                        <div className="text-gray-400 text-sm">
                          за {plan.duration_days} дней
                        </div>
                      </div>

                      {/* Преимущества */}
                      <div className="space-y-2 mb-4">
                        {plan.features.map((feature, index) => (
                          <div key={index} className="flex items-center text-gray-300 text-sm">
                            <CheckCircle className="w-4 h-4 mr-2 text-green-400" />
                            {feature}
                          </div>
                        ))}
                      </div>

                      {/* Кнопка действия */}
                      {userSubscription?.plan === plan.id ? (
                        <div className="text-center">
                          <div className="text-green-400 font-bold mb-2">✅ Активна</div>
                          <button
                            onClick={cancelSubscription}
                            className="bg-red-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-red-700 transition-all duration-300"
                          >
                            Отменить
                          </button>
                        </div>
                      ) : (
                        <button
                          disabled={isProcessing}
                          className={`w-full bg-gradient-to-r from-green-500 to-emerald-600 text-white font-bold py-3 rounded-lg transition-all duration-300 ${
                            isProcessing ? 'opacity-50 cursor-not-allowed' : 'hover:from-green-600 hover:to-emerald-700'
                          }`}
                        >
                          {isProcessing ? (
                            <>
                              <div className="inline-block animate-spin rounded-full h-5 w-5 border-b-2 border-white border-t-transparent"></div>
                              Обработка...
                            </>
                          ) : (
                            <>
                              <DollarSign className="w-5 h-5 mr-2" />
                              Купить подписку
                            </>
                          )}
                        </button>
                      )}
                    </motion.div>
                  ))}
                </div>
              </div>
            )}

            {/* Текущая подписка */}
            {activeTab === 'history' && (
              <div className="space-y-6">
                <div className="text-center mb-6">
                  <h3 className="text-2xl font-bold text-white mb-2">История подписок</h3>
                  <p className="text-gray-400">Ваша история покупок и активаций подписок</p>
                </div>

                <div className="space-y-3 max-h-96 overflow-y-auto">
                  {subscriptionHistory.length > 0 ? (
                    subscriptionHistory.map((sub) => (
                      <motion.div
                        key={sub.id}
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.3 }}
                        className="bg-gradient-to-r from-gray-800/50 to-gray-700/30 rounded-lg p-4 border border-gray-600/30"
                      >
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-2">
                              <div className={`w-3 h-3 rounded-full ${
                                sub.plan === 'basic' ? 'bg-gray-500' : 
                                sub.plan === 'silver' ? 'bg-purple-500' : 'bg-yellow-500'
                              }`} />
                              <span className="text-white font-bold text-sm ml-2 capitalize">
                                {sub.plan}
                              </span>
                            </div>
                            <div>
                              <h4 className="text-lg font-bold text-white">
                                {sub.price_bkc > 0 ? `${sub.price_bkc.toLocaleString()} BKC` : `${sub.price_ton} TON`}
                              </h4>
                              <div className="text-gray-400 text-sm">
                                {sub.duration_days} дней
                              </div>
                            </div>
                          </div>
                          <div className="text-right">
                            <div className={`text-lg font-bold ${
                              sub.status === 'active' ? 'text-green-400' : 
                              sub.status === 'completed' ? 'text-gray-400' : 'text-red-400'
                            }`}>
                              {sub.status.toUpperCase()}
                            </div>
                            <div className="text-gray-400 text-xs mt-1">
                              {sub.started_at}
                            </div>
                          </div>
                        </div>
                        <div className="text-gray-400 text-sm mt-2">
                          {sub.status === 'active' && (
                            <>
                              <div>Истекает: {sub.expires_at}</div>
                              <div>Осталось дней: {sub.days_left}</div>
                            </>
                          )}
                          {sub.status === 'completed' && (
                            <div>Завершена: {sub.completed_at || 'N/A'}</div>
                          )}
                        </div>
                      </motion.div>
                    ))
                  ) : (
                    <div className="text-center py-8">
                      <Calendar className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                      <p className="text-gray-400">История подписок пуста</p>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Правая панель - Статистика */}
        <div className="space-y-6">
          {/* Статистика */}
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-purple-500/20">
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-green-400" />
              Статистика подписок
            </h3>
            
            {userStats && (
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gradient-to-br from-purple-800/50 to-pink-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-purple-400">
                      {userStats.total_active_subscriptions.toLocaleString()}
                    </div>
                    <div className="text-gray-400 text-sm">Активных подписок</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-green-800/50 to-emerald-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-green-400">
                      {(userStats.revenue_today_bkc / 1000000).toFixed(1)}M
                    </div>
                    <div className="text-gray-400 text-sm">Доход BKC</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-yellow-800/50 to-orange-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-yellow-400">
                      {userStats.new_subscriptions_today}
                    </div>
                    <div className="text-gray-400 text-sm">Новых сегодня</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-red-800/50 to-pink-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-red-400">
                      {userStats.expiring_soon_count}
                    </div>
                    <div className="text-gray-400 text-sm">Истекают скоро</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-blue-800/50 to-cyan-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-400">
                      {userStats.churn_rate.toFixed(1)}%
                    </div>
                    <div className="text-gray-400 text-sm">Отток</div>
                  </div>
                </div>
              </div>
            )}

            {/* Преимущества подписок */}
            <div className="mt-6">
              <h4 className="text-lg font-bold text-white mb-4">Преимущества подписок</h4>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-600/30">
                  <h5 className="text-yellow-400 font-bold mb-2 flex items-center gap-2">
                    <Star className="w-5 h-5" />
                    Basic
                  </h5>
                  <ul className="text-gray-300 text-sm space-y-1">
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Бесплатно
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Лимит тапов: 5,000
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Налог: 10%
                    </li>
                    <li className="flex items-center gap-2">
                      <X className="w-4 h-4 text-red-400" />
                      Комиссии на переводы
                    </li>
                    <li className="flex items-center gap-2">
                      <X className="w-4 h-4 text-red-400" />
                      Комиссии на кредиты
                    </li>
                  </ul>
                </div>
                </div>
                <div className="bg-gradient-to-br from-purple-800/50 to-purple-700/30 rounded-lg p-4 border border-purple-500/30">
                  <h5 className="text-purple-400 font-bold mb-2 flex items-center gap-2">
                    <Award className="w-5 h-5" />
                    Silver
                  </h5>
                  <ul className="text-gray-300 text-sm space-y-1">
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      50,000 BKC/мес
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Лимит тапов: 15,000
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Ранняя барахолка
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Налог: 5%
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Комиссии на переводы
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Комиссии на кредиты
                    </li>
                  </ul>
                </div>
                </div>
                <div className="bg-gradient-to-br from-yellow-500/50 to-orange-600/30 rounded-lg p-4 border border-yellow-500/30">
                  <h5 className="text-yellow-400 font-bold mb-2 flex items-center gap-2">
                    <Crown className="w-5 h-5" />
                    Gold
                  </h5>
                  <ul className="text-gray-300 text-sm space-y-1">
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      200,000 BKC/мес
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Лимит тапов: 50,000
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      График в реальном времени
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Налог: 2%
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Без комиссий на переводы
                    </li>
                    <li className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      Без комиссий на кредиты
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SubscriptionSystem;

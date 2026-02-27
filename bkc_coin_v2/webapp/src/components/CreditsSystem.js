import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { DollarSign, TrendingUp, Users, Calendar, Shield, AlertTriangle, Clock, Target, Award, Wallet, CreditCard, UserCheck, Hammer } from 'lucide-react';
import toast from 'react-hot-toast';

const CreditsSystem = ({ user, onNavigate, apiService }) => {
  const [activeTab, setActiveTab] = useState('bank');
  const [bankLoans, setBankLoans] = useState([]);
  const [p2pLoans, setP2pLoans] = useState([]);
  const [loanHistory, setLoanHistory] = useState([]);
  const [userStats, setUserStats] = useState(null);
  const [isCollectorMode, setIsCollectorMode] = useState(false);
  const [overdueLoans, setOverdueLoans] = useState([]);
  const [availableNFTs, setAvailableNFTs] = useState([]);
  const [availableBKC, setAvailableBKC] = useState(0);
  
  // Загрузка данных
  useEffect(() => {
    loadInitialData();
    const interval = setInterval(loadInitialData, 30000);
    return () => clearInterval(interval);
  }, [user.id]);

  const loadInitialData = async () => {
    try {
      const [bankData, p2pData, historyData, statsData, nftData] = await Promise.all([
        apiService.getAvailableBankLoans(),
        apiService.getP2PLoans(),
        apiService.getLoanHistory(),
        apiService.getCreditsStats(),
        apiService.getUserNFTs(),
        apiService.getUserBalance()
      ]);

      setBankLoans(bankData.loans || []);
      setP2pLoans(p2pData.loans || []);
      setLoanHistory(historyData.history || []);
      setUserStats(statsData.stats);
      setAvailableNFTs(nftData.nfts || []);
      setAvailableBKC(nftData.bkc_balance || 0);
    } catch (error) {
      console.error('Failed to load credits data:', error);
      toast.error('Ошибка загрузки данных');
    }
  };

  // Взять системный кредит
  const takeBankLoan = async (loanId) => {
    try {
      const loan = bankLoans.find(l => l.id === parseInt(loanId));
      if (!loan) return;

      const response = await apiService.takeBankLoan({
        user_id: user.id,
        loan_id: loanId,
        amount: loan.amount_min
      });

      if (response.success) {
        toast.success(`Кредит ${loan.name} успешно взят!`);
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка взятия кредита');
      }
    } catch (error) {
      toast.error('Ошибка при взятии кредита');
    }
  };

  // Создать P2P кредит
  const createP2PLoan = async () => {
    const formData = {
      borrower_id: document.getElementById('borrower-id')?.value,
      amount: parseInt(document.getElementById('p2p-amount')?.value) || 0,
      term_days: parseInt(document.getElementById('p2p-term')?.value) || 7,
      collateral_type: document.getElementById('collateral-type')?.value,
      collateral_value: parseInt(document.getElementById('collateral-value')?.value) || 0
    };

    // Валидация
    if (!formData.borrower_id || !formData.amount || formData.amount < 10000) {
      toast.error('Заполните все поля корректно');
      return;
    }

    try {
      const response = await apiService.createP2PLoan(formData);
      if (response.success) {
        toast.success('P2P кредит успешно создан!');
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка создания кредита');
      }
    } catch (error) {
      toast.error('Ошибка при создании P2P кредита');
    }
  };

  // Активировать коллекторский режим
  const startCollectorMode = async () => {
    try {
      const response = await apiService.startCollectorMode(user.id);
      if (response.success) {
        setIsCollectorMode(true);
        toast.success('🏦 Коллекторский режим активирован!');
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка активации');
      }
    } catch (error) {
      toast.error('Ошибка активации коллектора');
    }
  };

  // Собрать долг
  const collectDebt = async (loanId) => {
    try {
      const response = await apiService.collectDebt(user.id, loanId);
      if (response.success) {
        toast.success('💰 Долг успешно взыскан!');
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка взыскания');
      }
    } catch (error) {
      toast.error('Ошибка при взыскании долга');
    }
  };

  // Погасить кредит
  const repayLoan = async (loanId, amount) => {
    try {
      const response = await apiService.repayLoan(user.id, loanId, amount);
      if (response.success) {
        toast.success('💸 Кредит успешно погашен!');
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка погашения');
      }
    } catch (error) {
      toast.error('Ошибка при погашении кредита');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-violet-900 p-4">
      {/* Header */}
      <div className="bg-black/20 backdrop-blur-lg rounded-2xl p-6 mb-6 border border-purple-500/20">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2 flex items-center gap-3">
              <CreditCard className="w-8 h-8 text-purple-400" />
              Кредитная система
            </h1>
            <p className="text-gray-400">Системные и P2P кредиты под залог NFT/BKC</p>
          </div>
          <div className="text-right">
            <div className="text-gray-400 text-sm mb-1">Ваш баланс</div>
            <div className="text-2xl font-bold">
              <span className="text-yellow-400">{availableBKC.toLocaleString()}</span>
              <span className="text-gray-400 ml-2">BKC</span>
            </div>
          </div>
        </div>

      {/* Основной контент */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Левая панель - Кредиты */}
        <div className="lg:col-span-2 space-y-6">
          {/* Вкладки */}
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-purple-500/20">
            <div className="flex space-x-2 mb-6">
              <button
                onClick={() => setActiveTab('bank')}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  activeTab === 'bank' 
                    ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <DollarSign className="w-5 h-5 mr-2" />
                Системные
              </button>
              <button
                onClick={() => setActiveTab('p2p')}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  activeTab === 'p2p' 
                    ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Users className="w-5 h-5 mr-2" />
                P2P
              </button>
              <button
                onClick={() => setActiveTab('collector')}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  activeTab === 'collector' 
                    ? 'bg-gradient-to-r from-red-600 to-orange-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Hammer className="w-5 h-5 mr-2" />
                Коллектор
              </button>
            </div>

            {/* Системные кредиты */}
            {activeTab === 'bank' && (
              <div className="space-y-6">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-xl font-bold text-white">🏦 Системные кредиты</h3>
                  <div className="text-sm text-gray-400">
                    Процентная ставка: <span className="text-yellow-400 font-bold">5-7%</span> в день
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {bankLoans.map((loan) => (
                    <motion.div
                      key={loan.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.3 }}
                      className="bg-gradient-to-br from-purple-800/50 to-pink-800/30 rounded-xl p-6 border border-purple-500/30 hover:border-purple-400/50 cursor-pointer transform hover:scale-105 transition-all duration-300"
                      onClick={() => takeBankLoan(loan.id)}
                    >
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h4 className="text-lg font-bold text-white mb-2">{loan.name}</h4>
                          <div className="bg-gradient-to-r from-yellow-400 to-orange-500 text-transparent bg-clip-text text-2xl font-bold">
                            {loan.interest_rate}%
                          </div>
                          <div className="text-gray-400 text-sm mt-1">{loan.description}</div>
                        </div>
                        <div className="text-right">
                          <div className="text-3xl font-bold text-yellow-400">
                            {loan.amount_min.toLocaleString()} - {loan.amount_max.toLocaleString()}
                          </div>
                          <div className="text-gray-400 text-sm">BKC</div>
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-sm">
                        <div>Срок: {loan.term_days} дней</div>
                        <div>Мин. ставка: {loan.amount_min.toLocaleString()} BKC</div>
                      </div>
                      <button className="w-full bg-gradient-to-r from-green-500 to-emerald-600 text-white font-bold py-3 rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all duration-300">
                        Взять кредит
                      </button>
                    </motion.div>
                  ))}
                </div>
              </div>
            )}

            {/* P2P кредиты */}
            {activeTab === 'p2p' && (
              <div className="space-y-6">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-xl font-bold text-white">👥 P2P кредиты</h3>
                  <button
                    onClick={() => onNavigate('create-p2p')}
                    className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-4 py-2 rounded-lg font-semibold hover:from-purple-700 hover:to-pink-700 transition-all duration-300"
                  >
                    ➕ Создать кредит
                  </button>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {p2pLoans.length > 0 ? (
                    p2pLoans.map((loan) => (
                      <motion.div
                        key={loan.id}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.3 }}
                        className="bg-gradient-to-br from-blue-800/50 to-cyan-800/30 rounded-xl p-6 border border-blue-500/30 hover:border-blue-400/50"
                      >
                        <div className="flex justify-between items-start mb-3">
                          <div>
                            <h4 className="text-lg font-bold text-white mb-2">
                              {loan.collateral_type === 'nft' ? '🎨' : '💰'} Залог: {loan.collateral_value.toLocaleString()} BKC
                            </h4>
                            <div className="text-gray-400 text-sm">
                              Заемщик: #{loan.borrower_id} | Срок: {loan.term_days} дней
                            </div>
                          </div>
                          <div className="text-right">
                            <div className={`text-2xl font-bold ${
                              loan.status === 'active' ? 'text-green-400' : 'text-red-400'
                            }`}>
                              {loan.status.toUpperCase()}
                            </div>
                          </div>
                        </div>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>Сумма: {loan.principal.toLocaleString()} BKC</div>
                          <div>Ставка: 3.0%/день</div>
                        </div>
                        {loan.status === 'active' && (
                          <button
                            onClick={() => repayLoan(loan.id, loan.total_due)}
                            className="w-full bg-gradient-to-r from-green-500 to-emerald-600 text-white font-bold py-2 rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all duration-300"
                          >
                            💸 Погасить
                          </button>
                        )}
                      </motion.div>
                    ))
                  ) : (
                    <div className="text-center py-12">
                      <div className="text-gray-400 mb-4">
                        <Users className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                      </div>
                      <p className="text-gray-400">P2P кредиты не найдены</p>
                      <button
                        onClick={() => onNavigate('create-p2p')}
                        className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-6 py-3 rounded-lg font-semibold hover:from-purple-700 hover:to-pink-700 transition-all duration-300"
                      >
                        ➕ Создать первый P2P кредит
                      </button>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Коллекторский режим */}
            {activeTab === 'collector' && (
              <div className="space-y-6">
                <div className="text-center mb-6">
                  <motion.div
                    animate={{ 
                      scale: [1, 1.05, 1], 
                      rotate: [0, 1, -1, 0] 
                    }}
                    transition={{ duration: 0.5, repeat: Infinity }}
                    className="inline-block"
                  >
                    <Hammer className="w-16 h-16 text-red-500 mb-4" />
                  </motion.div>
                  <h3 className="text-2xl font-bold text-white mb-2">🏦 Коллекторский режим</h3>
                  <p className="text-gray-400 mb-4">
                    Собирайте просроченные кредиты и получайте <span className="text-yellow-400 font-bold">26%</span> от суммы долга
                  </p>
                </div>

                {overdueLoans.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {overdueLoans.map((loan) => (
                      <motion.div
                        key={loan.id}
                        initial={{ opacity: 0, x: -50 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.5 }}
                        className="bg-gradient-to-br from-red-800/50 to-orange-800/30 rounded-xl p-6 border border-red-500/50 hover:border-red-400/50"
                      >
                        <div className="flex justify-between items-start mb-3">
                          <div>
                            <h4 className="text-lg font-bold text-white mb-2">
                              <AlertTriangle className="w-5 h-5 text-red-400 mr-2" />
                              Должник: #{loan.borrower_id}
                            </h4>
                            <div className="text-gray-400 text-sm">
                              Просрочка: {loan.days_overdue} дней
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-2xl font-bold text-red-400">
                              {loan.total_due.toLocaleString()} BKC
                            </div>
                          </div>
                        </div>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>Долг: {loan.principal.toLocaleString()} BKC</div>
                          <div>Комиссия: 26%</div>
                        </div>
                        <button
                          onClick={() => collectDebt(loan.id)}
                          className="w-full bg-gradient-to-r from-red-600 to-orange-600 text-white font-bold py-3 rounded-lg hover:from-red-700 hover:to-orange-700 transition-all duration-300"
                        >
                          🔪 Взыскать долг
                        </button>
                      </motion.div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-12">
                    <Shield className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                    <p className="text-gray-400">Просроченных кредитов не найдено</p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Правая панель - Статистика и история */}
        <div className="space-y-6">
          {/* Статистика */}
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-purple-500/20">
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-green-400" />
              Статистика
            </h3>
            
            {userStats && (
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gradient-to-br from-purple-800/50 to-pink-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-yellow-400">
                      {userStats.total_loans_today.toLocaleString()}
                    </div>
                    <div className="text-gray-400 text-sm">Кредитов сегодня</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-green-800/50 to-emerald-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-green-400">
                      {(userStats.total_loaned_today / 1000000).toFixed(1)}M
                    </div>
                    <div className="text-gray-400 text-sm">Выдано BKC</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-yellow-800/50 to-orange-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-yellow-400">
                      {userStats.total_interest_today.toLocaleString()}
                    </div>
                    <div className="text-gray-400 text-sm">Доход от %</div>
                  </div>
                </div>
                <div className="bg-gradient-to-br from-red-800/50 to-pink-800/30 rounded-lg p-4">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-red-400">
                      {userStats.active_collectors}
                    </div>
                    <div className="text-gray-400 text-sm">Активных коллекторов</div>
                  </div>
                </div>
              </div>
            )}

            {/* История кредитов */}
            <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-purple-500/20">
              <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                <Clock className="w-5 h-5 text-blue-400" />
                История кредитов
              </h3>
              
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {loanHistory.length > 0 ? (
                  loanHistory.map((loan) => (
                    <motion.div
                      key={loan.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.3 }}
                      className="bg-gradient-to-r from-gray-800/50 to-gray-700/30 rounded-lg p-4 border border-gray-600/30 hover:border-gray-500/50"
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-2">
                            <div className={`w-3 h-3 rounded-full ${
                              loan.type === 'bank' ? 'bg-purple-500' : 'bg-blue-500'
                            }`} />
                            <span className="text-white font-bold text-sm ml-2">
                              {loan.type === 'bank' ? 'БАНК' : 'P2P'}
                            </span>
                          </div>
                          <div>
                            <h4 className="text-lg font-bold text-white">
                              {loan.principal.toLocaleString()} BKC
                            </h4>
                            <div className="text-gray-400 text-sm">
                              {loan.interest_rate}%/день • {loan.term_days} дней
                            </div>
                          </div>
                        </div>
                        <div className="text-right">
                          <div className={`text-lg font-bold ${
                            loan.status === 'completed' ? 'text-green-400' : 'text-yellow-400'
                          }`}>
                            {loan.status.toUpperCase()}
                          </div>
                          <div className="text-gray-400 text-xs mt-1">
                            {loan.created_at}
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  ))
                ) : (
                  <div className="text-center py-8">
                    <Calendar className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                    <p className="text-gray-400">История кредитов пуста</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreditsSystem;

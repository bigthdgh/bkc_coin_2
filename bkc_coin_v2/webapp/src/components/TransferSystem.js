import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Send, Wallet, Users, TrendingUp, AlertTriangle, CheckCircle, X, ArrowUpDown, DollarSign, Clock, Shield } from 'lucide-react';
import toast from 'react-hot-toast';

const TransferSystem = ({ user, onNavigate, apiService }) => {
  const [activeTab, setActiveTab] = useState('send');
  const [transferType, setTransferType] = useState('bkc');
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [transferHistory, setTransferHistory] = useState([]);
  const [userBalance, setUserBalance] = useState(0);
  const [availableNFTs, setAvailableNFTs] = useState([]);
  const [isProcessing, setIsProcessing] = useState(false);
  const [showNFTSelect, setShowNFTSelect] = useState(false);
  const [selectedNFT, setSelectedNFT] = useState(null);
  
  // Загрузка данных
  useEffect(() => {
    loadInitialData();
    const interval = setInterval(loadInitialData, 30000);
    return () => clearInterval(interval);
  }, [user.id]);

  const loadInitialData = async () => {
    try {
      const [historyData, balanceData, nftData] = await Promise.all([
        apiService.getTransferHistory(user.id),
        apiService.getUserBalance(user.id),
        apiService.getUserNFTs(user.id)
      ]);

      setTransferHistory(historyData.history || []);
      setUserBalance(balanceData.balance || 0);
      setAvailableNFTs(nftData.nfts || []);
    } catch (error) {
      console.error('Failed to load transfer data:', error);
      toast.error('Ошибка загрузки данных');
    }
  };

  // Отправить перевод
  const sendTransfer = async () => {
    if (isProcessing) return;
    
    // Валидация
    if (!recipient || !amount || parseFloat(amount) <= 0) {
      toast.error('Заполните все поля корректно');
      return;
    }

    const transferAmount = parseFloat(amount);
    
    // Проверка баланса
    if (transferType === 'bkc' && transferAmount > userBalance) {
      toast.error('Недостаточно BKC на балансе');
      return;
    }

    // Проверка NFT для перевода
    if (transferType === 'nft' && !selectedNFT) {
      toast.error('Выберите NFT для перевода');
      return;
    }

    setIsProcessing(true);
    
    try {
      const transferData = {
        sender_id: user.id,
        recipient: recipient,
        amount: transferAmount,
        type: transferType,
        nft_id: selectedNFT?.id
      };

      const response = await apiService.sendTransfer(transferData);
      
      if (response.success) {
        toast.success('Перевод успешно отправлен!');
        setRecipient('');
        setAmount('');
        setSelectedNFT(null);
        setShowNFTSelect(false);
        loadInitialData();
      } else {
        toast.error(response.error || 'Ошибка перевода');
      }
    } catch (error) {
      toast.error('Ошибка при отправке перевода');
    } finally {
      setIsProcessing(false);
    }
  };

  // Выбрать NFT для перевода
  const selectNFT = (nft) => {
    setSelectedNFT(nft);
    setShowNFTSelect(false);
  };

  const getCommissionInfo = (amount, type) => {
    if (type === 'bkc') {
      return {
        hasCommission: false,
        commission: 0,
        totalAmount: amount
      };
    }
    
    // Для NFT и BKC переводов комиссия 5%
    const commission = amount * 0.05;
    return {
      hasCommission: true,
      commission: commission,
      totalAmount: amount + commission
    };
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-gray-900 to-black p-4">
      {/* Header */}
      <div className="bg-black/20 backdrop-blur-lg rounded-2xl p-6 mb-6 border border-gray-700/20">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2 flex items-center gap-3">
              <Send className="w-8 h-8 text-blue-400" />
              Переводы
            </h1>
            <p className="text-gray-400">Мгновенные переводы BKC и NFT между пользователями</p>
          </div>
          <div className="text-right">
            <div className="text-gray-400 text-sm mb-1">Ваш баланс</div>
            <div className="text-2xl font-bold">
              <span className="text-blue-400">{userBalance.toLocaleString()}</span>
              <span className="text-gray-400 ml-2">BKC</span>
            </div>
          </div>
        </div>
      </div>

      {/* Основной контент */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Левая панель - Форма перевода */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-gray-700/20">
            <h3 className="text-xl font-bold text-white mb-6">Создать перевод</h3>
            
            {/* Тип перевода */}
            <div className="flex space-x-2 mb-6">
              <button
                onClick={() => { setTransferType('bkc'); setShowNFTSelect(false); }}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  transferType === 'bkc' 
                    ? 'bg-gradient-to-r from-blue-600 to-cyan-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <DollarSign className="w-5 h-5 mr-2" />
                BKC
              </button>
              <button
                onClick={() => { setTransferType('nft'); setShowNFTSelect(true); }}
                className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                  transferType === 'nft' 
                    ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Wallet className="w-5 h-5 mr-2" />
                NFT
              </button>
            </div>

            {/* Форма перевода */}
            <div className="space-y-4">
              {/* Получатель */}
              <div>
                <label className="block text-gray-300 text-sm mb-2">Получатель</label>
                <input
                  type="text"
                  value={recipient}
                  onChange={(e) => setRecipient(e.target.value)}
                  placeholder="ID пользователя или @username"
                  className="w-full bg-gray-800 text-white rounded-lg p-4 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20"
                />
              </div>

              {/* Сумма */}
              <div>
                <label className="block text-gray-300 text-sm mb-2">
                  {transferType === 'bkc' ? 'Сумма BKC' : 'Выберите NFT'}
                </label>
                {transferType === 'bkc' ? (
                  <input
                    type="number"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="1000"
                    min="1"
                    step="0.1"
                    className="w-full bg-gray-800 text-white rounded-lg p-4 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20"
                  />
                ) : (
                  <div className="bg-gray-800 rounded-lg p-4 border border-gray-600/30">
                    <button
                      onClick={() => setShowNFTSelect(true)}
                      className="w-full bg-gray-700 text-gray-300 p-3 rounded-lg hover:bg-gray-600 transition-all duration-300"
                    >
                      {selectedNFT ? (
                        <div className="flex items-center justify-between">
                          <span>{selectedNFT.name}</span>
                          <X className="w-5 h-5 text-gray-400" />
                        </div>
                      ) : (
                        <div className="text-center text-gray-400">
                          Выберите NFT для перевода
                        </div>
                      )}
                    </button>
                  </div>
                )}
              </div>

              {/* Информация о комиссии */}
              {transferType === 'bkc' && amount && (
                <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-600/30">
                  {(() => {
                    const commissionInfo = getCommissionInfo(parseFloat(amount), transferType);
                    return (
                      <div className="space-y-2">
                        <div className="flex justify-between items-center">
                          <span className="text-gray-400">Комиссия:</span>
                          <span className={`font-bold ${
                            commissionInfo.hasCommission ? 'text-yellow-400' : 'text-green-400'
                          }`}>
                            {commissionInfo.hasCommission ? `${commissionInfo.commission.toLocaleString()} BKC` : 'Без комиссии'}
                          </span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-gray-400">Итого:</span>
                          <span className="text-xl font-bold text-white">
                            {commissionInfo.totalAmount.toLocaleString()} BKC
                          </span>
                        </div>
                      </div>
                      </div>
                    );
                  })()}
                </div>
              )}

              {/* Кнопка отправки */}
              <button
                onClick={sendTransfer}
                disabled={isProcessing}
                className={`w-full bg-gradient-to-r from-green-500 to-emerald-600 text-white font-bold py-4 rounded-lg transition-all duration-300 ${
                  isProcessing ? 'opacity-50 cursor-not-allowed' : 'hover:from-green-600 hover:to-emerald-700 transform hover:scale-105'
                }`}
              >
                {isProcessing ? (
                  <>
                    <div className="inline-block animate-spin rounded-full h-5 w-5 border-b-2 border-white border-t-transparent"></div>
                    Отправка...
                  </>
                ) : (
                  <>
                    <Send className="w-5 h-5 mr-2" />
                    Отправить перевод
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Правая панель - История и NFT */}
        <div className="space-y-6">
          {/* Доступные NFT */}
          {showNFTSelect && (
            <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-gray-700/20">
              <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                <Wallet className="w-5 h-5 text-purple-400" />
                Ваши NFT
              </h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {availableNFTs.length > 0 ? (
                  availableNFTs.map((nft) => (
                    <motion.div
                      key={nft.id}
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                      whileHover={{ scale: 1.05 }}
                      transition={{ duration: 0.3 }}
                      onClick={() => selectNFT(nft)}
                      className="bg-gradient-to-br from-purple-800/50 to-pink-800/30 rounded-lg p-4 border border-purple-500/30 cursor-pointer transform hover:scale-105 transition-all duration-300"
                    >
                      <div className="text-center">
                        <div className="text-2xl font-bold text-white mb-2">{nft.name}</div>
                        <div className="text-gray-400 text-sm">ID: #{nft.id}</div>
                        <div className="text-gray-400 text-sm">Редкость: {nft.rarity}</div>
                      </div>
                    </motion.div>
                  ))
                ) : (
                  <div className="text-center py-8">
                    <div className="text-gray-400 mb-4">
                      <AlertTriangle className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                    </div>
                    <p className="text-gray-400">У вас нет доступных NFT для перевода</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* История переводов */}
          <div className="bg-black/30 backdrop-blur-lg rounded-2xl p-6 border border-gray-700/20">
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <Clock className="w-5 h-5 text-blue-400" />
              История переводов
            </h3>
            
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {transferHistory.length > 0 ? (
                transferHistory.map((transfer) => (
                  <motion.div
                    key={transfer.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.3 }}
                    className="bg-gradient-to-r from-gray-800/50 to-gray-700/30 rounded-lg p-4 border border-gray-600/30"
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <div className={`w-3 h-3 rounded-full ${
                            transfer.type === 'bkc' ? 'bg-blue-500' : 'bg-purple-500'
                          }`} />
                          <span className="text-white font-bold text-sm ml-2">
                            {transfer.type === 'bkc' ? 'BKC' : 'NFT'}
                          </span>
                        </div>
                        <div>
                          <div className="text-lg font-bold text-white">
                            {transfer.type === 'bkc' ? `${transfer.amount.toLocaleString()} BKC` : transfer.nft_name}
                          </div>
                          <div className="text-gray-400 text-sm">
                            {transfer.type === 'bkc' ? (
                              <>От: {user.username} → {transfer.recipient}</>
                            ) : (
                              <>NFT #{transfer.nft_id}</>
                            )}
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className={`text-lg font-bold ${
                          transfer.status === 'completed' ? 'text-green-400' : 'text-yellow-400'
                        }`}>
                          {transfer.status.toUpperCase()}
                        </div>
                        <div className="text-gray-400 text-xs mt-1">
                          {transfer.created_at}
                        </div>
                      </div>
                    </div>
                    {transfer.commission > 0 && (
                      <div className="text-gray-400 text-sm mt-2">
                        Комиссия: {transfer.commission.toLocaleString()} BKC
                      </div>
                    )}
                  </motion.div>
                ))
              ) : (
                <div className="text-center py-8">
                  <div className="text-gray-400 mb-4">
                    <Users className="w-16 h-16 mx-auto mb-4 text-gray-500" />
                  </div>
                  <p className="text-gray-400">История переводов пуста</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Модальное окно выбора NFT */}
      <AnimatePresence>
        {showNFTSelect && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
            onClick={() => setShowNFTSelect(false)}
          >
            <motion.div
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.8, opacity: 0 }}
              className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-2xl p-8 max-w-4xl mx-4 border border-gray-600/50"
            >
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-xl font-bold text-white">Выберите NFT для перевода</h3>
                <button
                  onClick={() => setShowNFTSelect(false)}
                  className="text-gray-400 hover:text-gray-200"
                >
                  <X className="w-6 h-6" />
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-96 overflow-y-auto">
                {availableNFTs.map((nft) => (
                  <motion.div
                    key={nft.id}
                    whileHover={{ scale: 1.02 }}
                    onClick={() => selectNFT(nft)}
                    className="bg-gradient-to-br from-purple-800/50 to-pink-800/30 rounded-lg p-4 border border-purple-500/30 cursor-pointer transform hover:scale-105 transition-all duration-300"
                  >
                    <div className="text-center">
                      <div className="text-2xl font-bold text-white mb-2">{nft.name}</div>
                      <div className="text-gray-400 text-sm">ID: #{nft.id}</div>
                      <div className="text-gray-400 text-sm">Редкость: {nft.rarity}</div>
                      <div className="text-gray-400 text-xs mt-2">{nft.description}</div>
                    </div>
                  </motion.div>
                ))}
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default TransferSystem;

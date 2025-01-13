import React, { useState } from 'react';
import MainPage from './MainPage';

const LandingPage = () => {
  const [showRecordingPage, setShowRecordingPage] = useState(false);
  const [email, setEmail] = useState('');

  if (showRecordingPage) {
    return <MainPage email={email} />;
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-purple-500">
          Meet<span className="text-orange-400">Buddy</span> <span role="img" aria-label="laptop">👩‍💻</span>
        </h1>
        <p className="text-gray-400">Your personal remote meetings assistant.</p>
      </div>

      {/* Email Input Section */}
      <div className="flex flex-col items-center space-y-4">
        <p className="text-lg font-medium text-white">Enter your <span className="text-orange-400">email</span> to get started</p>
        <div className="flex items-center space-x-2">
          <input
            type="email"
            placeholder="yourname@example.com"
            className="px-4 py-2 rounded-lg bg-gray-800 text-gray-300 focus:outline-none focus:ring-2 focus:ring-purple-500"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <button
            className="bg-orange-400 hover:bg-orange-500 text-white font-semibold py-2 px-4 rounded-lg"
            onClick={() => setShowRecordingPage(true)}
          >
            Submit
          </button>
        </div>
      </div>

      {/* Features Section */}
      <div className="text-left mt-12">
        <h2 className="text-xl font-bold mb-4">I can:</h2>
        <ul className="space-y-2">
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">➤</span>
            <span className="text-gray-300">Watch and listen to your meetings</span>
          </li>
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">✏️</span>
            <span className="text-gray-300">Transcribe voices and scan screen shares</span>
          </li>
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">🔎️</span>
            <span className="text-gray-300">Gather and analyze meeting data</span>
          </li>
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">📚</span>
            <span className="text-gray-300">Create a clear summary of key points</span>
          </li>
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">✨</span>
            <span className="text-gray-300">Generate an easy-to-read report</span>
          </li>
          <li className="flex items-center">
            <span className="text-purple-500 mr-2">📧</span>
            <span className="text-gray-300">Email it to you or make it available to download</span>
          </li>
        </ul>
      </div>
    </div>
  );
};

export default LandingPage;

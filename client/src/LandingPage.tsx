import { useState } from "react";
import MainPage from "./MainPage";

const LandingPage = () => {
  const [showRecordingPage, setShowRecordingPage] = useState(false);
  const [email, setEmail] = useState("");

  if (showRecordingPage) {
    return <MainPage email={email} />;
  }

  return (
    <div className="min-h-screen bg-neutral-800 text-white flex flex-col items-center justify-center">
      {/* Title Section */}
      <div className="text-center mb-20">
        <h1 className="text-6xl font-bold text-purple-500">
          Meet<span className="text-amber-400">Buddy</span>{" "}
          <span role="img" aria-label="laptop">
            👨‍💻
          </span>
        </h1>
        <p className="text-neutral-700 mt-4 text-xl font-bold">Your personal remote meetings assistant.</p>
      </div>

      {/* Two-Column Layout */}
      <div className="grid grid-cols-1 md:grid-cols-2 max-w-6xl w-full px-6">
        {/* Left Column - Email Input Section */}
        <div className="flex flex-col items-center justify-center space-y-6 border-r-2 border-neutral-700">
          <p className="text-xl font-medium text-white text-center">
            Enter your <span className="text-amber-400">email</span> to get started:
          </p>
          <div className="flex items-center space-x-2 text-lg">
            <input
              type="email"
              placeholder="jurson.ziomal@żabol.com"
              className="px-4 py-2 rounded-lg bg-neutral-900 text-neutral-300 focus:outline-none focus:ring-2 focus:ring-purple-500 w-full mr-4"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <button
              className="bg-purple-500 hover:text-amber-400 text-white font-semibold py-2 px-4 rounded-lg"
              onClick={() => setShowRecordingPage(true)}
            >
              Submit
            </button>
          </div>
        </div>

        {/* Right Column - Features Section */}
        <div className="text-left flex flex-col items-center justify-center">
          <h2 className="text-xl font-bold mb-4 text-center">I can:</h2>
          <ul className="space-y-2">
            <li className="flex items-center">
              <span className="text-purple-500 mr-3">➤</span>
              <span className="text-gray-300">
                <span className="text-purple-400">Watch</span> and
                <span className="text-purple-400"> listen</span> to your meetings
              </span>
              <span className="text-neutral-600 font-bold ml-2">&lt;</span>
            </li>
            <li className="flex items-center">
              <span className="text-neutral-600 font-bold ml-2">&gt;</span>
              <span className="text-neutral-300 ml-3">
                <span className="text-purple-400">Transcribe</span> voices and
                <span className="text-purple-400"> scan</span> screen shares
              </span>
              <span className="text-purple-400 ml-2">✏️</span>
            </li>
            <li className="flex items-center">
              <span className="text-purple-400 mr-2">🔎️</span>
              <span className="text-neutral-300">
                <span className="text-purple-400">Gather</span> and
                <span className="text-purple-400"> analyze</span> meeting data
              </span>
              <span className="text-neutral-600 font-bold ml-2">&lt;</span>
            </li>
            <li className="flex items-center">
              <span className="text-neutral-600 font-bold ml-2">&gt;</span>
              <span className="text-gray-300 ml-3">
                Create a<span className="text-purple-400"> clear summary</span> of key points
              </span>
              <span className="text-purple-400 ml-2">📚</span>
            </li>
            <li className="flex items-center">
              <span className="text-purple-400 mr-2">✨</span>
              <span className="text-gray-300">
                Generate an
                <span className="text-purple-400"> easy-to-read report</span>
              </span>
              <span className="text-neutral-700">&lt;</span>
            </li>
            <li className="flex items-center">
              <span className="text-neutral-600 font-bold ml-2">&gt;</span>
              <span className="text-gray-300 ml-3">
                <span className="text-purple-400">Email</span> it to you or make it available to download
              </span>
              <span className="text-purple-400 ml-2">📧</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;

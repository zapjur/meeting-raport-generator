import React, { useState } from "react";
import MediaCapture from "./MediaCapture";

const MainPage: React.FC = () => {
  const [isRecording, setIsRecording] = useState(false);

  const toggleRecording = () => {
    setIsRecording(!isRecording);
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-purple-500">
          Meet<span className="text-orange-400">Buddy</span> <span role="img" aria-label="laptop">👩‍💻</span>
        </h1>
        <p className="text-gray-400">Your personal remote meetings assistant.</p>
      </div>

      {/* Recording Section */}
      <div className="bg-gray-800 p-8 rounded-lg shadow-lg flex flex-col items-center mb-12">
        <button
          onClick={toggleRecording}
          className={`${
            isRecording ? "bg-green-500 hover:bg-green-600" : "bg-red-500 hover:bg-red-600"
          } text-white font-semibold py-3 px-6 rounded-lg mb-6 transition duration-300 ease-in-out`}
        >
          {isRecording ? "Stop capturing" : "Start capturing now!"}
        </button>

        <MediaCapture isRecording={isRecording} />

        <p className="text-gray-400 mb-4">or</p>

        <div className="flex items-center space-x-4 mb-6">
          <p className="text-white text-lg">Start capturing at:</p>
          <span className="bg-orange-400 text-white font-semibold py-2 px-4 rounded-lg text-xl">
            14:45
          </span>
          <button className="bg-orange-400 hover:bg-orange-500 text-white font-semibold py-2 px-4 rounded-lg text-lg transition duration-300 ease-in-out">
            Submit
          </button>
        </div>
      </div>

      {/* Features Checklist */}
      <div className="flex flex-col items-start text-left space-y-2 max-w-md w-full">
        {[
          "Scan screen share",
          "Transcribe Voices",
          "Create AI Summary",
          "Create meeting report",
          "Email it to you",
          "Make available for download",
        ].map((feature, index) => (
          <div key={index} className="flex items-center space-x-3">
            <input
              type="checkbox"
              checked={index !== 1 && index !== 5} // Checked for some items
              className="form-checkbox h-5 w-5 text-purple-500 border-gray-700 rounded"
              readOnly
            />
            <span className="text-gray-300 text-lg">{feature}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default MainPage;

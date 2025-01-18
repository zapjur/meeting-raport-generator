import React, { useState } from "react";
import MediaCapture from "./MediaCapture";

const MainPage: React.FC = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [meetingId, setMeetingId] = useState<string | null>(null);

  const toggleRecording = async () => {
    setIsRecording(!isRecording);

    // Fetch the meeting ID when starting the recording
    if (!isRecording) {
      try {
        const response = await fetch("http://127.0.0.1:8080/generate-meeting-id");
        const data = await response.json();
        if (data?.meeting_id) {
          setMeetingId(data.meeting_id);
        } else {
          throw new Error("Invalid response format");
        }
      } catch (error) {
        console.error("Error fetching meeting ID:", error);
        setMeetingId(null); // Reset meeting ID on error
      }
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-purple-500">
          Meet<span className="text-orange-400">Buddy</span>{" "}
          <span role="img" aria-label="laptop">
            👩‍💻
          </span>
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

        {meetingId && (
          <p className="mt-4 text-lg text-green-400">
            Your Meeting ID: <span className="font-bold">{meetingId}</span>
          </p>
        )}
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

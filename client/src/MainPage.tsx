import React, { useState } from "react";
import MediaCapture from "./MediaCapture";

interface MainPageProps {
  email: string;
}

const MainPage: React.FC<MainPageProps> = ({ email }) => {
  const [isRecording, setIsRecording] = useState(false);
  const [meetingId, setMeetingId] = useState<string | null>(null);

  const toggleRecording = async () => {
    if (isRecording) {
      // Stop recording and send the meeting ID to the endpoint
      if (meetingId) {
        try {
          const response = await fetch("http://127.0.0.1:8080/end-meeting", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ meeting_id: meetingId }),
          });

          if (!response.ok) {
            throw new Error(`Failed to end meeting: ${response.statusText}`);
          }

          const data = await response.json();
          console.log(data.message); // Log the response message
        } catch (error) {
          console.error("Error ending the meeting:", error);
        }
      }

      setMeetingId(null); // Clear the meeting ID after ending
    } else {
      // Start recording and fetch a new meeting ID
      try {
        const response = await fetch(`http://127.0.0.1:8080/generate-meeting-id?email=${email}`);
        const data = await response.json();
        if (data?.meeting_id) {
          setMeetingId(data.meeting_id);
        } else {
          throw new Error("Invalid response format");
        }
      } catch (error) {
        console.error("Error fetching meeting ID:", error);
        setMeetingId(null);
      }
    }
    setIsRecording(!isRecording);
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

        <MediaCapture isRecording={isRecording} meetingId={meetingId} />

        {meetingId && (
          <p className="mt-4 text-lg text-green-400">
            Your Meeting ID: <span className="font-bold">{meetingId}</span>
          </p>
        )}
      </div>
    </div>
  );
};

export default MainPage;

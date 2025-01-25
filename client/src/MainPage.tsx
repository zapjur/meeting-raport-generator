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
          console.log(data.message);
        } catch (error) {
          console.error("Error ending the meeting:", error);
        }
      }

      setMeetingId(null);
    } else {
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
    <div className="min-h-screen bg-neutral-800 text-white flex flex-col items-center justify-center">
      {/* Header */}
      <div className="text-center mb-20">
        <h1 className="text-6xl font-bold text-purple-500">
          Meet<span className="text-amber-400">Buddy</span>{" "}
          <span role="img" aria-label="laptop">
            👨‍💻
          </span>
        </h1>
        <p className="text-neutral-700 mt-4 text-xl font-bold">Your personal remote meetings assistant.</p>
      </div>

      {/* Recording Section */}
      <div className="bg-neutral-900 p-8 rounded-lg shadow-lg flex flex-col items-center mb-12">
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

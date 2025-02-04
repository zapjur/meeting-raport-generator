import React, { useState, useEffect } from "react";
import MediaCapture from "./MediaCapture";

interface MainPageProps {
  email: string;
}

const MainPage: React.FC<MainPageProps> = ({ email }) => {
  // Set initial time values based on current time
  const currentTime = new Date();
  const [isRecording, setIsRecording] = useState(false);
  const [meetingId, setMeetingId] = useState<string | null>(null);
  const [selectedHour, setSelectedHour] = useState<number>(currentTime.getHours());
  const [selectedMinute, setSelectedMinute] = useState<number>(currentTime.getMinutes());
  const [timeToStart, setTimeToStart] = useState<number | null>(null);
  const [isTimeSet, setIsTimeSet] = useState(false); // Track if the time has been set

  useEffect(() => {
    // Check if the exact time has passed, if so, allow recording to start
    const interval = setInterval(() => {
      const currentTime = new Date();
      const targetTime = new Date(currentTime);
      targetTime.setHours(selectedHour, selectedMinute, 0, 0); // Set target time

      if (
        timeToStart !== null &&
        currentTime >= targetTime &&
        currentTime <= new Date(targetTime.getTime() + 60000) &&
        !isRecording &&
        isTimeSet
      ) {
        // Automatically start recording when the time matches exactly
        toggleRecording();
      }
    }, 1000); // Check every second

    return () => clearInterval(interval); // Clean up the interval on unmount
  }, [selectedHour, selectedMinute, timeToStart, isRecording, isTimeSet]);

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
      setIsTimeSet(false); // Reset time set flag when stopping the recording
    } else {
      // Start waiting for the selected time
      setTimeToStart(new Date().getTime());
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

  const setTimeForRecording = () => {
    // This will only set the time when "Set" is clicked
    setTimeToStart(new Date().getTime());
    setIsTimeSet(true); // Mark the time as set
  };

  return (
    <div className="min-h-screen bg-neutral-800 text-white flex flex-col items-center justify-center">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-6xl font-bold text-purple-500">
          Meet<span className="text-amber-400">Buddy</span>{" "}
          <span role="img" aria-label="laptop">
            👨‍💻
          </span>
        </h1>
        <p className="text-neutral-700 mt-4 text-xl font-bold">Your personal remote meetings assistant.</p>
      </div>

      {/* Recording Section (Two-Column Layout) */}
      <div className="flex items-center justify-center space-x-6 mb-12">
        <div className="bg-neutral-900 p-8 rounded-lg shadow-lg flex flex-col items-center">
          <button
            onClick={toggleRecording}
            className={`${
              isRecording ? "bg-green-500 hover:bg-green-600" : "bg-red-500 hover:bg-red-600"
            } text-white font-semibold py-3 px-6 rounded-lg mb-6 transition duration-300 ease-in-out`}
          >
            {isRecording ? "Stop capturing" : "Start capturing now!"}
          </button>

          {/* "or" text */}
          <p className="text-center text-neutral-400 text-2xl">or</p>

          {/* Time Input for hour and minute */}
          <div className="mt-4 flex flex-col items-center space-y-2">
            <p className="text-neutral-600 text-xs mb-1">Start recording at:</p>
            <div className="flex items-center space-x-4">
              {/* Hour Input */}
              <div className="flex flex-col items-center">
                <input
                  type="number"
                  value={selectedHour}
                  onChange={(e) => setSelectedHour(parseInt(e.target.value))}
                  className="py-2 pl-4 rounded-lg text-white bg-neutral-800 text-center"
                  min="0"
                  max="23"
                  placeholder="00"
                />
              </div>

              {/* Colon between hour and minute */}
              <span className="text-neutral-400 text-lg">:</span>

              {/* Minute Input */}
              <div className="flex flex-col items-center">
                <input
                  type="number"
                  value={selectedMinute}
                  onChange={(e) => setSelectedMinute(parseInt(e.target.value))}
                  className="py-2 pl-4 rounded-lg text-white bg-neutral-800 text-center"
                  min="0"
                  max="59"
                  placeholder="00"
                />
              </div>

              {/* Set Button */}
              <button
                onClick={setTimeForRecording}
                className={`${
                  isTimeSet ? "bg-neutral-800 text-white" : "bg-blue-500 text-white"
                } font-semibold py-2 px-4 rounded-lg`}
              >
                Set
              </button>
            </div>
          </div>

          <MediaCapture isRecording={isRecording} meetingId={meetingId} />

          {meetingId && (
            <p className="mt-4 text-lg text-green-400">
              Your Meeting ID: <span className="font-bold">{meetingId}</span>
            </p>
          )}
        </div>

        {/* Arrow Between Columns */}
        <div className="text-white text-4xl">➡</div>

        <img src="/record_instruction.png" alt="Recording instructions" className="w-1/2 object-contain" />
      </div>
    </div>
  );
};

export default MainPage;

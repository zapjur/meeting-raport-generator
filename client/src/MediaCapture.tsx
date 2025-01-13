import React, { useState, useRef } from 'react';

// Declare ImageCapture type if not available
declare class ImageCapture {
  constructor(track: MediaStreamTrack);
  grabFrame(): Promise<ImageBitmap>;
}

const MediaCapture: React.FC = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [screenshotInterval, setScreenshotInterval] = useState<NodeJS.Timeout | null>(null);
  const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const previousFrameRef = useRef<ImageData | null>(null);

  const startRecording = async () => {
    console.log("Starting screen and audio capture...");
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
      console.log("Screen and audio stream obtained", stream);

      // Start capturing screenshots
      const interval = setInterval(() => captureScreenshot(stream), 1000);
      setScreenshotInterval(interval);

      // Setup audio recording
      const audioRecorder = setupAudioRecorder(stream);
      audioRecorder.start();
      setMediaRecorder(audioRecorder);

      setIsRecording(true);
    } catch (err) {
      console.error("Error starting screen/audio capture:", err);
    }
  };

  const stopRecording = () => {
    console.log("Stopping capture...");
    if (screenshotInterval) clearInterval(screenshotInterval);

    if (mediaRecorder) {
      mediaRecorder.stop();
      setMediaRecorder(null);
    }

    setIsRecording(false);
  };

  const setupAudioRecorder = (stream: MediaStream): MediaRecorder => {
    const audioChunks: Blob[] = [];
    const recorder = new MediaRecorder(stream);

    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        audioChunks.push(event.data);
      }
    };

    recorder.onstop = async () => {
      const audioBlob = new Blob(audioChunks, { type: "audio/webm" });
      const formData = new FormData();
      formData.append("audio", audioBlob, `audio-${Date.now()}.webm`);

      try {
        const response = await fetch("http://localhost:8080/capture-audio", {
          method: "POST",
          body: formData,
        });

        const result = await response.json();
        console.log("Audio upload response:", result);
      } catch (err) {
        console.error("Error sending audio to server:", err);
      }
    };

    return recorder;
  };

  const captureScreenshot = async (stream: MediaStream) => {
    if (canvasRef.current) {
      const canvas = canvasRef.current;
      const videoTrack = stream.getVideoTracks()[0];
      const imageCapture = new ImageCapture(videoTrack);

      try {
        const imageBitmap = await imageCapture.grabFrame();
        const ctx = canvas.getContext('2d');

        if (ctx) {
          canvas.width = imageBitmap.width;
          canvas.height = imageBitmap.height;
          ctx.drawImage(imageBitmap, 0, 0);

          const currentFrameData = ctx.getImageData(0, 0, canvas.width, canvas.height);
          if (!previousFrameRef.current || compareFrames(previousFrameRef.current, currentFrameData)) {
            previousFrameRef.current = currentFrameData;

            canvas.toBlob(async (blob) => {
              if (blob) {
                const formData = new FormData();
                formData.append("screenshot", blob, `screenshot-${Date.now()}.png`);

                try {
                  const response = await fetch("http://localhost:8080/capture-screenshots", {
                    method: "POST",
                    body: formData,
                  });
                  const result = await response.json();
                  console.log("Screenshot upload response:", result);
                } catch (err) {
                  console.error("Error sending screenshot to server:", err);
                }
              }
            }, 'image/png');
          }
        }
      } catch (err) {
        console.error("Error capturing frame:", err);
      }
    }
  };

  const compareFrames = (previousFrame: ImageData, currentFrame: ImageData) => {
    const pixelCount = previousFrame.width * previousFrame.height;
    let diffCount = 0;

    for (let i = 0; i < pixelCount; i++) {
      const prevIndex = i * 4;
      const currIndex = i * 4;

      const rDiff = Math.abs(previousFrame.data[prevIndex] - currentFrame.data[currIndex]);
      const gDiff = Math.abs(previousFrame.data[prevIndex + 1] - currentFrame.data[currIndex + 1]);
      const bDiff = Math.abs(previousFrame.data[prevIndex + 2] - currentFrame.data[currIndex + 2]);
      const aDiff = Math.abs(previousFrame.data[prevIndex + 3] - currentFrame.data[currIndex + 3]);

      if (rDiff + gDiff + bDiff + aDiff > 50) {
        diffCount++;
      }
    }

    return (diffCount / pixelCount) > 0.01;
  };

  return (
    <div className="bg-gray-800 p-8 rounded-lg shadow-lg flex flex-col items-center mb-12">
      <button
        onClick={isRecording ? stopRecording : startRecording}
        className={`${
          isRecording ? "bg-green-500 hover:bg-green-600" : "bg-red-500 hover:bg-red-600"
        } text-white font-semibold py-3 px-6 rounded-lg mb-4`}
      >
        {isRecording ? "Stop Capturing" : "Start Capturing"}
      </button>
      <canvas ref={canvasRef} style={{ display: 'none' }} />
    </div>
  );
};

export default MediaCapture;

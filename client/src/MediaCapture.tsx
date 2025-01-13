import React, { useEffect, useRef } from 'react';

// Declare ImageCapture type if not available
declare class ImageCapture {
  constructor(track: MediaStreamTrack);
  grabFrame(): Promise<ImageBitmap>;
}

const MediaCapture: React.FC<{ isRecording: boolean }> = ({ isRecording }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const previousFrameRef = useRef<ImageData | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null); // Keep reference to the MediaStream
  const screenshotIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null); // Keep reference to MediaRecorder
  const audioChunksRef = useRef<Blob[]>([]); // Store audio chunks

  // Start the screenshot capture process
  const startRecording = async () => {
    console.log("Starting screen capture...");
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
      console.log("Screen and audio stream obtained", stream);

      // Keep reference to the media stream
      mediaStreamRef.current = stream;

      // Start capturing screenshots every second
      screenshotIntervalRef.current = setInterval(() => captureScreenshot(stream), 1000);

      // Start recording audio in 10-second chunks
      startAudioRecording(stream);
    } catch (err) {
      console.error("Error starting screen capture:", err);
    }
  };

  // Stop the screenshot capture and audio recording
  const stopRecording = () => {
    console.log("Stopping capture...");
    if (screenshotIntervalRef.current) clearInterval(screenshotIntervalRef.current);

    if (mediaStreamRef.current) {
      mediaStreamRef.current.getTracks().forEach(track => track.stop()); // Stop all tracks
    }

    if (mediaRecorderRef.current) {
      mediaRecorderRef.current.stop(); // Stop audio recording
    }

    // Clear previous audio chunks
    audioChunksRef.current = [];
  };

  // Capture screenshots from the screen
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

  // Compare two frames to check for changes
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

  // Start recording audio in 10-second chunks
  const startAudioRecording = (stream: MediaStream) => {
    // Create the MediaRecorder
    const mediaRecorder = new MediaRecorder(stream);
    mediaRecorderRef.current = mediaRecorder;

    // Audio chunk storage
    audioChunksRef.current = [];

    // Data available event handler
    mediaRecorder.ondataavailable = (e) => {
      audioChunksRef.current.push(e.data);
    };

    // Stop event handler to process the audio and send it to the server
    const processAudioAndRestartRecording = () => {
      const audioBlob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
      audioChunksRef.current = []; // Clear the chunks
      const formData = new FormData();
      formData.append("audio", audioBlob, `audio-${Date.now()}.webm`);

      // Send the audio to the server
      fetch("http://localhost:8080/capture-audio", {
        method: "POST",
        body: formData,
      })
        .then(response => response.json())
        .then(data => console.log("Audio upload response:", data))
        .catch(err => console.error("Error uploading audio:", err));
    };

    // Start the recording
    const startNewRecording = () => {
      mediaRecorder.start();
      console.log('Audio recording started.');

      // Set a timeout to stop recording after 10 seconds
      setTimeout(() => {
        if (mediaRecorder.state === 'recording') {
          console.log("Stopping recording after 10 seconds.");
          mediaRecorder.stop(); // This will trigger the onstop event
        }
      }, 10000); // Stop recording after 10 seconds
    };

    // Event handler for when the recording stops
    mediaRecorder.onstop = () => {
      processAudioAndRestartRecording(); // Process the audio and send it
      startNewRecording(); // Start a new recording after stopping
    };

    // Start the initial recording
    startNewRecording();
  };

  useEffect(() => {
    // Only start recording if `isRecording` is true
    if (isRecording && !mediaStreamRef.current) {
      startRecording();
    } else if (!isRecording && mediaStreamRef.current) {
      stopRecording();
    }

    return () => {
      // Cleanup when component unmounts
      stopRecording();
    };
  }, [isRecording]);

  return (
    <>
      <video style={{ display: 'none' }} autoPlay muted />
      <canvas ref={canvasRef} style={{ display: 'none' }} />
    </>
  );
};

export default MediaCapture;

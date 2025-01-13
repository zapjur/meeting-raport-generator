import React, { useEffect, useRef } from 'react';

// Declare ImageCapture type if not available
declare class ImageCapture {
  constructor(track: MediaStreamTrack);
  grabFrame(): Promise<ImageBitmap>;
}

const MediaCapture: React.FC<{ isRecording: boolean }> = ({ isRecording }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const previousFrameRef = useRef<ImageData | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null); 
  const screenshotIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null); 
  const audioChunksRef = useRef<Blob[]>([]);

  // Function to start capturing media (video + audio)
  const startRecording = async () => {
    console.log("Starting screen capture...");
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
      console.log("Screen and audio stream obtained", stream);

      // Store the media stream reference
      mediaStreamRef.current = stream;

      // Start capturing screenshots at regular intervals
      screenshotIntervalRef.current = setInterval(() => captureScreenshot(stream), 1000);

      // Start audio recording in chunks
      startAudioRecording(stream);
    } catch (err) {
      console.error("Error starting screen capture:", err);
    }
  };

  // Function to stop media capture and audio recording
  const stopRecording = () => {
    console.log("Stopping capture...");
    if (screenshotIntervalRef.current) clearInterval(screenshotIntervalRef.current);

    // Stop all media stream tracks
    mediaStreamRef.current?.getTracks().forEach(track => track.stop());

    // Stop audio recording
    mediaRecorderRef.current?.stop();

    // Clear audio chunks
    audioChunksRef.current = [];


  };

  // Function to capture screenshots and upload them to the server
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

            // Upload the screenshot
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

  // Compare two frames for significant changes
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
    const mediaRecorder = new MediaRecorder(stream);
    mediaRecorderRef.current = mediaRecorder;

    // Store audio chunks
    audioChunksRef.current = [];

    // Store audio data when available
    mediaRecorder.ondataavailable = (e) => {
      audioChunksRef.current.push(e.data);
    };

    // Process and send audio data when recording stops
    const processAudioAndRestartRecording = () => {
      const audioBlob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
      audioChunksRef.current = []; 

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

    // Start recording and automatically stop after 10 seconds
    const startNewRecording = () => {
      mediaRecorder.start();
      console.log('Audio recording started.');

      setTimeout(() => {
        if (mediaRecorder.state === 'recording') {
          console.log("Stopping recording after 10 seconds.");
          mediaRecorder.stop(); 
        }
      }, 10000); 
    };

    // Event handler when the recording stops
    mediaRecorder.onstop = () => {
      processAudioAndRestartRecording();
      startNewRecording();
    };

    // Start the initial recording
    startNewRecording();
  };

  // Effect to handle recording based on `isRecording` prop
  useEffect(() => {
    if (isRecording && !mediaStreamRef.current) {
      startRecording();
    } else if (!isRecording && mediaStreamRef.current) {
      stopRecording();
    }

    return () => {
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

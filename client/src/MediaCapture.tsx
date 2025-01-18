import React, { useEffect, useRef } from "react";

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
  const isAudioRecordingRef = useRef<boolean>(false);

  const startRecording = async () => {
    console.log("Starting screen capture...");
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
      console.log("Screen and audio stream obtained", stream);

      mediaStreamRef.current = stream;

      screenshotIntervalRef.current = setInterval(() => captureScreenshot(stream), 1000);

      startAudioRecording(stream);
    } catch (err) {
      console.error("Error starting screen capture:", err);
    }
  };

  const stopRecording = () => {
    console.log("Stopping capture...");
    if (screenshotIntervalRef.current) clearInterval(screenshotIntervalRef.current);

    mediaStreamRef.current?.getTracks().forEach((track) => track.stop());

    isAudioRecordingRef.current = false;
    mediaRecorderRef.current?.stop();

    // Reset all references to allow re-initialization
    mediaStreamRef.current = null;
    mediaRecorderRef.current = null;
    audioChunksRef.current = [];
  };

  const captureScreenshot = async (stream: MediaStream) => {
    if (canvasRef.current) {
      const canvas = canvasRef.current;
      const videoTrack = stream.getVideoTracks()[0];
      const imageCapture = new ImageCapture(videoTrack);

      try {
        const imageBitmap = await imageCapture.grabFrame();
        const ctx = canvas.getContext("2d");

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
                  const response = await fetch("http://127.0.0.1:8080/capture-screenshots", {
                    method: "POST",
                    body: formData,
                  });
                  const result = await response.json();
                  console.log("Screenshot upload response:", result);
                } catch (err) {
                  console.error("Error sending screenshot to server:", err);
                }
              }
            }, "image/png");
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

    return diffCount / pixelCount > 0.01;
  };

  const startAudioRecording = (stream: MediaStream) => {
    const mediaRecorder = new MediaRecorder(stream);
    mediaRecorderRef.current = mediaRecorder;

    audioChunksRef.current = [];
    isAudioRecordingRef.current = true;

    mediaRecorder.ondataavailable = (e) => {
      audioChunksRef.current.push(e.data);
    };

    const processAudioAndRestartRecording = () => {
      if (!isAudioRecordingRef.current) return;

      const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
      audioChunksRef.current = [];

      const formData = new FormData();
      formData.append("audio", audioBlob, `audio-${Date.now()}.webm`);

      fetch("http://127.0.0.1:8080/capture-audio", {
        method: "POST",
        body: formData,
      })
        .then((response) => response.json())
        .then((data) => console.log("Audio upload response:", data))
        .catch((err) => console.error("Error uploading audio:", err));
    };

    const startNewRecording = () => {
      if (!isAudioRecordingRef.current) return;

      mediaRecorder.start();
      console.log("Audio recording started.");

      setTimeout(() => {
        if (mediaRecorder.state === "recording" && isAudioRecordingRef.current) {
          console.log("Stopping recording after 10 seconds.");
          mediaRecorder.stop();
        }
      }, 10000);
    };

    mediaRecorder.onstop = () => {
      processAudioAndRestartRecording();
      startNewRecording();
    };

    startNewRecording();
  };

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
      <video style={{ display: "none" }} autoPlay muted />
      <canvas ref={canvasRef} style={{ display: "none" }} />
    </>
  );
};

export default MediaCapture;

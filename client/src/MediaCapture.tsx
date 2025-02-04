import React, { useEffect, useRef, useState } from "react";

declare class ImageCapture {
  constructor(track: MediaStreamTrack);
  grabFrame(): Promise<ImageBitmap>;
}

interface MediaCaptureProps {
  isRecording: boolean;
  meetingId: string | null;
}

// CONF
const SCREENSHOT_INTERVAL_MS = 2000; // 2 sek
const AUDIO_CAPTURE_INTERVAL_MS = 20000; // 15 sek
const FRAME_CHANGE_THRESHOLD = 0.1; //10 %

const MediaCapture: React.FC<MediaCaptureProps> = ({ isRecording, meetingId }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const previousFrameRef = useRef<ImageData | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const isAudioRecordingRef = useRef<boolean>(false);
  const lastScreenshotTimeRef = useRef<number>(0);
  const screenshotIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startRecording = async () => {
    if (!meetingId) {
      console.error("Meeting ID is required to start recording.");
      return;
    }

    console.log("Starting screen capture...");
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
      console.log("Screen and audio stream obtained", stream);

      mediaStreamRef.current = stream;

      // Ensure we clear the previous interval before setting a new one
      if (screenshotIntervalRef.current) {
        clearInterval(screenshotIntervalRef.current);
      }

      screenshotIntervalRef.current = setInterval(() => captureScreenshot(stream), SCREENSHOT_INTERVAL_MS);

      startAudioRecording(stream);
    } catch (err) {
      console.error("Error starting screen capture:", err);
    }
  };

  const stopRecording = () => {
    console.log("Stopping capture...");
    // Clear the interval to stop screenshots from being captured
    if (screenshotIntervalRef.current) clearInterval(screenshotIntervalRef.current);

    // Stop the media stream tracks
    mediaStreamRef.current?.getTracks().forEach((track) => track.stop());

    isAudioRecordingRef.current = false;
    mediaRecorderRef.current?.stop();

    mediaStreamRef.current = null;
    mediaRecorderRef.current = null;
    audioChunksRef.current = [];
  };

  const captureScreenshot = async (stream: MediaStream) => {
    if (!meetingId) {
      console.error("Meeting ID is required to capture screenshots.");
      return;
    }

    const now = Date.now();

    if (now - lastScreenshotTimeRef.current < SCREENSHOT_INTERVAL_MS) {
      return;
    }

    if (canvasRef.current) {
      const canvas = canvasRef.current;
      const videoTrack = stream.getVideoTracks()[0];

      if (videoTrack.readyState === "live") {
        const imageCapture = new ImageCapture(videoTrack);

        try {
          const imageBitmap = await imageCapture.grabFrame();
          const ctx = canvas.getContext("2d");

          if (ctx) {
            canvas.width = imageBitmap.width;
            canvas.height = imageBitmap.height;
            ctx.drawImage(imageBitmap, 0, 0);

            const currentFrameData = ctx.getImageData(0, 0, canvas.width, canvas.height);

            // Capture the screenshot only if the frame has changed
            if (!previousFrameRef.current || compareFrames(previousFrameRef.current, currentFrameData)) {
              previousFrameRef.current = currentFrameData;

              // Update the last screenshot time to the current time
              lastScreenshotTimeRef.current = now;

              // Capture and upload the screenshot
              canvas.toBlob(async (blob) => {
                if (blob) {
                  const formData = new FormData();
                  formData.append("screenshot", blob, `screenshot-${Date.now()}.png`);
                  formData.append("meeting_id", meetingId);

                  try {
                    const response = await fetch("http://127.0.0.1:8080/capture-screenshots", {
                      method: "POST",
                      body: formData,
                    });
                    const result = await response.text();
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
      } else {
        console.error("Video track is not in a valid state.");
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

    return diffCount / pixelCount > FRAME_CHANGE_THRESHOLD;
  };

  const startAudioRecording = (stream: MediaStream) => {
    if (!meetingId) {
      console.error("Meeting ID is required to start audio recording.");
      return;
    }

    const mediaRecorder = new MediaRecorder(stream);
    mediaRecorderRef.current = mediaRecorder;

    audioChunksRef.current = [];
    isAudioRecordingRef.current = true;

    mediaRecorder.ondataavailable = (e) => {
      audioChunksRef.current.push(e.data);
    };

    const processAudioAndRestartRecording = async () => {
      if (!isAudioRecordingRef.current) return;

      const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
      audioChunksRef.current = [];

      const formData = new FormData();
      formData.append("audio", audioBlob, `audio-${Date.now()}.webm`);
      formData.append("meeting_id", meetingId);

      try {
        const response = await fetch("http://127.0.0.1:8080/capture-audio", {
          method: "POST",
          body: formData,
        });
        const result = await response.text();
        console.log("Audio upload response:", result);
      } catch (err) {
        console.error("Error uploading audio:", err);
      }
    };

    const startNewRecording = () => {
      if (!isAudioRecordingRef.current) return;

      mediaRecorder.start();
      console.log("Audio recording started.");

      setTimeout(() => {
        if (mediaRecorder.state === "recording" && isAudioRecordingRef.current) {
          console.log("Stopping recording after interval.");
          mediaRecorder.stop();
        }
      }, AUDIO_CAPTURE_INTERVAL_MS);
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

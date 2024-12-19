import { useEffect } from 'react';
import * as monaco from 'monaco-editor';
import { MessageType } from '../types/types';

export const useEventSource = ({
  apiBaseUrl,
  onScript,
  onVideo,
  onError,
  generationTimeoutRef,
  editorRef,
  setIsGenerating,
}: {
  apiBaseUrl: string;
  onScript: (script: string) => void;
  onVideo: (url: string) => void;
  onError: (error: string) => void;
  generationTimeoutRef: React.MutableRefObject<number | null>;
  editorRef: React.MutableRefObject<monaco.editor.IStandaloneCodeEditor | null>;
  setIsGenerating: (state: boolean) => void;
}) => {
  useEffect(() => {
    const initializeEventSource = async () => {
      try {
        const healthzResponse = await fetch(`${apiBaseUrl}/healthz`, {
          method: 'GET',
          credentials: 'include',
        });

        if (!healthzResponse.ok) {
          throw new Error(`Health check failed with status: ${healthzResponse.status}`);
        }

        const eventSource = new EventSource(`${apiBaseUrl}/events`, { withCredentials: true });

        eventSource.onmessage = (event) => {
          const message: MessageType = JSON.parse(event.data);

          if (generationTimeoutRef.current) {
            clearTimeout(generationTimeoutRef.current);
          }

          switch (message.type) {
            case 'script': {
              const receivedScript = message.content;
              onScript(receivedScript);
              if (editorRef.current) {
                editorRef.current.setValue(receivedScript);
              }
              setIsGenerating(false);
              break;
            }
            case 'video':
              onVideo(message.content);
              setIsGenerating(false);
              break;
          }
        };

        eventSource.onerror = (error) => {
          console.error('EventSource failed:', error);
          eventSource.close();
          setIsGenerating(false);
          onError('Connection error. Please refresh and try again.');
        };

        return () => {
          if (generationTimeoutRef.current) {
            clearTimeout(generationTimeoutRef.current);
          }
          eventSource.close();
        };
      } catch (error) {
        console.error('Initialization failed:', error);
        onError('Health check failed. Unable to connect to the server.');
        setIsGenerating(false);
      }
    };

    initializeEventSource();
  }, [apiBaseUrl, onScript, onVideo, onError, generationTimeoutRef, editorRef, setIsGenerating]);
}
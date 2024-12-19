export interface MessageType {
    type: 'script' | 'video' | 'compiled';
    sessionId: string;
    status: 'success' | 'error';
    content: string;
    details?: Record<string, unknown>;
  }
  
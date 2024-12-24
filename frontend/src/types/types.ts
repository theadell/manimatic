export type EventKind =
  | 'compile_requested'
  | 'compile_succeeded'
  | 'compile_failed'
  | 'generate_succeeded'
  | 'generate_failed';

export interface Event<T extends EventKind> {
  kind: T;
  sessionId: string;
  data: EventData<T>;
}

export type EventData<T extends EventKind> = T extends 'compile_requested'
  ? CompileRequest
  : T extends 'compile_succeeded'
  ? CompileSuccess
  : T extends 'compile_failed'
  ? CompileError
  : T extends 'generate_succeeded'
  ? GenerateSuccess
  : T extends 'generate_failed'
  ? GenerateError
  : never;

export interface CompileRequest {
  script: string;
}

export interface CompileSuccess {
  video_url: string;
}

export interface CompileError {
  message: string;
  stdout: string;
  stderr: string;
  line?: number;
}

export interface GenerateSuccess {
  script: string;
}

export interface GenerateError {
  message: string;
  details?: string;
  model: string;
}
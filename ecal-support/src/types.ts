export interface EcalAstNode {
  allowescapes: boolean;
  children: EcalAstNode[];
  id: number;
  identifier: boolean;
  line: number;
  linepos: number;
  name: string;
  pos: number;
  source: string;
  value: any;
}

export interface ThreadInspection {
  callStack: string[];
  callStackNode?: EcalAstNode[];
  callStackVsSnapshot?: Record<string, any>[];
  callStackVsSnapshotGlobal?: Record<string, any>[];
  threadRunning: boolean;
  code?: string;
  node?: EcalAstNode;
  vs?: Record<string, any>;
  vsGlobal?: Record<string, any>;
}

export interface ThreadStatus {
  callStack: string[];
  threadRunning?: boolean;
}

export interface DebugStatus {
  breakonstart: boolean;
  breakpoints: Record<string, boolean>;
  sources: string[];
  threads: Record<string, ThreadStatus>;
}

/**
 * Log output stream for this client.
 */
export interface LogOutputStream {
  log(value: string): void;
  error(value: string): void;
}

export interface ClientBreakEvent {
  tid: number;
  inspection: ThreadInspection;
}

export enum ContType {
  Resume = "Resume",
  StepIn = "StepIn",
  StepOver = "StepOver",
  StepOut = "StepOut",
}

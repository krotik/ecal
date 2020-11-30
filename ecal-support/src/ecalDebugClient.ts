/**
 * Debug client implementation for the ECAL debugger.
 */

import * as net from "net";
import { EventEmitter } from "events";
import PromiseSocket from "promise-socket";
import {
  LogOutputStream,
  DebugStatus,
  ThreadInspection,
  ContType,
} from "./types";

interface BacklogCommand {
  cmd: string;
  args?: string[];
}

/**
 * Debug client for ECAL debug server.
 */
export class ECALDebugClient extends EventEmitter {
  private socket: PromiseSocket<net.Socket>;
  private socketLock: any;
  private connected: boolean = false;
  private backlog: BacklogCommand[] = [];
  private threadInspection: Record<number, ThreadInspection> = {};
  private doReload: boolean = false;

  /**
   * Create a new debug client.
   */
  public constructor(protected out: LogOutputStream) {
    super();
    this.socket = new PromiseSocket(new net.Socket());

    const AsyncLock = require("async-lock");
    this.socketLock = new AsyncLock();
  }

  public async connect(host: string, port: number) {
    try {
      this.out.log(`Connecting to: ${host}:${port}`);
      await this.socket.connect({ port, host });
      this.connected = true;
      this.pollEvents(); // Start emitting events
    } catch (e) {
      this.out.error(`Could not connect to debug server: ${e}`);
    }
  }

  public async status(): Promise<DebugStatus | null> {
    try {
      return (await this.sendCommand("status")) as DebugStatus;
    } catch (e) {
      this.out.error(`Could not query for status: ${e}`);
      return null;
    }
  }

  public reload() {
    this.doReload = true;
  }

  public async describe(tid: number): Promise<ThreadInspection | null> {
    try {
      return (await this.sendCommand("describe", [
        String(tid),
      ])) as ThreadInspection;
    } catch (e) {
      this.out.error(`Could not inspect thread ${tid}: ${e}`);
      return null;
    }
  }

  public async cont(tid: number, type: ContType) {
    try {
      await this.sendCommand("cont", [String(tid), type]);
      delete this.threadInspection[tid];
    } catch (e) {
      this.out.error(`Could not continue thread ${tid}: ${e}`);
    }
  }

  public async setBreakpoint(breakpoint: string) {
    try {
      (await this.sendCommand(`break ${breakpoint}`)) as DebugStatus;
    } catch (e) {
      this.out.error(`Could not set breakpoint ${breakpoint}: ${e}`);
    }
  }

  public async clearBreakpoints(source: string) {
    try {
      (await this.sendCommand("rmbreak", [source])) as DebugStatus;
    } catch (e) {
      this.out.error(`Could not remove breakpoints for ${source}: ${e}`);
    }
  }

  public async shutdown() {
    this.connected = false;
    await this.socket.destroy();
  }

  /**
   * PollEvents is the polling loop for debug events.
   */
  private async pollEvents() {
    let nextLoop = 1000;
    try {
      const status = await this.status();

      this.emit("status", status);

      for (const [tidString, thread] of Object.entries(status?.threads || [])) {
        const tid = parseInt(tidString);

        if (thread.threadRunning === false && !this.threadInspection[tid]) {
          // A thread was stopped inspect it

          let inspection: ThreadInspection = {
            callStack: [],
            threadRunning: false,
          };

          try {
            inspection = (await this.sendCommand("describe", [
              String(tid),
            ])) as ThreadInspection;
          } catch (e) {
            this.out.error(`Could not get description for ${tid}: ${e}`);
          }

          this.threadInspection[tid] = inspection;

          this.emit("pauseOnBreakpoint", { tid, inspection });
        }
      }

      if (this.doReload) {
        this.doReload = false;
        this.out.log(`Reloading interpreter state`);
        try {
          await this.sendCommandString("@reload\r\n");
        } catch (e) {
          this.out.error(`Could not reload the interpreter state: ${e}`);
        }
      }
    } catch (e) {
      this.out.error(`Error during event loop: ${e}`);
      nextLoop = 5000;
    }

    if (this.connected) {
      setTimeout(this.pollEvents.bind(this), nextLoop);
    } else {
      this.out.log("Stop emitting events" + nextLoop);
    }
  }

  public async sendCommand(cmd: string, args?: string[]): Promise<any> {
    // Create or process the backlog depending on the connection status

    if (!this.connected) {
      this.backlog.push({
        cmd,
        args,
      });
      return null;
    } else if (this.backlog.length > 0) {
      const backlog = this.backlog;
      this.backlog = [];
      for (const item of backlog) {
        await this.sendCommand(item.cmd, item.args);
      }
    }

    return await this.sendCommandString(
      `##${cmd} ${args ? args.join(" ") : ""}\r\n`
    );
  }

  public async sendCommandString(cmdString: string): Promise<any> {
    // Socket needs to be locked. Reading and writing to the socket is seen
    // by the interpreter as async (i/o bound) code. Separate calls to
    // sendCommand will be executed in different event loops. Without the lock
    // the different sendCommand calls would mix their responses.

    return await this.socketLock.acquire("socket", async () => {
      await this.socket.write(cmdString, "utf8");

      let text = "";

      while (!text.endsWith("\n\n")) {
        text += await this.socket.read(1);
      }

      let res: any = {};

      try {
        res = JSON.parse(text);
      } catch (e) {
        throw new Error(`Could not parse response: ${text} - error:${e}`);
      }
      if (res?.DebuggerError) {
        throw new Error(
          `Unexpected internal error for command "${cmdString}": ${res.DebuggerError}`
        );
      }
      if (res?.EncodedOutput !== undefined) {
        res = Buffer.from(res.EncodedOutput, "base64").toString("utf8");
      }
      return res;
    });
  }
}

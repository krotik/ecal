/**
 * Debug Adapter for VS Code to support the ECAL debugger.
 *
 * Based on the debugger extension guide:
 * https://code.visualstudio.com/api/extension-guides/debugger-extension
 */

import {
  logger,
  Logger,
  LoggingDebugSession,
  Thread,
  Source,
  Breakpoint,
  InitializedEvent,
  BreakpointEvent,
  StoppedEvent,
  StackFrame,
  Scope,
  Variable,
} from "vscode-debugadapter";
import { DebugProtocol } from "vscode-debugprotocol";
import { WaitGroup } from "@jpwilliams/waitgroup";
import { ECALDebugClient } from "./ecalDebugClient";
import * as vscode from "vscode";
import { ClientBreakEvent, ContType, DebugStatus } from "./types";
import * as path from "path";

/**
 * ECALDebugArguments are the arguments which VSCode can pass to the debug adapter.
 * This defines the parameter which a VSCode instance using the ECAL extension can pass to the
 * debug adapter from a lauch configuration ('.vscode/launch.json') in a project folder.
 */
interface ECALDebugArguments extends DebugProtocol.LaunchRequestArguments {
  host: string; // Host of the ECAL debug server
  port: number; // Port of the ECAL debug server
  dir: string; // Root directory for ECAL interpreter
  executeOnEntry?: boolean; // Flag if the debugged script should be executed when the debug session is started
  trace?: boolean; // Flag to enable verbose logging of the adapter protocol
}

/**
 * Debug adapter implementation.
 *
 * Uses: https://github.com/microsoft/vscode-debugadapter-node
 *
 * See the Debug Adapter Protocol (DAP) documentation:
 * https://microsoft.github.io/debug-adapter-protocol/overview#How_it_works
 */
export class ECALDebugSession extends LoggingDebugSession {
  /**
   * Client to the ECAL debug server
   */
  private client: ECALDebugClient;

  /**
   * Output channel for log messages
   */
  private extout: vscode.OutputChannel = vscode.window.createOutputChannel(
    "ECAL Debug Session"
  );

  /**
   * WaitGroup to wait the finish of the configuration sequence
   */
  private wgConfig = new WaitGroup();

  private config: ECALDebugArguments = {} as ECALDebugArguments;

  private unconfirmedBreakpoints: DebugProtocol.Breakpoint[] = [];
  private frameVariableScopes: Record<number, Record<string, any>> = {};
  private frameVariableGlobalScopes: Record<number, Record<string, any>> = {};

  private bpCount: number = 1;
  private sfCount: number = 1;
  private vsCount: number = 1;

  private bpIds: Record<string, number> = {};
  private sfIds: Record<string, number> = {};
  private vsIds: Record<string, number> = {};

  /**
   * Create a new debug adapter which is used for one debug session.
   */
  public constructor() {
    super("mock-debug.txt");

    this.extout.appendLine("Creating Debug Session");
    this.client = new ECALDebugClient(new LogChannelAdapter(this.extout));

    // Add event handlers

    this.client.on("pauseOnBreakpoint", (e: ClientBreakEvent) => {
      this.sendEvent(new StoppedEvent("breakpoint", e.tid));
    });

    this.client.on("status", (e: DebugStatus) => {
      try {
        if (this.unconfirmedBreakpoints.length > 0) {
          for (const toConfirm of this.unconfirmedBreakpoints) {
            for (const [breakpointString, ok] of Object.entries(
              e.breakpoints
            )) {
              const line = parseInt(breakpointString.split(":")[1]);
              if (ok) {
                if (
                  toConfirm.line === line &&
                  toConfirm.source?.name === breakpointString
                ) {
                  toConfirm.verified = true;
                  this.sendEvent(new BreakpointEvent("changed", toConfirm));
                }
              }
            }
          }
          this.unconfirmedBreakpoints = [];
        }
      } catch (e) {
        console.error(e);
      }
    });

    // Lines and columns start at 1
    this.setDebuggerLinesStartAt1(true);
    this.setDebuggerColumnsStartAt1(true);

    // Increment the config WaitGroup counter for configurationDoneRequest()
    this.wgConfig.add(1);
  }

  /**
   * Called as the first step in the DAP. The client (e.g. VSCode)
   * interrogates the debug adapter on the features which it provides.
   */
  protected initializeRequest(
    response: DebugProtocol.InitializeResponse
  ): void {
    response.body = response.body || {};

    // The adapter implements the configurationDoneRequest.
    response.body.supportsConfigurationDoneRequest = true;

    // make VS Code send the breakpointLocations request
    response.body.supportsBreakpointLocationsRequest = true;

    // make VS Code provide "Step in Target" functionality
    response.body.supportsStepInTargetsRequest = true;

    this.sendResponse(response);

    this.sendEvent(new InitializedEvent());
  }

  /**
   * Called as part of the "configuration Done" step in the DAP. The client (e.g. VSCode) has
   * finished the initialization of the debug adapter.
   */
  protected configurationDoneRequest(
    response: DebugProtocol.ConfigurationDoneResponse,
    args: DebugProtocol.ConfigurationDoneArguments
  ): void {
    super.configurationDoneRequest(response, args);
    this.wgConfig.done();
  }

  /**
   * The client (e.g. VSCode) asks the debug adapter to start the debuggee communication.
   */
  protected async launchRequest(
    response: DebugProtocol.LaunchResponse,
    args: ECALDebugArguments
  ) {
    this.config = args; // Store the configuration

    // Setup logging either verbose or just on errors

    logger.setup(
      args.trace ? Logger.LogLevel.Verbose : Logger.LogLevel.Error,
      false
    );

    await this.wgConfig.wait(); // Wait for configuration sequence to finish

    this.extout.appendLine(`Configuration loaded: ${JSON.stringify(args)}`);

    await this.client.connect(args.host, args.port);

    if (args.executeOnEntry) {
      this.client.reload();
    }

    this.sendResponse(response);
  }

  protected async setBreakPointsRequest(
    response: DebugProtocol.SetBreakpointsResponse,
    args: DebugProtocol.SetBreakpointsArguments
  ): Promise<void> {
    let breakpoints: DebugProtocol.Breakpoint[] = [];

    if (args.source.path?.indexOf(this.config.dir) === 0) {
      const sourcePath = args.source.path.slice(this.config.dir.length + 1);

      // Clear all breakpoints of the file

      await this.client.clearBreakpoints(sourcePath);

      // Send all breakpoint requests to the debug server

      for (const sbp of args.breakpoints || []) {
        await this.client.setBreakpoint(`${sourcePath}:${sbp.line}`);
      }

      // Confirm that the breakpoints have been set

      const status = await this.client.status();

      if (status) {
        breakpoints = (args.lines || []).map((line) => {
          const breakpointString = `${sourcePath}:${line}`;

          const bp: DebugProtocol.Breakpoint = new Breakpoint(
            status.breakpoints[breakpointString],
            line,
            undefined,
            new Source(breakpointString, args.source.path)
          );
          bp.id = this.getBreakPointId(breakpointString);

          return bp;
        });
      } else {
        breakpoints = (args.breakpoints || []).map((sbp) => {
          const breakpointString = `${sourcePath}:${sbp.line}`;

          const bp: DebugProtocol.Breakpoint = new Breakpoint(
            false,
            sbp.line,
            undefined,
            new Source(breakpointString, args.source.path)
          );
          bp.id = this.getBreakPointId(breakpointString);

          return bp;
        });

        this.unconfirmedBreakpoints = breakpoints;
      }
    }

    response.body = {
      breakpoints,
    };

    this.sendResponse(response);
  }

  protected async breakpointLocationsRequest(
    response: DebugProtocol.BreakpointLocationsResponse,
    args: DebugProtocol.BreakpointLocationsArguments
  ) {
    let breakpoints: DebugProtocol.BreakpointLocation[] = [];

    if (args.source.path?.indexOf(this.config.dir) === 0) {
      const sourcePath = args.source.path.slice(this.config.dir.length + 1);
      const status = await this.client.status();

      if (status) {
        for (const [breakpointString, v] of Object.entries(
          status.breakpoints
        )) {
          if (v) {
            const line = parseInt(breakpointString.split(":")[1]);
            if (`${sourcePath}:${line}` === breakpointString) {
              breakpoints.push({
                line,
              });
            }
          }
        }
      }
    }
    response.body = {
      breakpoints,
    };

    this.sendResponse(response);
  }

  protected async threadsRequest(
    response: DebugProtocol.ThreadsResponse
  ): Promise<void> {
    const status = await this.client.status();
    const threads = [];

    if (status) {
      for (const tid of Object.keys(status.threads)) {
        threads.push(new Thread(parseInt(tid), `Thread ${tid}`));
      }
    } else {
      threads.push(new Thread(1, "Thread 1"));
    }

    response.body = {
      threads,
    };

    this.sendResponse(response);
  }

  protected async stackTraceRequest(
    response: DebugProtocol.StackTraceResponse,
    args: DebugProtocol.StackTraceArguments
  ) {
    const stackFrames: StackFrame[] = [];
    const status = await this.client.status();
    const threadStatus = status?.threads[String(args.threadId)];

    if (threadStatus?.threadRunning === false) {
      const ins = await this.client.describe(args.threadId);

      if (ins) {
        for (const [i, sf] of ins.callStack.entries()) {
          const sfNode = ins.callStackNode![i];
          const frameId = this.getStackFrameId(args.threadId, sf, i);
          const breakpointString = `${sfNode.source}:${sfNode.line}`;

          stackFrames.unshift(
            new StackFrame(
              frameId,
              sf,
              new Source(
                breakpointString,
                path.join(this.config.dir, sfNode.source)
              ),
              sfNode.line
            )
          );
          this.frameVariableScopes[frameId] = ins.callStackVsSnapshot![i];
          this.frameVariableGlobalScopes[
            frameId
          ] = ins.callStackVsSnapshotGlobal![i];
        }

        const frameId = this.getStackFrameId(
          args.threadId,
          ins.code!,
          ins.callStack.length
        );
        const breakpointString = `${ins.node!.source}:${ins.node!.line}`;

        stackFrames.unshift(
          new StackFrame(
            frameId,
            ins.code!,
            new Source(
              breakpointString,
              path.join(this.config.dir, ins.node!.source)
            ),
            ins.node!.line
          )
        );
        this.frameVariableScopes[frameId] = ins.vs!;
        this.frameVariableGlobalScopes[frameId] = ins.vsGlobal!;
      }
    }

    response.body = {
      stackFrames,
    };
    this.sendResponse(response);
  }

  protected scopesRequest(
    response: DebugProtocol.ScopesResponse,
    args: DebugProtocol.ScopesArguments
  ): void {
    response.body = {
      scopes: [
        new Scope("Local", this.getVariableScopeId(args.frameId, "local")),
        new Scope("Global", this.getVariableScopeId(args.frameId, "global")),
      ],
    };

    this.sendResponse(response);
  }

  protected async variablesRequest(
    response: DebugProtocol.VariablesResponse,
    args: DebugProtocol.VariablesArguments
  ) {
    let vs: Record<string, any> = {};
    let variables: Variable[] = [];

    const [frameId, scopeType] = this.getScopeLookupInfo(
      args.variablesReference
    );

    if (scopeType === "local") {
      vs = this.frameVariableScopes[frameId];
    } else if (scopeType === "global") {
      vs = this.frameVariableGlobalScopes[frameId];
    }

    if (vs) {
      for (const [name, val] of Object.entries(vs)) {
        let valString: string;

        try {
          valString = JSON.stringify(val);
        } catch (e) {
          valString = String(val);
        }

        variables.push(new Variable(name, valString));
      }
    }

    response.body = {
      variables,
    };
    this.sendResponse(response);
  }

  protected async continueRequest(
    response: DebugProtocol.ContinueResponse,
    args: DebugProtocol.ContinueArguments
  ) {
    await this.client.cont(args.threadId, ContType.Resume);
    response.body = {
      allThreadsContinued: false,
    };
    this.sendResponse(response);
  }

  protected async nextRequest(
    response: DebugProtocol.NextResponse,
    args: DebugProtocol.NextArguments
  ) {
    await this.client.cont(args.threadId, ContType.StepOver);
    this.sendResponse(response);
  }

  protected async stepInRequest(
    response: DebugProtocol.StepInResponse,
    args: DebugProtocol.StepInArguments
  ) {
    await this.client.cont(args.threadId, ContType.StepIn);
    this.sendResponse(response);
  }

  protected async stepOutRequest(
    response: DebugProtocol.StepOutResponse,
    args: DebugProtocol.StepOutArguments
  ) {
    await this.client.cont(args.threadId, ContType.StepOut);
    this.sendResponse(response);
  }

  protected async evaluateRequest(
    response: DebugProtocol.EvaluateResponse,
    args: DebugProtocol.EvaluateArguments
  ): Promise<void> {
    let result: any;

    try {
      result = await this.client.sendCommandString(`${args.expression}\r\n`);

      if (typeof result !== "string") {
        result = JSON.stringify(result, null, "  ");
      }
    } catch (e) {
      result = String(e);
    }

    response.body = {
      result,
      variablesReference: 0,
    };

    this.sendResponse(response);
  }

  public shutdown() {
    this.client
      ?.shutdown()
      .then(() => {
        this.extout.appendLine("Debug Session has finished");
      })
      .catch((e) => {
        this.extout.appendLine(
          `Debug Session has finished with an error: ${e}`
        );
      });
  }

  // Id functions
  // ============

  /**
   * Map a given breakpoint string to a breakpoint ID.
   */
  private getBreakPointId(breakpointString: string): number {
    let id = this.bpIds[breakpointString];
    if (!id) {
      id = this.bpCount++;
      this.bpIds[breakpointString] = id;
    }
    return id;
  }

  /**
   * Map a given stackframe to a stackframe ID.
   */
  private getStackFrameId(
    threadId: string | number,
    frameString: string,
    frameIndex: number
  ): number {
    const storageString = `${threadId}###${frameString}###${frameIndex}`;
    let id = this.sfIds[storageString];
    if (!id) {
      id = this.sfCount++;
      this.sfIds[storageString] = id;
    }
    return id;
  }

  /**
   * Map a given variable scope to a variable scope ID.
   */
  private getVariableScopeId(frameId: number, scopeType: string): number {
    const storageString = `${frameId}###${scopeType}`;
    let id = this.vsIds[storageString];
    if (!id) {
      id = this.vsCount++;
      this.vsIds[storageString] = id;
    }
    return id;
  }

  /**
   * Map a given variable scope ID to a variable scope.
   */
  private getScopeLookupInfo(vsId: number): [number, string] {
    for (const [k, v] of Object.entries(this.vsIds)) {
      if (v === vsId) {
        const s = k.split("###");
        return [parseInt(s[0]), s[1]];
      }
    }

    return [-1, ""];
  }
}

class LogChannelAdapter {
  private out: vscode.OutputChannel;

  constructor(out: vscode.OutputChannel) {
    this.out = out;
  }

  log(value: string): void {
    this.out.appendLine(value);
  }

  error(value: string): void {
    this.out.appendLine(`Error: ${value}`);
    setTimeout(() => {
      this.out.show(true);
    }, 500);
  }
}

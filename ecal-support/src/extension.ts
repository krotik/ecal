import * as vscode from "vscode";
import { ProviderResult } from "vscode";
import { ECALDebugSession } from "./ecalDebugAdapter";

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.debug.registerDebugAdapterDescriptorFactory(
      "ecaldebug",
      new InlineDebugAdapterFactory()
    )
  );
}

export function deactivate() {}

class InlineDebugAdapterFactory
  implements vscode.DebugAdapterDescriptorFactory {
  createDebugAdapterDescriptor(
    _session: vscode.DebugSession
  ): ProviderResult<vscode.DebugAdapterDescriptor> {
    // Declare the ECALDebugSession as an DebugAdapterInlineImplementation so extension and adapter can
    // run in-process (making it possible to easily debug the adapter)
    return new vscode.DebugAdapterInlineImplementation(new ECALDebugSession());
  }
}

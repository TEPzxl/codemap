export type GraphMode = "function" | "package";

export function entrypointTargetMode(currentMode: GraphMode): GraphMode {
  return currentMode;
}

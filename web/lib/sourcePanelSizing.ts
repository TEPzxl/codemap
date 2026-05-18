export const sourcePanelDefaultHeight = 280;
export const sourcePanelMinHeight = 160;
export const sourcePanelMaxViewportRatio = 0.65;

export function clampSourcePanelHeight(height: number, viewportHeight: number): number {
  const maxHeight = Math.max(sourcePanelMinHeight, Math.floor(viewportHeight * sourcePanelMaxViewportRatio));
  return Math.min(Math.max(Math.round(height), sourcePanelMinHeight), maxHeight);
}

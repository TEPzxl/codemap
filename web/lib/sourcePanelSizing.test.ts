import test from "node:test";
import assert from "node:assert/strict";

import { clampSourcePanelHeight } from "./sourcePanelSizing";

test("clampSourcePanelHeight keeps source panel within usable bounds", () => {
  assert.equal(clampSourcePanelHeight(80, 900), 160);
  assert.equal(clampSourcePanelHeight(320, 900), 320);
  assert.equal(clampSourcePanelHeight(900, 900), 585);
});

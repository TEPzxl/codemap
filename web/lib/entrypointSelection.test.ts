import test from "node:test";
import assert from "node:assert/strict";

import { entrypointTargetMode } from "./entrypointSelection";

test("entrypoint selection preserves package graph mode", () => {
  assert.equal(entrypointTargetMode("package"), "package");
});

test("entrypoint selection keeps function graph mode", () => {
  assert.equal(entrypointTargetMode("function"), "function");
});

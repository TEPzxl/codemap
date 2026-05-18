import test from "node:test";
import assert from "node:assert/strict";

import { parseViewState, serializeViewState, viewStateFromGraphRequest } from "./viewState";

test("parseViewState restores entry, depth, direction, filters, and package from URL params", () => {
  assert.deepEqual(parseViewState("?entry=main.main&depth=5&direction=upstream&show_external=true&show_unresolved=true&show_interface=true&expand_interface=true&package=example/internal"), {
    entry: "main.main",
    depth: 5,
    direction: "upstream",
    packageFilter: "example/internal",
    showExternal: true,
    showUnresolved: true,
    showInterface: true,
    expandInterface: true,
  });
});

test("parseViewState ignores invalid depth and treats absent filters as false", () => {
  assert.deepEqual(parseViewState("?entry=main.main&depth=bad"), {
    entry: "main.main",
    depth: undefined,
    direction: undefined,
    packageFilter: undefined,
    showExternal: false,
    showUnresolved: false,
    showInterface: false,
    expandInterface: false,
  });
});

test("serializeViewState writes canonical query names and omits disabled filters", () => {
  const params = serializeViewState({
    entry: "main.main",
    depth: 4,
    direction: "both",
    showExternal: true,
    showUnresolved: false,
    showInterface: true,
    expandInterface: false,
    packagePrefix: "example/internal/service",
  });

  assert.equal(params.toString(), "entry=main.main&depth=4&direction=both&show_external=true&show_interface=true&package=example%2Finternal%2Fservice");
});

test("viewStateFromGraphRequest exposes package as packageFilter for UI state", () => {
  assert.deepEqual(
    viewStateFromGraphRequest({
      entry: "main.main",
      depth: 2,
      direction: "upstream",
      packagePrefix: "example/internal",
      showExternal: true,
    }),
    {
      entry: "main.main",
      depth: 2,
      direction: "upstream",
      packageFilter: "example/internal",
      showExternal: true,
      showUnresolved: false,
      showInterface: false,
      expandInterface: false,
    },
  );
});

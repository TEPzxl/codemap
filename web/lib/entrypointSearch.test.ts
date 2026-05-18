import test from "node:test";
import assert from "node:assert/strict";

import { searchEntrypoints } from "./entrypointSearch";
import type { Entrypoint } from "@/types/graph";

const entrypoints: Entrypoint[] = [
  {
    id: "example/cmd/server.main",
    label: "main",
    package: "example/cmd/server",
    file: "cmd/server/main.go",
    kind: "function",
    reasons: ["main-function"],
  },
  {
    id: "example/internal/app.Run",
    label: "Run",
    package: "example/internal/app",
    file: "internal/app/app.go",
    kind: "function",
    reasons: ["exported-function", "name:Run"],
  },
  {
    id: "example/internal/handler.(*UserHandler).CreateUser",
    label: "UserHandler.CreateUser",
    package: "example/internal/handler",
    file: "internal/handler/user.go",
    kind: "method",
    reasons: ["receiver:Handler"],
  },
  {
    id: "example/internal/jobs.(*Worker).Run",
    label: "Worker.Run",
    package: "example/internal/jobs",
    file: "internal/jobs/worker.go",
    kind: "method",
    reasons: ["contains-goroutine"],
  },
];

test("searchEntrypoints returns suggested entrypoints by heuristic group when query is empty", () => {
  assert.deepEqual(
    searchEntrypoints(entrypoints, "", 2).map((group) => ({
      label: group.label,
      ids: group.items.map((entrypoint) => entrypoint.id),
    })),
    [
      { label: "Main", ids: ["example/cmd/server.main"] },
      { label: "Startup", ids: ["example/internal/app.Run"] },
    ],
  );
});

test("searchEntrypoints matches label id package file and reasons", () => {
  assert.deepEqual(searchEntrypoints(entrypoints, "CreateUser").flatMap((group) => group.items.map((entrypoint) => entrypoint.id)), [
    "example/internal/handler.(*UserHandler).CreateUser",
  ]);
  assert.deepEqual(searchEntrypoints(entrypoints, "internal/jobs").flatMap((group) => group.items.map((entrypoint) => entrypoint.id)), [
    "example/internal/jobs.(*Worker).Run",
  ]);
  assert.deepEqual(searchEntrypoints(entrypoints, "worker.go").flatMap((group) => group.items.map((entrypoint) => entrypoint.id)), [
    "example/internal/jobs.(*Worker).Run",
  ]);
  assert.deepEqual(searchEntrypoints(entrypoints, "goroutine").flatMap((group) => group.items.map((entrypoint) => entrypoint.id)), [
    "example/internal/jobs.(*Worker).Run",
  ]);
});

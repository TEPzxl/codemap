import test from "node:test";
import assert from "node:assert/strict";

import { searchCurrentGraphNodes } from "./currentGraphSearch";
import type { Graph } from "@/types/graph";

const graph: Graph = {
  entry: "example.com/app.main",
  nodes: [
    {
      id: "example.com/app.main",
      label: "main",
      kind: "function",
      package: "example.com/app",
      file: "cmd/app/main.go",
      start_line: 1,
      end_line: 5,
      is_external: false,
    },
    {
      id: "example.com/app/internal/service.(*UserService).CreateUser",
      label: "UserService.CreateUser",
      kind: "method",
      package: "example.com/app/internal/service",
      receiver: "*UserService",
      file: "internal/service/user.go",
      start_line: 10,
      end_line: 20,
      is_external: false,
    },
    {
      id: "fmt.Println",
      label: "Println",
      kind: "external",
      package: "fmt",
      file: "",
      start_line: 0,
      end_line: 0,
      is_external: true,
    },
  ],
  edges: [],
  warnings: [],
};

test("searchCurrentGraphNodes matches label id package and file", () => {
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "userservice").map((node) => node.id),
    ["example.com/app/internal/service.(*UserService).CreateUser"],
  );
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "CreateUser").map((node) => node.id),
    ["example.com/app/internal/service.(*UserService).CreateUser"],
  );
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "internal/service").map((node) => node.id),
    ["example.com/app/internal/service.(*UserService).CreateUser"],
  );
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "cmd/app/main.go").map((node) => node.id),
    ["example.com/app.main"],
  );
});

test("searchCurrentGraphNodes returns bounded stable results", () => {
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "", 2).map((node) => node.id),
    [],
  );
  assert.deepEqual(
    searchCurrentGraphNodes(graph, "example.com", 1).map((node) => node.id),
    ["example.com/app.main"],
  );
});
